package main

// HookErrorResponse represents main error response json structure
type HookErrorResponse struct {
	Message string `json:"message"`
}

// NewErrorResponse creates response object with error message
func NewErrorResponse(message string) *HookErrorResponse {
	return &HookErrorResponse{message}
}

func (h *HookErrorResponse) Error() string {
	return h.Message
}
