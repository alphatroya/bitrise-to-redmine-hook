package main

import (
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

	if err = payload.ValidateInternalAndSuccess(); err != nil {
		json.NewEncoder(w).Encode(NewResponse(err.Error()))
		return
	}

	settings, err := h.settingsBuilder.build()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
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

	response := batchTransaction(RedmineDoneMarker{}, issues, settings, payload.BuildNumber)
	_ = sendMailgunNotification(response, settings.host, payload.BuildNumber, issues.Issues, "v1")

	json.NewEncoder(w).Encode(response)
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
