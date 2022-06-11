package swarm

import (
	"fmt"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestProjectsService_ListProjects(t *testing.T) {
	Convey("test ProjectsService_ListProjects", t, func() {
		mux, server, client := setup(t)
		defer teardown(server)

		mux.HandleFunc("/api/v9/projects", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			fmt.Fprint(w, `{
			  "projects": [
				{
				  "id": "xxx-mainline",
				  "branches": [
					{
					  "id": "client",
					  "name": "Client",
					  "workflow": "6",
					  "paths": [],
					  "defaults": {
						"reviewers": {
						  "users": {
							"eyotang": [],
							"tangyongqiang": []
						  }
						}
					  },
					  "minimumUpVotes": null,
					  "retainDefaultReviewers": false,
					  "moderators": [],
					  "moderators-groups": []
					}
				  ],
				  "members": [
					"eyotang",
					"tangyq"
				  ],
				  "name": "Exxx_Mainline"
				},
				{
				  "id": "main",
				  "branches": [
					{
					  "id": "artdev",
					  "name": "ArtDev",
					  "workflow": "6",
					  "paths": [],
					  "defaults": {
						"reviewers": {
						  "users": {
								"eyotang": {
									"required": true
								},
								"tangyongqiang": []
                        	}
						}
					  },
					  "minimumUpVotes": null,
					  "retainDefaultReviewers": false,
					  "moderators": [],
					  "moderators-groups": []
					},
					{
					  "id": "hhq",
					  "name": "HHQ",
					  "workflow": "5",
					  "paths": [],
					  "defaults": {
						"reviewers": {
						  "users": {
							"tangyongqiang": []
						  }
						}
					  },
					  "minimumUpVotes": null,
					  "retainDefaultReviewers": false,
					  "moderators": [
						"eyotang",
						"tangyq"
					  ],
					  "moderators-groups": []
					}
				  ],
				  "deleted": false,
				  "description": "",
				  "members": [
					"eyotang",
					"tangyongqiang",
					"swarm"
				  ],
				  "minimumUpVotes": null,
				  "name": "DMXX.YYY",
				  "owners": [
					"root"
				  ]
				}
			  ]
			}`)
		})

		opt := &ListProjectsOptions{}

		projects, _, err := client.Projects.ListProjects(opt)
		So(err, ShouldBeNil)

		want := []*Project{
			{
				ID:          "xxx-mainline",
				Name:        "Exxx_Mainline",
				Description: "",
				Members:     []string{"eyotang", "tangyq"},
				Branches: []Branch{
					{ID: "client", Name: "Client", Workflow: "6", Paths: []string{}},
				},
			},
			{
				ID:      "main",
				Name:    "DMXX.YYY",
				Members: []string{"eyotang", "tangyongqiang", "swarm"},
				Branches: []Branch{
					{ID: "artdev", Name: "ArtDev", Workflow: "6", Paths: []string{}},
					{ID: "hhq", Name: "HHQ", Workflow: "5", Paths: []string{}},
				},
			},
		}
		users := make(map[string]interface{})
		users["eyotang"] = []interface{}{}
		users["tangyongqiang"] = []interface{}{}
		want[0].Branches[0].Defaults.Reviewers.Users = users

		users = make(map[string]interface{})
		users["eyotang"] = map[string]interface{}{"required": true}
		users["tangyongqiang"] = []interface{}{}
		want[1].Branches[0].Defaults.Reviewers.Users = users

		users = make(map[string]interface{})
		users["tangyongqiang"] = []interface{}{}
		want[1].Branches[1].Defaults.Reviewers.Users = users

		So(projects, ShouldResemble, want)
	})
}

func TestProjectsService_GetProject(t *testing.T) {
	Convey("test ProjectsService_GetProject", t, func() {
		mux, server, client := setup(t)
		defer teardown(server)

		mux.HandleFunc("/api/v9/projects/xxx-mainline", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			fmt.Fprint(w, `{
			"project": {
				"id": "xxx-mainline",
				"branches": [
					{
						"id": "client",
						"name": "Client",
						"workflow": "6",
						"paths": [
							"//xxx.Mainline/abvc_ArtDev/Assets/...",
							"//xxx.Mainline/abvc_ArtDev/Assets/Scripts/..."
						],
						"defaults": {
							"reviewers": {
								"users": {
									"eyotang": {
										"required": true
									},
									"tangyongqiang": []
								}
							}
						},
						"minimumUpVotes": null,
						"retainDefaultReviewers": false,
						"moderators": [],
						"moderators-groups": []
					}
				],
				"jobview": "",
				"members": [
					"eyotang",
					"swarm"
				],
				"name": "Exxx_Mainline"
			}
			}`)
		})

		projects, _, err := client.Projects.GetProject("xxx-mainline")
		So(err, ShouldBeNil)

		want := &Project{
			ID:          "xxx-mainline",
			Name:        "Exxx_Mainline",
			Description: "",
			Members:     []string{"eyotang", "swarm"},
			Branches: []Branch{
				{
					ID:       "client",
					Name:     "Client",
					Workflow: "6",
					Paths:    []string{"//xxx.Mainline/abvc_ArtDev/Assets/...", "//xxx.Mainline/abvc_ArtDev/Assets/Scripts/..."},
				},
			},
		}

		users := make(map[string]interface{})
		users["eyotang"] = map[string]interface{}{"required": true}
		users["tangyongqiang"] = []interface{}{}
		want.Branches[0].Defaults.Reviewers.Users = users

		So(projects, ShouldResemble, want)
	})
}

func TestProjectsService_CreateProject(t *testing.T) {
	Convey("test ProjectsService_CreateProject", t, func() {
		mux, server, client := setup(t)
		defer teardown(server)

		mux.HandleFunc("/api/v9/projects", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPost)
			fmt.Fprint(w, `{
				"project": {
					"id": "got-dev",
					"branches": [],
					"defaults": {
						"reviewers": []
					},
					"deleted": false,
					"deploy": {
						"enabled": false,
						"url": null
					},
					"description": null,
					"emailFlags": [],
					"jobview": null,
					"members": [
						"eyotang",
						"tangyongqiang"
					],
					"minimumUpVotes": null,
					"name": "Got-dev",
					"owners": [],
					"private": false,
					"retainDefaultReviewers": false,
					"subgroups": [],
					"tests": {
						"enabled": false,
						"url": null
					},
					"workflow": null
				},
				"readme": "",
				"mode": "add"
			}`)
		})

		opt := &CreateProjectOptions{
			Name:    String("got-dev"),
			Members: []*string{String("eyotang"), String("tangyongqiang")},
		}
		projects, _, err := client.Projects.CreateProject(opt)
		So(err, ShouldBeNil)

		want := &Project{
			ID:          "got-dev",
			Name:        "Got-dev",
			Description: "",
			Members:     []string{"eyotang", "tangyongqiang"},
			Branches:    []Branch{},
		}

		So(projects, ShouldResemble, want)
	})
}

func TestProjectsService_CreateProjectWithBranch(t *testing.T) {
	Convey("test ProjectsService_CreateProjectWithBranch", t, func() {
		mux, server, client := setup(t)
		defer teardown(server)

		mux.HandleFunc("/api/v9/projects", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPost)
			fmt.Fprint(w, `{
				"project": {
					"id": "got-dev",
					"branches": [
						{
						"id": "client",
						"name": "Client",
						"workflow": "6",
						"paths": [
							"//xxx.Mainline/abvc_ArtDev/Assets/...",
							"//xxx.Mainline/abvc_ArtDev/Assets/Scripts/..."
						],
						"defaults": {
							"reviewers": {
								"users": {
									"eyotang": {
										"required": true
									},
									"tangyongqiang": []
								}
							}
						},
						"minimumUpVotes": null,
						"retainDefaultReviewers": false,
						"moderators": [],
						"moderators-groups": []
					}
					],
					"defaults": {
						"reviewers": []
					},
					"deleted": false,
					"deploy": {
						"enabled": false,
						"url": null
					},
					"description": null,
					"emailFlags": [],
					"jobview": null,
					"members": [
						"eyotang",
						"tangyongqiang"
					],
					"minimumUpVotes": null,
					"name": "Got-dev",
					"owners": [],
					"private": false,
					"retainDefaultReviewers": false,
					"subgroups": [],
					"tests": {
						"enabled": false,
						"url": null
					},
					"workflow": null
				},
				"readme": "",
				"mode": "add"
			}`)
		})

		opt := &CreateProjectOptions{
			Name:    String("got-dev"),
			Members: []*string{String("eyotang"), String("tangyongqiang")},
			Branches: []*BranchOptions{
				{
					Name:     String("Client"),
					Workflow: String("6"),
					Paths:    String("//xxx.Mainline/abvc_ArtDev/Assets/...\n//xxx.Mainline/abvc_ArtDev/Assets/Scripts/..."),
					Defaults: new(DefaultsOptions),
				},
			},
		}
		reviewers := make(map[string]*ReviewerOptions)
		reviewers["eyotang"] = &ReviewerOptions{Required: String("true")}
		reviewers["tangyongqiang"] = &ReviewerOptions{Required: String("false")}
		opt.Branches[0].Defaults.Reviewers = reviewers
		projects, _, err := client.Projects.CreateProject(opt)
		So(err, ShouldBeNil)

		want := &Project{
			ID:          "got-dev",
			Name:        "Got-dev",
			Description: "",
			Members:     []string{"eyotang", "tangyongqiang"},
			Branches: []Branch{
				{
					ID:       "client",
					Name:     "Client",
					Workflow: "6",
					Paths:    []string{"//xxx.Mainline/abvc_ArtDev/Assets/...", "//xxx.Mainline/abvc_ArtDev/Assets/Scripts/..."},
				},
			},
		}
		users := make(map[string]interface{})
		users["eyotang"] = map[string]interface{}{"required": true}
		users["tangyongqiang"] = []interface{}{}
		want.Branches[0].Defaults.Reviewers.Users = users

		So(projects, ShouldResemble, want)
	})
}

func TestProjectsService_UpdateProject(t *testing.T) {
	Convey("test ProjectsService_UpdateProject", t, func() {
		mux, server, client := setup(t)
		defer teardown(server)

		mux.HandleFunc("/api/v9/projects/got-dev", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPatch)
			fmt.Fprint(w, `{
				"project": {
					"id": "got-dev",
					"branches": [
						{
						"id": "client",
						"name": "Client",
						"workflow": "6",
						"paths": [
							"//xxx.Mainline/abvc_ArtDev/Assets/...",
							"//xxx.Mainline/abvc_ArtDev/Assets/Scripts/..."
						],
						"defaults": {
							"reviewers": {
								"users": {
									"eyotang": {
										"required": true
									},
									"tangyongqiang": []
								}
							}
						},
						"minimumUpVotes": null,
						"retainDefaultReviewers": false,
						"moderators": [],
						"moderators-groups": []
					}
					],
					"defaults": {
						"reviewers": []
					},
					"deleted": false,
					"deploy": {
						"enabled": false,
						"url": null
					},
					"description": null,
					"emailFlags": [],
					"jobview": null,
					"members": [
						"eyotang",
						"tangyongqiang"
					],
					"minimumUpVotes": null,
					"name": "Got-dev",
					"owners": [],
					"private": false,
					"retainDefaultReviewers": false,
					"subgroups": [],
					"tests": {
						"enabled": false,
						"url": null
					},
					"workflow": null
				},
				"readme": "",
				"mode": "add"
			}`)
		})

		opt := &UpdateProjectOptions{
			Name:    String("got-dev"),
			Members: []*string{String("eyotang"), String("tangyongqiang")},
			Branches: []*BranchOptions{
				{
					Name:     String("Client"),
					Workflow: String("6"),
					Paths:    String("//xxx.Mainline/abvc_ArtDev/Assets/...\n//xxx.Mainline/abvc_ArtDev/Assets/Scripts/..."),
					Defaults: new(DefaultsOptions),
				},
			},
		}
		reviewers := make(map[string]*ReviewerOptions)
		reviewers["eyotang"] = &ReviewerOptions{Required: String("true")}
		reviewers["tangyongqiang"] = &ReviewerOptions{Required: String("false")}
		opt.Branches[0].Defaults.Reviewers = reviewers
		projects, _, err := client.Projects.UpdateProject("got-dev", opt)
		So(err, ShouldBeNil)

		want := &Project{
			ID:          "got-dev",
			Name:        "Got-dev",
			Description: "",
			Members:     []string{"eyotang", "tangyongqiang"},
			Branches: []Branch{
				{
					ID:       "client",
					Name:     "Client",
					Workflow: "6",
					Paths:    []string{"//xxx.Mainline/abvc_ArtDev/Assets/...", "//xxx.Mainline/abvc_ArtDev/Assets/Scripts/..."},
				},
			},
		}
		users := make(map[string]interface{})
		users["eyotang"] = map[string]interface{}{"required": true}
		users["tangyongqiang"] = []interface{}{}
		want.Branches[0].Defaults.Reviewers.Users = users

		So(projects, ShouldResemble, want)
	})
}

func TestProjectsService_UpdateProjectWithUnknownUser(t *testing.T) {
	Convey("test ProjectsService_UpdateProjectWithUnknownUser", t, func() {
		mux, server, client := setup(t)
		defer teardown(server)

		mux.HandleFunc("/api/v9/projects/got-dev", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPatch)
			err := `{
				"error": "Bad Request",
				"isValid": false,
				"details": {
					"branches": "Unknown user id(s): tangyongqiang"
				}
			}`
			http.Error(w, err, http.StatusBadRequest)
		})

		opt := &UpdateProjectOptions{
			Name:    String("got-dev"),
			Members: []*string{String("eyotang"), String("tangyongqiang")},
			Branches: []*BranchOptions{
				{
					Name:     String("Client"),
					Workflow: String("6"),
					Paths:    String("//xxx.Mainline/abvc_ArtDev/Assets/...\n//xxx.Mainline/abvc_ArtDev/Assets/Scripts/..."),
					Defaults: new(DefaultsOptions),
				},
			},
		}
		reviewers := make(map[string]*ReviewerOptions)
		reviewers["eyotang"] = &ReviewerOptions{Required: String("true")}
		reviewers["tangyongqiang"] = &ReviewerOptions{Required: String("false")}
		opt.Branches[0].Defaults.Reviewers = reviewers
		_, _, err := client.Projects.UpdateProject("got-dev", opt)
		So(err, ShouldNotBeNil)
		errRsp, ok := err.(*ErrorResponse)
		So(ok, ShouldBeTrue)
		want := "{details: {branches: Unknown user id(s): tangyongqiang}}, {error: Bad Request}, {isValid: failed to parse unexpected error type: bool}"

		So(errRsp.Message, ShouldEqual, want)
	})
}

func TestProjectsService_UpdateProjectReviewer(t *testing.T) {
	Convey("test ProjectsService_UpdateProjectReviewer", t, func() {
		client, err := NewBasicAuthClient("root", "531E89C85298D7C86349158A64944AD8", WithBaseURL("http://10.154.0.59"))
		So(err, ShouldBeNil)

		opt := &UpdateProjectOptions{
			Name:    String("Got-dev"),
			Members: []*string{String("swarm"), String("root")},
			Branches: []*BranchOptions{
				{
					ID:       String("client"),
					Name:     String("Client"),
					Workflow: String("6"),
					Paths:    String("//Elrond.Mainline/Elrond_ArtDev/Assets/...\n//Elrond.Mainline/Elrond_ArtDev/Assets/Scripts/..."),
					Defaults: new(DefaultsOptions),
				},
			},
		}
		reviewers := make(map[string]*ReviewerOptions, 2)
		reviewers["lejiajun"] = &ReviewerOptions{Required: String("true")}
		reviewers["swarm"] = &ReviewerOptions{Required: String("false")}
		opt.Branches[0].Defaults.Reviewers = reviewers
		projects, _, err := client.Projects.UpdateProject("got-dev", opt)
		So(err, ShouldBeNil)

		want := &Project{
			ID:          "got-dev",
			Name:        "Got-dev",
			Description: "",
			Members:     []string{"root", "swarm"},
			Branches: []Branch{
				{
					ID:       "client",
					Name:     "Client",
					Workflow: "6",
					Paths:    []string{"//Elrond.Mainline/Elrond_ArtDev/Assets/...", "//Elrond.Mainline/Elrond_ArtDev/Assets/Scripts/..."},
				},
			},
		}
		users := make(map[string]interface{})
		users["lejiajun"] = map[string]interface{}{"required": true}
		users["swarm"] = []interface{}{}
		want.Branches[0].Defaults.Reviewers.Users = users

		So(projects, ShouldResemble, want)
	})
}
