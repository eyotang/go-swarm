package swarm

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWorkflowService_ListWorkflows(t *testing.T) {
	Convey("test WorkflowService_ListWorkflows", t, func() {
		mux, server, client := setup(t)
		defer teardown(server)

		mux.HandleFunc("/api/v9/workflows", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			fmt.Fprint(w, `{
			  "workflows": [
				{
				  "on_submit": {
					"with_review": {
					  "rule": "approved",
					  "mode": "inherit"
					},
					"without_review": {
					  "rule": "auto_create",
					  "mode": "inherit"
					}
				  },
				  "name": "test",
				  "description": "",
				  "shared": false,
				  "owners": [
					"root"
				  ],
				  "end_rules": {
					"update": {
					  "rule": "no_revision",
					  "mode": "inherit"
					}
				  },
				  "auto_approve": {
					"rule": "never",
					"mode": "inherit"
				  },
				  "counted_votes": {
					"rule": "members",
					"mode": "inherit"
				  },
				  "group_exclusions": {
					"rule": [],
					"mode": "inherit"
				  },
				  "user_exclusions": {
					"rule": [],
					"mode": "inherit"
				  },
				  "id": 3,
				  "user_restrictions": []
				},
				{
				  "on_submit": {
					"with_review": {
					  "rule": "approved",
					  "mode": "inherit"
					},
					"without_review": {
					  "rule": "auto_create",
					  "mode": "inherit"
					}
				  },
				  "name": "Most Restrictive Workflow",
				  "description": "",
				  "shared": false,
				  "owners": [
					"root"
				  ],
				  "end_rules": {
					"update": {
					  "rule": "no_revision",
					  "mode": "inherit"
					}
				  },
				  "auto_approve": {
					"rule": "never",
					"mode": "inherit"
				  },
				  "counted_votes": {
					"rule": "anyone",
					"mode": "inherit"
				  },
				  "group_exclusions": {
					"rule": [],
					"mode": "inherit"
				  },
				  "user_exclusions": {
					"rule": [],
					"mode": "inherit"
				  },
				  "id": 1,
				  "user_restrictions": null
				}
			  ]
			}`)
		})

		opt := &ListWorkflowsOptions{}

		workflows, _, err := client.Workflows.ListWorkflows(opt)
		So(err, ShouldBeNil)

		want := []*Workflow{
			{
				ID:          3,
				Name:        "test",
				Description: "",
				Shared:      false,
				Owners:      []string{"root"},
				OnSubmit: OnSubmit{
					WithReview:    ReviewRule{Rule: "approved", Mode: "inherit"},
					WithoutReview: ReviewRule{Rule: "auto_create", Mode: "inherit"},
				},
				EndRules: EndRule{
					Update: ReviewRule{Rule: "no_revision", Mode: "inherit"},
				},
				AutoApprove:    ReviewRule{Rule: "never", Mode: "inherit"},
				CountedVotes:   ReviewRule{Rule: "members", Mode: "inherit"},
				GroupExclusion: ReviewRule{Rule: []interface{}{}, Mode: "inherit"},
				UserExclusion:  ReviewRule{Rule: []interface{}{}, Mode: "inherit"},
			},
			{
				ID:          1,
				Name:        "Most Restrictive Workflow",
				Description: "",
				Shared:      false,
				Owners:      []string{"root"},
				OnSubmit: OnSubmit{
					WithReview:    ReviewRule{Rule: "approved", Mode: "inherit"},
					WithoutReview: ReviewRule{Rule: "auto_create", Mode: "inherit"},
				},
				EndRules: EndRule{
					Update: ReviewRule{Rule: "no_revision", Mode: "inherit"},
				},
				AutoApprove:    ReviewRule{Rule: "never", Mode: "inherit"},
				CountedVotes:   ReviewRule{Rule: "anyone", Mode: "inherit"},
				GroupExclusion: ReviewRule{Rule: []interface{}{}, Mode: "inherit"},
				UserExclusion:  ReviewRule{Rule: []interface{}{}, Mode: "inherit"},
			},
		}

		So(workflows, ShouldResemble, want)
	})
}

func TestWorkflowsService_GetWorkflow(t *testing.T) {
	Convey("test WorkflowsService_GetWorkflow", t, func() {
		mux, server, client := setup(t)
		defer teardown(server)

		mux.HandleFunc("/api/v9/workflows/3", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			fmt.Fprint(w, `{
			  "workflow": {
				"on_submit": {
				  "with_review": {
					"rule": "approved",
					"mode": "inherit"
				  },
				  "without_review": {
					"rule": "auto_create",
					"mode": "inherit"
				  }
				},
				"name": "test",
				"description": "",
				"shared": false,
				"owners": [
				  "root"
				],
				"end_rules": {
				  "update": {
					"rule": "no_revision",
					"mode": "inherit"
				  }
				},
				"auto_approve": {
				  "rule": "never",
				  "mode": "inherit"
				},
				"counted_votes": {
				  "rule": "members",
				  "mode": "inherit"
				},
				"group_exclusions": {
				  "rule": [],
				  "mode": "inherit"
				},
				"user_exclusions": {
				  "rule": [],
				  "mode": "inherit"
				},
				"id": 3,
				"user_restrictions": []
			  }
			}`)
		})

		workflow, _, err := client.Workflows.GetWorkflow(3)
		So(err, ShouldBeNil)

		want := &Workflow{
			ID:          3,
			Name:        "test",
			Description: "",
			Shared:      false,
			Owners:      []string{"root"},
			OnSubmit: OnSubmit{
				WithReview:    ReviewRule{Rule: "approved", Mode: "inherit"},
				WithoutReview: ReviewRule{Rule: "auto_create", Mode: "inherit"},
			},
			EndRules: EndRule{
				Update: ReviewRule{Rule: "no_revision", Mode: "inherit"},
			},
			AutoApprove:    ReviewRule{Rule: "never", Mode: "inherit"},
			CountedVotes:   ReviewRule{Rule: "members", Mode: "inherit"},
			GroupExclusion: ReviewRule{Rule: []interface{}{}, Mode: "inherit"},
			UserExclusion:  ReviewRule{Rule: []interface{}{}, Mode: "inherit"},
		}

		So(workflow, ShouldResemble, want)
	})
}

func TestWorkflowsService_GetWorkflow3(t *testing.T) {
	Convey("test WorkflowsService_GetWorkflow3", t, func() {
		client, err := NewBasicAuthClient("root", "531E89C85298D7C86349158A64944AD8", WithBaseURL("http://10.154.0.59"))
		So(err, ShouldBeNil)

		workflow, _, err := client.Workflows.GetWorkflow(3)
		So(err, ShouldBeNil)

		want := &Workflow{
			ID:          3,
			Name:        "test",
			Description: "",
			Shared:      false,
			Owners:      []string{"root"},
			OnSubmit: OnSubmit{
				WithReview:    ReviewRule{Rule: "approved", Mode: "inherit"},
				WithoutReview: ReviewRule{Rule: "auto_create", Mode: "inherit"},
			},
			EndRules: EndRule{
				Update: ReviewRule{Rule: "no_revision", Mode: "inherit"},
			},
			AutoApprove:    ReviewRule{Rule: "never", Mode: "inherit"},
			CountedVotes:   ReviewRule{Rule: "members", Mode: "inherit"},
			GroupExclusion: ReviewRule{Rule: []interface{}{}, Mode: "inherit"},
			UserExclusion:  ReviewRule{Rule: []interface{}{}, Mode: "inherit"},
		}

		So(workflow, ShouldResemble, want)
	})
}

func TestWorkflowsService_SetGlobalWorkflow(t *testing.T) {
	Convey("test WorkflowsService_SetGlobalWorkflow", t, func() {
		username := os.Getenv("USERNAME")
		password := os.Getenv("PASSWORD")
		url := os.Getenv("URL")

		// client is the Gitlab client being tested.
		client, err := NewBasicAuthClient(username, password, WithBaseURL(url))

		group := "swarm-group-Admin"
		user := "tangyongqiang"
		user2 := "swarm"

		// 获取Global Workflow
		workflow, _, err := client.Workflows.GetWorkflow(0)
		So(err, ShouldBeNil)

		want := &Workflow{
			ID:          0,
			Name:        "Global Workflow",
			Description: "Updated by v10 api.",
			Shared:      false,
			Owners:      []string{"root", "swarm", "tangyongqiang"},
			OnSubmit: OnSubmit{
				WithReview:    ReviewRule{Rule: "no_checking", Mode: "default"},
				WithoutReview: ReviewRule{Rule: "no_checking", Mode: "default"},
			},
			EndRules: EndRule{
				Update: ReviewRule{Rule: "no_checking", Mode: "default"},
			},
			AutoApprove:    ReviewRule{Rule: "never", Mode: "default"},
			CountedVotes:   ReviewRule{Rule: "anyone", Mode: "default"},
			GroupExclusion: ReviewRule{Rule: []interface{}{group}, Mode: "policy"},
			UserExclusion:  ReviewRule{Rule: []interface{}{user}, Mode: "policy"},
		}

		So(workflow, ShouldResemble, want)

		// 设置Global Workflow

		err = client.Workflows.SetGlobalExclusions([]string{group}, []string{user, user2})
		So(err, ShouldBeNil)
	})
}
