package main

type HookPayload struct {
	BuildNumber            int    `json:"build_number"`
	BuildStatus            int    `json:"build_status"`
	BuildTriggeredWorkflow string `json:"build_triggered_workflow"`
}
