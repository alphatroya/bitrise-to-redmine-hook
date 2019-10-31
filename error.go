package main

type HookErrorResponse struct {
	Message string `json:"message"`
}

func NewErrorResponse(message string) *HookErrorResponse {
	return &HookErrorResponse{message}
}

func (h *HookErrorResponse) Error() string {
	return h.Message
}
