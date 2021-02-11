package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Stamper is a handler for moving ready to build tasks to done state
type Stamper struct {
	settings *Settings
	rdb      Storage
	logger   *log.Logger
}

// NewStamper creates handler class configured by settings and connected to redis client
func NewStamper(settings *Settings, storage Storage, logger *log.Logger) *Stamper {
	return &Stamper{settings, storage, logger}
}

func (s *Stamper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp, statusCode, err := s.handleEvent(r)
	s.logger.Printf("Create new response with status code: %d", statusCode)
	if err != nil {
		s.logger.Printf("Answer response with ERROR message: %s", err.Error())
		http.Error(w, err.Error(), statusCode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	s.logger.Printf("Answer response with SUCCESS message: %+v", resp)
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Stamper) handleEvent(r *http.Request) (*HookResponse, int, error) {
	rp := r.Header.Get("REDMINE_PROJECT")
	if len(rp) == 0 {
		return nil, http.StatusBadRequest, errors.New("REDMINE_PROJECT header is absent in the hook header")
	}

	payload, err := s.readAndParsePayload(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	et := r.Header.Get("Bitrise-Event-Type")
	s.logger.Printf("Received bitrise event %s", et)
	switch et {
	case "build/triggered":
		return s.handleTriggeredEvent(payload, rp)
	case "build/finished":
		return s.handleFinishedEvent(payload, rp)
	default:
		return nil, http.StatusOK, fmt.Errorf("Unsupported bitrise event type %s", et)
	}
}

func (s *Stamper) handleTriggeredEvent(payload *HookPayload, redmineProject string) (*HookResponse, int, error) {
	if err := payload.ValidateInternal(); err != nil {
		return nil, http.StatusOK, err
	}

	iContainer, err := issues(s.settings, redmineProject)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("Wrong error from server: %s", err)
	}

	data, err := json.Marshal(iContainer)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("Can't serialize data to string: %s", err)
	}
	err = s.rdb.Set(payload.BuildSlug, data, 4*time.Hour).Err()
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("Can't write new cache with build: %+v\nerror: %s", payload, err)
	}

	var logItems []int
	for _, issue := range iContainer.Issues {
		logItems = append(logItems, issue.ID)
	}
	return &HookResponse{fmt.Sprintf("Caching issue data was completed (Build: %s)", payload.BuildSlug), logItems, []int{}}, http.StatusOK, nil
}

func (s *Stamper) handleFinishedEvent(payload *HookPayload, redmineProject string) (*HookResponse, int, error) {
	if err := payload.ValidateInternalAndSuccess(); err != nil {
		return nil, http.StatusOK, err
	}

	cached, err := s.rdb.Get(payload.BuildSlug).Result()
	var issuesList *IssuesContainer
	version := "v2"
	if err != nil {
		issuesList, err = issues(s.settings, redmineProject)
		if err != nil {
			return nil, http.StatusBadRequest, fmt.Errorf("Wrong error from server: %s", err)
		}
	} else {
		version += " cached"
		issuesList = new(IssuesContainer)
		_ = json.Unmarshal([]byte(cached), issuesList)
	}

	response := batchTransaction(RedmineDoneMarker{}, issuesList, s.settings, payload.BuildNumber)
	_ = sendMailgunNotification(response, s.settings.host, payload.BuildNumber, issuesList.Issues, version)

	return response, http.StatusOK, nil
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
	s.logger.Printf("Received input payload json: %+v", payload)

	return payload, nil
}
