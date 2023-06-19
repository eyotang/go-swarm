package swarm

import (
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
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
	ID          uint     `json:"id" query:"-"`
	Name        string   `json:"name" query:"name"`
	Description string   `json:"description" query:"description"`
	Shared      bool     `json:"shared" query:"shared"`
	Owners      []string `json:"owners" query:"owners"`
	// review rules
	OnSubmit       OnSubmit   `json:"on_submit" query:"on_submit"`
	EndRules       EndRule    `json:"end_rules" query:"end_rules"`
	AutoApprove    ReviewRule `json:"auto_approve" query:"auto_approve"`
	CountedVotes   ReviewRule `json:"counted_votes" query:"counted_votes"`
	GroupExclusion ReviewRule `json:"group_exclusions" query:"group_exclusions"`
	UserExclusion  ReviewRule `json:"user_exclusions" query:"user_exclusions"`
}

type ReviewRule struct {
	Rule interface{} `json:"rule" query:"rule"`
	Mode string      `json:"mode" query:"mode"`
}

type OnSubmit struct {
	WithReview    ReviewRule `json:"with_review" query:"with_review"`
	WithoutReview ReviewRule `json:"without_review" query:"without_review"`
}

type EndRule struct {
	Update ReviewRule `json:"update" query:"update"`
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
	flowId, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("workflows/%s", PathEscape(flowId))

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

func (s *WorkflowService) SetGlobalExclusions(groups []string, users []string) (err error) {
	var workflow *Workflow
	if workflow, _, err = s.GetWorkflow(0); err != nil {
		return
	}

	swarmGroups := make([]string, 0)
	for _, group := range groups {
		swarmGroups = append(swarmGroups, addPrefix(group, "swarm-group-"))
	}
	workflow.GroupExclusion.Rule = swarmGroups
	workflow.UserExclusion.Rule = users

	if err = s.updateWorkflow(0, workflow); err != nil {
		return
	}

	return
}

func (s *WorkflowService) updateWorkflow(pid interface{}, workflow *Workflow) (err error) {
	var (
		flowId string
		req    *retryablehttp.Request
	)
	flowId, err = parseID(pid)
	if err != nil {
		return
	}
	u := fmt.Sprintf(apiV10Path+"workflows/%s", PathEscape(flowId))
	if len(workflow.Description) <= 0 {
		workflow.Description = "Updated by v10 api."
	}

	if req, err = s.client.NewRequest(http.MethodPut, u, workflow, nil); err != nil {
		return
	}
	var r *struct {
		Data *struct {
			Workflows []*Workflow `json:"workflows"`
		} `json:"data"`
	}
	if _, err = s.client.Do(req, &r); err != nil {
		return
	}
	return
}
