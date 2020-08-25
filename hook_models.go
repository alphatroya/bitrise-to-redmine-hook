package main

type HookPayload struct {
	BuildSlug              string `json:"build_slug"`
	BuildNumber            int    `json:"build_number"`
	AppSlug                string `json:"app_slug"`
	BuildStatus            int    `json:"build_status"`
	BuildTriggeredWorkflow string `json:"build_triggered_workflow"`
	Git                    struct {
		Provider      string      `json:"provider"`
		SrcBranch     string      `json:"src_branch"`
		DstBranch     string      `json:"dst_branch"`
		PullRequestID int         `json:"pull_request_id"`
		Tag           interface{} `json:"tag"`
	} `json:"git"`
}
