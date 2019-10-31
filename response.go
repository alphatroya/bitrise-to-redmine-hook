package main

type HookResponse struct {
	Message  string `json:"message"`
	Success  []int  `json:"success"`
	Failures []int  `json:"failures"`
}

func NewResponse(message string) *HookResponse {
	return &HookResponse{message, []int{}, []int{}}
}
