package main

type HookResponse struct {
	Message string `json:"message"`
}

func NewResponse(message string) *HookResponse {
	return &HookResponse{message}
}
