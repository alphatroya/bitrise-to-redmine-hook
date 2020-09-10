package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Handler struct {
	settingsBuilder SettingsBuilder
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("Bitrise-Event-Type") != "build/finished" {
		json.NewEncoder(w).Encode(NewResponse("Skipping done transition: build status is not finished"))
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errJSON := NewErrorResponse("Received wrong request data payload")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errJSON)
		return
	}

	payload := new(HookPayload)
	err = json.Unmarshal(data, payload)
	if err != nil {
		errJSON := NewErrorResponse("Can't decode request payload json data")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errJSON)
		return
	}

	if payload.BuildTriggeredWorkflow != "internal" {
		json.NewEncoder(w).Encode(NewResponse("Skipping done transition: build workflow is not internal"))
		return
	}

	if payload.BuildStatus != 1 {
		json.NewEncoder(w).Encode(NewResponse("Skipping done transition: build status is not success"))
		return
	}

	settings, errorResponse := h.settingsBuilder.build()
	if errorResponse != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	redmineProject := r.Header.Get("REDMINE_PROJECT")
	if len(redmineProject) == 0 {
		errJSON := NewErrorResponse("REDMINE_PROJECT header is not set")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errJSON)
		return
	}

	issues, err := issues(settings, redmineProject)
	if err != nil {
		errJSON := NewErrorResponse(fmt.Sprintf("Wrong error from server: %s", err))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errJSON)
		return
	}

	response := NewResponse("Successful completed task")
	for _, issue := range issues.Issues {
		err = markAsDone(issue, settings, payload.BuildNumber)
		if err != nil {
			response.Failures = append(response.Failures, issue.ID)
			continue
		}
		response.Success = append(response.Success, issue.ID)
	}
	_ = sendMailgunNotification(response, settings.host, payload.BuildNumber, issues.Issues)

	json.NewEncoder(w).Encode(response)
}

func markAsDone(issue *Issue, settings *Settings, buildNumber int) error {
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

func issues(settings *Settings, project string) (*IssuesList, error) {
	request, err := http.NewRequest("GET", settings.host+"/issues.json?status_id="+settings.rtbStatus+"&project_id="+project, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("X-Redmine-API-Key", settings.authToken)
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode >= 400 {
		return nil, fmt.Errorf("Received wrong status code %d", response.StatusCode)
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	result := new(IssuesList)
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
