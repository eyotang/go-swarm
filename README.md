# go-swarm

A Swarm API client enabling Go programs to interact with Swarm in a simple and uniform way.

## NOTE

Release v0.1.0 (released on 11-06-2022)  contains some backwards
incompatible changes that were needed to partly support the V9 Swarm API.

## Coverage

This API client package covers most of the existing Swarm API calls and is updated regularly
to add new and/or missing endpoints. Currently, the following services are supported:

- [x] Projects

## Usage

```go
import "github.com/eyotang/go-swarm"
```

Construct a new Swarm client, then use the various services on the client to
access different parts of the Swarm API. For example, to list all
users:

```go
sw, err := swarm.NewBasicAuthClient("username", "password/ticket")
if err != nil {
  log.Fatalf("Failed to create client: %v", err)
}
projects, _, err := sw.Projects.ListProjects(&swarm.ListProjectsOptions{})
```

There are a few `With...` option functions that can be used to customize
the API client. For example, to set a custom base URL:

```go
sw, err := swarm.NewBasicAuthClient("username", "password/ticket", gitlab.WithBaseURL("https://swarm.url"))
if err != nil {
  log.Fatalf("Failed to create client: %v", err)
}
projects, _, err := sw.Projects.ListProjects(&swarm.ListProjectsOptions{})
```

Some API methods have optional parameters that can be passed. For example,
to list all projects for user "svanharmelen":

```go
sw := swarm.NewBasicAuthClient("username", "password/ticket")
opt := &ListProjectsOptions{Fields: "name,branches"}
projects, _, err := sw.Projects.ListProjects(opt)
```

### Examples

The [examples](https://github.com/xanzy/go-gitlab/tree/master/examples) directory
contains a couple for clear examples, of which one is partially listed here as well:

```go
package main

import (
	"log"

	"github.com/eyotang/go-swarm"
)

func main() {
    sw, err := swarm.NewBasicAuthClient("username", "password", swarm.WithBasicURL("http://swarm.url"))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create new project
	opt := &CreateProjectOptions{
        Name: swarm.String("got-dev"),
        Members: []*string{swarm.String("eyotang"), swarm.String("tangyongqiang")}
    }
	project, _, err := sw.Projects.CreateProject(p)
	if err != nil {
		log.Fatal(err)
	}
}
```

For complete usage of go-gitlab, see the full [package docs](https://godoc.org/github.com/xanzy/go-gitlab).

## ToDo

- The biggest thing this package still needs is tests :disappointed:

## Issues

- If you have an issue: report it on the [issue tracker](https://github.com/xanzy/go-gitlab/issues)

## Author

笨大神 (<bensetang@126.com>)

## License

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>
