package main

import "errors"

type HookPayload struct {
	BuildSlug              string `json:"build_slug"`
	BuildNumber            int    `json:"build_number"`
	BuildStatus            int    `json:"build_status"`
	BuildTriggeredWorkflow string `json:"build_triggered_workflow"`
}

// ValidateInternal check out hook payload for only internal events
func (h *HookPayload) ValidateInternal() error {
	if h.BuildTriggeredWorkflow != "internal" {
		return errors.New("Skipping done transition: build workflow is not internal")
	}

	return nil
}

// ValidateInternalAndSuccess check out hook payload for only internal and success events
func (h *HookPayload) ValidateInternalAndSuccess() error {
	if err := h.ValidateInternal(); err != nil {
		return err
	}

	if h.BuildStatus != 1 {
		return errors.New("Skipping done transition: build status is not success")
	}

	return nil
}
