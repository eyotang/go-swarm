package swarm

import (
	"fmt"
	"net/http"
)

type WorkflowService struct {
	client *Client
}

type ListWorkflowsOptions struct {
	Fields  *string `query:"fields"`
	NoCache *string `query:"noCache"`
}

type Workflow struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Shared      bool     `json:"shared"`
	Owners      []string `json:"owners"`
	// review rules
	OnSubmit       OnSubmit   `json:"on_submit"`
	EndRules       EndRule    `json:"end_rules"`
	AutoApprove    ReviewRule `json:"auto_approve"`
	CountedVotes   ReviewRule `json:"counted_votes"`
	GroupExclusion ReviewRule `json:"group_exclusions"`
	UserExclusion  ReviewRule `json:"user_exclusions"`
}

type ReviewRule struct {
	Rule interface{} `json:"rule"`
	Mode string      `json:"mode"`
}

type OnSubmit struct {
	WithReview    ReviewRule `json:"with_review"`
	WithoutReview ReviewRule `json:"without_review"`
}

type EndRule struct {
	Update ReviewRule `json:"update"`
}

func (w Workflow) String() string {
	return Stringify(w)
}

// ListWorkflows gets a list of workflows.
//
// Swarm API docs: https://www.perforce.com/manuals/swarm/Content/Swarm/swarm-apidoc_endpoint_projects.html#Get_List_of_Projects_..420
func (s *WorkflowService) ListWorkflows(opt *ListWorkflowsOptions, options ...RequestOptionFunc) ([]*Workflow, *Response, error) {
	u := "workflows"

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var r *struct {
		Workflows []*Workflow `json:"workflows"`
	}
	resp, err := s.client.Do(req, &r)
	if err != nil {
		return nil, resp, err
	}

	return r.Workflows, resp, err
}

func (s *WorkflowService) GetWorkflow(pid interface{}, options ...RequestOptionFunc) (*Workflow, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("workflows/%s", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	var r *struct {
		Workflow *Workflow `json:"workflow"`
	}
	resp, err := s.client.Do(req, &r)
	if err != nil {
		return nil, resp, err
	}

	return r.Workflow, resp, err
}
