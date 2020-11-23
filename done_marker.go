package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type DoneMarker interface {
	markAsDone(issue *Issue, settings *Settings, buildNumber int) error
}

type RedmineDoneMarker struct{}

func (r RedmineDoneMarker) markAsDone(issue *Issue, settings *Settings, buildNumber int) error {
	type PayloadCustomField struct {
		ID    int64  `json:"id"`
		Value string `json:"value"`
	}

	type PayloadIssue struct {
		AssignedToId string                `json:"assigned_to_id"`
		StatusId     string                `json:"status_id"`
		CustomFields []*PayloadCustomField `json:"custom_fields"`
	}

	type Payload struct {
		Issue *PayloadIssue `json:"issue"`
	}

	requestBody := Payload{
		Issue: &PayloadIssue{
			AssignedToId: fmt.Sprintf("%d", issue.Author.ID),
			StatusId:     settings.doneStatus,
			CustomFields: []*PayloadCustomField{
				{settings.buildFieldID, fmt.Sprintf("%d", buildNumber)},
			},
		},
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	buffer := bytes.NewBuffer(body)

	request, err := http.NewRequest("PUT", settings.host+fmt.Sprintf("/issues/%d.json", issue.ID), buffer)
	if err != nil {
		return err
	}
	request.Header.Set("X-Redmine-API-Key", settings.authToken)
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
