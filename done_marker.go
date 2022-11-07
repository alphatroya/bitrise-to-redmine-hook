package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alphatroya/ci-redmine-bindings/settings"
)

// DoneMarker defines interface for issue processing task
type DoneMarker interface {
	markAsDone(issue *Issue, settings *settings.Settings, buildNumber int) error
}

// RedmineDoneMarker move all issues to Done state with build number printing
type RedmineDoneMarker struct{}

func (r RedmineDoneMarker) markAsDone(issue *Issue, settings *settings.Settings, buildNumber int) error {
	type PayloadCustomField struct {
		ID    int64  `json:"id"`
		Value string `json:"value"`
	}

	type PayloadIssue struct {
		AssignedToID string                `json:"assigned_to_id"`
		StatusID     string                `json:"status_id"`
		CustomFields []*PayloadCustomField `json:"custom_fields"`
	}

	type Payload struct {
		Issue *PayloadIssue `json:"issue"`
	}

	requestBody := Payload{
		Issue: &PayloadIssue{
			AssignedToID: fmt.Sprintf("%d", issue.Author.ID),
			StatusID:     settings.DoneStatus,
			CustomFields: []*PayloadCustomField{
				{settings.BuildFieldID, fmt.Sprintf("%d", buildNumber)},
			},
		},
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	buffer := bytes.NewBuffer(body)

	request, err := http.NewRequest("PUT", settings.Host+fmt.Sprintf("/issues/%d.json", issue.ID), buffer)
	if err != nil {
		return err
	}
	request.Header.Set("X-Redmine-API-Key", settings.AuthToken)
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode >= 400 {
		return fmt.Errorf("Received wrong status code %d", response.StatusCode)
	}
	return nil
}
