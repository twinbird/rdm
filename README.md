# rdm

Redmine command line client.

# Install

```sh
$ go get golang.org/x/text/encoding/japanese
$ go get github.com/twinbird/rdm
```

# Usage

Listing all project list.

```sh
$ rdm
```

Listing all issues in project list.

```sh
$ rdm -i [project id]
```

Output all project list to MS Excel Format.

```sh
$ rdm -E [filepath]
```

Output all issues to MS Excel Format in project.

```sh
$ rdm -i [project id] -E [filepath]
```
