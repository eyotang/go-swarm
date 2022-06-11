package swarm

import (
	"fmt"
	"net/http"
)

// ProjectsService handles communication with the project/group
// access requests related methods of the Swarm API.
//
// Swarm API docs: https://www.perforce.com/manuals/swarm/Content/Swarm/swarm-apidoc_endpoints.html
type ProjectsService struct {
	client *Client
}

type ListProjectsOptions struct {
	Fields   *string `query:"fields"`
	Workflow *string `query:"fields"`
}

// Project represents a project in swarm.
//
// Swarm API docs:
// https://www.perforce.com/manuals/swarm/Content/Swarm/swarm-apidoc_endpoints.html
type Project struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Members     []string `json:"members"`
	Branches    []Branch `json:"branches"`
}

type Branch struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Workflow string   `json:"workflow"`
	Paths    []string `json:"paths"`
	Defaults struct {
		Reviewers struct {
			Users map[string]interface{} `json:"users"`
		} `json:"reviewers"`
	} `json:"defaults"`
}

func (p Project) String() string {
	return Stringify(p)
}

// ListProjects gets a list of projects.
//
// Swarm API docs: https://www.perforce.com/manuals/swarm/Content/Swarm/swarm-apidoc_endpoint_projects.html#Get_List_of_Projects_..420
func (s *ProjectsService) ListProjects(opt *ListProjectsOptions, options ...RequestOptionFunc) ([]*Project, *Response, error) {
	u := "projects"

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var p *struct {
		Projects []*Project `json:"projects"`
	}
	resp, err := s.client.Do(req, &p)
	if err != nil {
		return nil, resp, err
	}

	return p.Projects, resp, err
}

func (s *ProjectsService) GetProject(pid interface{}, options ...RequestOptionFunc) (*Project, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	var p *struct {
		Project *Project `json:"project"`
	}
	resp, err := s.client.Do(req, &p)
	if err != nil {
		return nil, resp, err
	}

	return p.Project, resp, err
}

type CreateProjectOptions struct {
	Name     *string          `query:"name"`
	Members  []*string        `query:"members"`
	Branches []*BranchOptions `query:"branches"`
}

type BranchOptions struct {
	ID       *string          `query:"id"`
	Name     *string          `query:"name"`
	Workflow *string          `query:"workflow"`
	Paths    *string          `query:"paths"`
	Defaults *DefaultsOptions `query:"defaults"`
}

type DefaultsOptions struct {
	Reviewers map[string]*ReviewerOptions `query:"reviewers"`
}

type ReviewerOptions struct {
	Required *string `query:"required"`
}

func (s *ProjectsService) CreateProject(opt *CreateProjectOptions, options ...RequestOptionFunc) (*Project, *Response, error) {
	u := "projects"

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	p := new(struct {
		*Project `json:"project"`
	})
	resp, err := s.client.Do(req, p)
	if err != nil {
		return nil, resp, err
	}

	return p.Project, resp, err
}

func (s *ProjectsService) DeleteProject(pid interface{}, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

type UpdateProjectOptions struct {
	Name     *string          `query:"name"`
	Members  []*string        `query:"members"`
	Branches []*BranchOptions `query:"branches"`
}

func (s *ProjectsService) UpdateProject(pid interface{}, opt *UpdateProjectOptions, options ...RequestOptionFunc) (*Project, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPatch, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	p := new(struct {
		Project *Project `json:"project"`
	})
	resp, err := s.client.Do(req, p)
	if err != nil {
		return nil, resp, err
	}

	return p.Project, resp, err
}
