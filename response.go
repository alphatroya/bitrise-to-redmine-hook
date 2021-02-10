package main

// HookResponse represents success message reponse
type HookResponse struct {
	Message  string `json:"message"`
	Success  []int  `json:"success"`
	Failures []int  `json:"failures"`
}

// NewResponse create empty response with message
func NewResponse(message string) *HookResponse {
	return &HookResponse{message, []int{}, []int{}}
}
