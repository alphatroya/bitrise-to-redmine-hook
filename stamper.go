package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-redis/redis"
)

// Stamper is a handler for moving ready to build tasks to done state
type Stamper struct {
	settingsBuilder SettingsBuilder
	rdb             *redis.Client
}

// NewStamper creates handler class configured by settings and connected to redis client
func NewStamper(settingsBuilder SettingsBuilder, redisURL string) (*Stamper, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	rdb := redis.NewClient(options)
	_, err = rdb.Ping().Result()
	if err != nil {
		return nil, err
	}
	return &Stamper{settingsBuilder, rdb}, nil
}

func (t *Stamper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Header.Get("Bitrise-Event-Type") {
	case "build/triggered":
		t.handleTriggeredEvent(w, r)
	case "build/finished":
		t.handleFinishedEvent(w, r)
	}
}

func (t *Stamper) writeErrResponse(w http.ResponseWriter, statusCode int, err error) {
	t.writeResponse(w, statusCode, err.Error())
}

func (t *Stamper) writeResponse(w http.ResponseWriter, statusCode int, message string) {
	messageJSON := NewErrorResponse(message)
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(messageJSON)
}

func (t *Stamper) handleTriggeredEvent(w http.ResponseWriter, r *http.Request) {
	payload, err := t.readPayload(r)
	if err != nil {
		t.writeErrResponse(w, http.StatusBadRequest, err)
		return
	}

	if err = payload.ValidateInternal(); err != nil {
		t.writeErrResponse(w, http.StatusOK, err)
		return
	}

	settings, errorResponse := t.settingsBuilder.build()
	if errorResponse != nil {
		t.writeResponse(w, http.StatusInternalServerError, errorResponse.Message)
		return
	}

	redmineProject := r.Header.Get("REDMINE_PROJECT")
	if len(redmineProject) == 0 {
		t.writeResponse(w, http.StatusBadRequest, "REDMINE_PROJECT header is not set")
		return
	}

	issues, err := issues(settings, redmineProject)
	if err != nil {
		t.writeResponse(w, http.StatusBadRequest, fmt.Sprintf("Wrong error from server: %s", err))
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

func (t *Stamper) handleFinishedEvent(w http.ResponseWriter, r *http.Request) {
	payload, err := t.readPayload(r)
	if err != nil {
		t.writeErrResponse(w, http.StatusBadRequest, err)
		return
	}

	redmineProject := r.Header.Get("REDMINE_PROJECT")
	if len(redmineProject) == 0 {
		t.writeResponse(w, http.StatusBadRequest, "REDMINE_PROJECT header is not set")
		return
	}

	if err = payload.ValidateInternalAndSuccess(); err != nil {
		t.writeErrResponse(w, http.StatusOK, err)
		return
	}

	settings, errorResponse := t.settingsBuilder.build()
	if errorResponse != nil {
		t.writeResponse(w, http.StatusInternalServerError, errorResponse.Message)
		return
	}

	cached, err := t.rdb.Get(payload.BuildSlug).Result()
	var issuesList *IssuesList
	version := "v2"
	if err != nil {
		issuesList, err = issues(settings, redmineProject)
		if err != nil {
			t.writeResponse(w, http.StatusBadRequest, fmt.Sprintf("Wrong error from server: %s", err))
			return
		}
	} else {
		version += " cached"
		issuesList = new(IssuesList)
		json.Unmarshal([]byte(cached), issuesList)
	}

	response := batchTransaction(RedmineDoneMarker{}, issuesList, settings, payload.BuildNumber)
	_ = sendMailgunNotification(response, settings.host, payload.BuildNumber, issuesList.Issues, version)

	json.NewEncoder(w).Encode(response)
}

func (t *Stamper) readPayload(r *http.Request) (*HookPayload, error) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.New("Received wrong request data payload")
	}

	payload := new(HookPayload)
	err = json.Unmarshal(data, payload)
	if err != nil {
		return nil, errors.New("Can't decode request payload json data")
	}

	return payload, nil
}
