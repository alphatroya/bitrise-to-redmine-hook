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
	handlerV1       *Handler
	rdb             *redis.Client
}

func NewHandlerV2(settingsBuilder SettingsBuilder, redisUrl string) *HandlerV2 {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisUrl,
		Password: "",
		DB:       0,
	})
	return &HandlerV2{settingsBuilder, &Handler{settingsBuilder}, rdb}
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

	data, err := json.Marshal(issues)
	t.rdb.Set(payload.BuildSlug, data, time.Hour)
}

func (t *HandlerV2) handleFinishedEvent(w http.ResponseWriter, r *http.Request) {
	payload, errResponse, err := t.readPayload(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errResponse)
		return
	}

	cached, err := t.rdb.Get(payload.BuildSlug).Result()
	if err != nil {
		t.handlerV1.ServeHTTP(w, r)
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

	issues := new(IssuesList)
	json.Unmarshal([]byte(cached), issues)
	response := batchTransaction(RedmineDoneMarker{}, issues, settings, payload.BuildNumber)
	_ = sendMailgunNotification(response, settings.host, payload.BuildNumber, issues.Issues, "v2")

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
