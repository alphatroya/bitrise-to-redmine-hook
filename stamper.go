package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Stamper is a handler for moving ready to build tasks to done state
type Stamper struct {
	settings *Settings
	rdb      Storage
}

// NewStamper creates handler class configured by settings and connected to redis client
func NewStamper(settings *Settings, storage Storage) *Stamper {
	return &Stamper{settings, storage}
}

func (s *Stamper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rp := r.Header.Get("REDMINE_PROJECT")
	if len(rp) == 0 {
		s.writeErrResponse(w, http.StatusBadRequest, errors.New("REDMINE_PROJECT header is not set"))
		return
	}

	payload, err := s.readAndParsePayload(r)
	if err != nil {
		s.writeErrResponse(w, http.StatusBadRequest, err)
	}

	statusCode := http.StatusOK
	switch r.Header.Get("Bitrise-Event-Type") {
	case "build/triggered":
		statusCode, err = s.handleTriggeredEvent(w, payload, rp)
	case "build/finished":
		statusCode, err = s.handleFinishedEvent(w, payload, rp)
	}
	if err != nil {
		s.writeErrResponse(w, statusCode, err)
	}
}

func (s *Stamper) handleTriggeredEvent(w http.ResponseWriter, payload *HookPayload, redmineProject string) (int, error) {
	if err := payload.ValidateInternal(); err != nil {
		return http.StatusOK, err
	}

	iContainer, err := issues(s.settings, redmineProject)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("Wrong error from server: %s", err)
	}

	data, err := json.Marshal(iContainer)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Can't serialize data to string: %s", err)
	}
	err = s.rdb.Set(payload.BuildSlug, data, 4*time.Hour).Err()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Can't write new cache with build: %+v\nerror: %s", payload, err)
	}

	var logItems []int
	for _, issue := range iContainer.Issues {
		logItems = append(logItems, issue.ID)
	}
	json.NewEncoder(w).Encode(HookResponse{fmt.Sprintf("Caching issue data was completed (Build: %s)", payload.BuildSlug), logItems, []int{}})
	return http.StatusOK, nil
}

func (s *Stamper) handleFinishedEvent(w http.ResponseWriter, payload *HookPayload, redmineProject string) (int, error) {
	if err := payload.ValidateInternalAndSuccess(); err != nil {
		return http.StatusOK, err
	}

	cached, err := s.rdb.Get(payload.BuildSlug).Result()
	var issuesList *IssuesContainer
	version := "v2"
	if err != nil {
		issuesList, err = issues(s.settings, redmineProject)
		if err != nil {
			return http.StatusBadRequest, fmt.Errorf("Wrong error from server: %s", err)
		}
	} else {
		version += " cached"
		issuesList = new(IssuesContainer)
		_ = json.Unmarshal([]byte(cached), issuesList)
	}

	response := batchTransaction(RedmineDoneMarker{}, issuesList, s.settings, payload.BuildNumber)
	_ = sendMailgunNotification(response, s.settings.host, payload.BuildNumber, issuesList.Issues, version)

	json.NewEncoder(w).Encode(response)
	return http.StatusOK, nil
}

func (s *Stamper) readAndParsePayload(r *http.Request) (*HookPayload, error) {
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

func (s *Stamper) writeErrResponse(w http.ResponseWriter, statusCode int, err error) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(NewErrorResponse(err.Error()))
}
