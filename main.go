// Redmine command line utility
package main

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
)

var setting Setting

const (
	ProjectListPath = "/projects.xml"
	IssueListPath   = "/issues.xml"
	SettingFilename = ".rdm"
	APIKeyQueryKey  = "?key="
)

type Projects struct {
	Project    []Project `xml:"project"`
	TotalCount int       `xml:"total_count,attr"`
}

type Project struct {
	Id          int    `xml:"id"`
	Name        string `xml:"name"`
	Identifier  string `xml:"identifier"`
	Description string `xml:"description"`
	CreatedOn   string `xml:"created_on"`
}

func projectToStdout() {
	projects := &Projects{}
	projects = projectList(projects, 0)

	fmt.Println("ID  : Name        : Description")
	for _, p := range projects.Project {
		fmt.Printf("%d:%s:%s\n", p.Id, p.Name, p.Description)
	}
}

// Save all project list to xlsx file
func projectToXlsx(outPath string) {
	projects := &Projects{}
	projects = projectList(projects, 0)

	xlfile := xlsx.NewFile()
	sheet, err := xlfile.AddSheet("Project List")
	if err != nil {
		panic(err)
	}

	// Header
	header := sheet.AddRow()
	header.AddCell().Value = "No"
	header.AddCell().Value = "Project ID"
	header.AddCell().Value = "Project Name"
	header.AddCell().Value = "Description"

	for i, p := range projects.Project {
		row := sheet.AddRow()

		// No
		row.AddCell().Value = strconv.Itoa(i)

		// Project Info
		row.AddCell().Value = strconv.Itoa(p.Id)
		row.AddCell().Value = p.Name
		row.AddCell().Value = p.Description
	}

	if err := xlfile.Save(outPath); err != nil {
		panic(err)
	}
}

// Get recursively all project
func projectList(mergeTgt *Projects, offset int) *Projects {
	const limit = 100
	url := setting.HostURL + ProjectListPath + APIKeyQueryKey + setting.APIKey + "&offset=" + strconv.Itoa(offset) + "&limit=" + strconv.Itoa(limit)
	res, _ := http.Get(url)
	defer res.Body.Close()

	xmldoc, _ := ioutil.ReadAll(res.Body)
	p := &Projects{}
	if err := xml.Unmarshal(xmldoc, p); err != nil {
		panic(err)
	}
	mergeTgt.Project = append(mergeTgt.Project, p.Project...)
	if p.TotalCount > offset+limit {
		mergeTgt = projectList(mergeTgt, offset+limit)
	}
	return mergeTgt
}

type Setting struct {
	HostURL string
	APIKey  string
}

func (setting *Setting) Load() error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	fn := filepath.Join(usr.HomeDir, SettingFilename)
	f, err := os.Open(fn)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(f)
	// HostURL
	scanner.Scan()
	setting.HostURL = scanner.Text()
	// APIKey
	scanner.Scan()
	setting.APIKey = scanner.Text()

	return nil
}

func (setting *Setting) Save() error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	fn := filepath.Join(usr.HomeDir, SettingFilename)
	saveStr := setting.HostURL + "\r\n" + setting.APIKey
	if err := ioutil.WriteFile(fn, []byte(saveStr), os.ModePerm); err != nil {
		return err
	}
	return nil
}

func (s *Setting) Dialog() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("Please input Redmine Host URL:")
	scanner.Scan()
	s.HostURL = scanner.Text()
	fmt.Printf("Please input Redmine API Key:")
	scanner.Scan()
	s.APIKey = scanner.Text()
}

type Issues struct {
	Issues     []Issue `xml:"issue"`
	TotalCount int     `xml:"total_count,attr"`
}

type TrackerTag struct {
	Name string `xml:"name,attr"`
}

type StatusTag struct {
	Name string `xml:"name,attr"`
}

type PriorityTag struct {
	Name string `xml:"name,attr"`
}

type AuthorTag struct {
	Name string `xml:"name,attr"`
}

type Assigned struct {
	Name string `xml:"name,attr"`
}

type Issue struct {
	Tracker     TrackerTag  `xml:"tracker"`
	Status      StatusTag   `xml:"status"`
	Priority    PriorityTag `xml:"priority"`
	Author      AuthorTag   `xml:"author"`
	Assigned    Assigned    `xml:"assigned_to"`
	Subject     string      `xml:"subject"`
	Description string      `xml:"description"`
	StartDate   string      `xml:"start_date"`
	DueDate     string      `xml:"due_date"`
}

// Get recursively all project
func issuesList(pid int, closed bool, mergeTgt *Issues, offset int) *Issues {
	const limit = 100
	url := setting.HostURL + IssueListPath + APIKeyQueryKey + setting.APIKey + "&offset=" + strconv.Itoa(offset) + "&limit=" + strconv.Itoa(limit) + "&project_id=" + strconv.Itoa(pid)
	if closed {
		url += "&status_id=closed"
	}
	res, _ := http.Get(url)
	defer res.Body.Close()

	xmldoc, _ := ioutil.ReadAll(res.Body)
	issues := &Issues{}
	if err := xml.Unmarshal(xmldoc, issues); err != nil {
		panic(err)
	}
	mergeTgt.Issues = append(mergeTgt.Issues, issues.Issues...)
	if issues.TotalCount > offset+limit {
		mergeTgt = issuesList(pid, closed, mergeTgt, offset+limit)
	}
	return mergeTgt
}

func issuesListStdout(pid int) {
	issues := &Issues{}
	issues = issuesList(pid, false, issues, 0)
	issues = issuesList(pid, true, issues, 0)

	fmt.Println("Tracker : Status : Priority : Author : Assigned : Subject : Description : StartDate : DueDate")
	for _, is := range issues.Issues {
		fmt.Printf("%s:%s:%s:%s:%s:%s:%s:%s:%s\n",
			is.Tracker.Name, is.Status.Name, is.Priority.Name, is.Author.Name, is.Assigned.Name, is.Subject, is.Description, is.StartDate, is.DueDate)
	}
}

// Save all issues list to xlsx file
func issuesToXlsx(pid int, outPath string) {
	issues := &Issues{}
	issues = issuesList(pid, false, issues, 0)
	issues = issuesList(pid, true, issues, 0)

	xlfile := xlsx.NewFile()
	sheet, err := xlfile.AddSheet("Issues List")
	if err != nil {
		panic(err)
	}

	// Header
	header := sheet.AddRow()
	header.AddCell().Value = "Traker"
	header.AddCell().Value = "Status"
	header.AddCell().Value = "Priority"
	header.AddCell().Value = "Author"
	header.AddCell().Value = "Assigned"
	header.AddCell().Value = "Subject"
	header.AddCell().Value = "Description"
	header.AddCell().Value = "StartDate"
	header.AddCell().Value = "DueDate"

	// Body
	for _, is := range issues.Issues {
		row := sheet.AddRow()

		row.AddCell().Value = is.Tracker.Name
		row.AddCell().Value = is.Status.Name
		row.AddCell().Value = is.Priority.Name
		row.AddCell().Value = is.Author.Name
		row.AddCell().Value = is.Assigned.Name
		row.AddCell().Value = is.Subject
		row.AddCell().Value = is.Description
		row.AddCell().Value = is.StartDate
		row.AddCell().Value = is.DueDate
	}

	if err := xlfile.Save(outPath); err != nil {
		panic(err)
	}
}

func main() {
	if err := setting.Load(); err != nil {
		setting.Dialog()
		if err = setting.Save(); err != nil {
			log.Fatal(err)
		}
	}

	doExcelOut := flag.Bool("E", false, "Output in Excel format.")
	fileOutPath := flag.String("f", "redmine.xlsx", "Excel file output path.(use with 'E' option)")
	issueOutprojectId := flag.Int("i", -1, "Output Issues list of the specified project.")

	flag.Parse()

	if *doExcelOut {
		if *issueOutprojectId >= 0 {
			issuesToXlsx(*issueOutprojectId, *fileOutPath)
		} else {
			projectToXlsx(*fileOutPath)
		}
	} else {
		if *issueOutprojectId >= 0 {
			issuesListStdout(*issueOutprojectId)
		} else {
			projectToStdout()
		}
	}
}
