package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-redis/redis"
)

type HandlerV2 struct {
	settingsBuilder SettingsBuilder
	rdb             *redis.Client
}

func NewHandlerV2(settingsBuilder SettingsBuilder, redisUrl string) *HandlerV2 {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisUrl,
		Password: "",
		DB:       0,
	})
	return &HandlerV2{settingsBuilder, rdb}
}

func (t *HandlerV2) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Header.Get("Bitrise-Event-Type") {
	case "build/triggered":
		t.handleTriggeredEvent(w, r)
	case "build/finished":
		t.handleFinishedEvent(w, r)
	}
}

func (t *HandlerV2) handleTriggeredEvent(w http.ResponseWriter, r *http.Request) {
	payload, errJson, err := t.readPayload(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errJson)
		return
	}

	if payload.BuildTriggeredWorkflow != "internal" {
		json.NewEncoder(w).Encode(NewResponse("Skipping done transition: build workflow is not internal"))
		return
	}

	settings, errorResponse := t.settingsBuilder.build()
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

	// TODO: we don't need Mashal here, just use Request body here
	data, err := json.Marshal(issues)
	if err != nil {
		errJSON := NewErrorResponse(fmt.Sprintf("Can't serialize data to string: %s", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errJSON)
		return
	}
	err = t.rdb.Set(payload.BuildSlug, data, 4*time.Hour).Err()
	if err != nil {
		errJSON := NewErrorResponse(fmt.Sprintf("Can't write new cache with build: %+v\nerror: %s", payload, err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errJSON)
		return
	}

	logItems := []int{}
	for _, issue := range issues.Issues {
		logItems = append(logItems, issue.ID)
	}
	json.NewEncoder(w).Encode(HookResponse{fmt.Sprintf("Caching issue data was completed (Build: %s)", payload.BuildSlug), logItems, []int{}})
}

func (t *HandlerV2) handleFinishedEvent(w http.ResponseWriter, r *http.Request) {
	payload, errResponse, err := t.readPayload(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errResponse)
		return
	}

	redmineProject := r.Header.Get("REDMINE_PROJECT")
	if len(redmineProject) == 0 {
		errJSON := NewErrorResponse("REDMINE_PROJECT header is not set")
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

	settings, errorResponse := t.settingsBuilder.build()
	if errorResponse != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	cached, err := t.rdb.Get(payload.BuildSlug).Result()
	var issuesList *IssuesList
	var cacheType string
	if err != nil {
		cacheType = "not cached"
		issuesList, err = issues(settings, redmineProject)
		if err != nil {
			errJSON := NewErrorResponse(fmt.Sprintf("Wrong error from server: %s", err))
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errJSON)
			return
		}
	} else {
		cacheType = "cached"
		issuesList = new(IssuesList)
		json.Unmarshal([]byte(cached), issuesList)
	}

	response := batchTransaction(RedmineDoneMarker{}, issuesList, settings, payload.BuildNumber)
	_ = sendMailgunNotification(response, settings.host, payload.BuildNumber, issuesList.Issues, "v2 "+cacheType)

	json.NewEncoder(w).Encode(response)
}

func (t *HandlerV2) readPayload(r *http.Request) (*HookPayload, *HookErrorResponse, error) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, NewErrorResponse("Received wrong request data payload"), err
	}

	payload := new(HookPayload)
	err = json.Unmarshal(data, payload)
	if err != nil {
		return nil, NewErrorResponse("Can't decode request payload json data"), err
	}

	return payload, nil, nil
}
