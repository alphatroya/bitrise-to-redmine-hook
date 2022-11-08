package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/alphatroya/ci-redmine-bindings/settings"
	"github.com/rs/zerolog"
)

// Stamper is a handler for moving ready to build tasks to done state
type Stamper struct {
	settings *settings.Config
	rdb      Storage
}

// NewStamper creates handler class configured by settings and connected to redis client
func NewStamper(settings *settings.Config, storage Storage) *Stamper {
	return &Stamper{settings: settings, rdb: storage}
}

func (s *Stamper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := *zerolog.Ctx(r.Context())
	logger.Debug().
		Msg("received incomming request")

	projectID := r.Header.Get("REDMINE_PROJECT")
	if projectID == "" {
		err := errors.New("REDMINE_PROJECT header isn't set in request headers")
		logger.Error().
			Err(err).
			Msg("wrong incoming headers")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger = logger.With().Str("r_project", projectID).Logger()

	resp, statusCode, err := s.handleEvent(r.WithContext(logger.WithContext(r.Context())), projectID)
	logger.Debug().
		Int("status code", statusCode).
		Msg("create a new response")
	if err != nil {
		logger.Error().
			Err(err).
			Msg("event processing failed")
		http.Error(w, err.Error(), statusCode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	logger.Info().
		Interface("response", resp).
		Int("status code", statusCode).
		Msg("success response")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Stamper) handleEvent(r *http.Request, projectID string) (*HookResponse, int, error) {
	payload, err := s.readAndParsePayload(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	et := r.Header.Get("Bitrise-Event-Type")

	zerolog.Ctx(r.Context()).
		Debug().
		Str("bitrise event", et).
		Msg("received bitrise event header")
	switch et {
	case "build/triggered":
		return s.handleTriggeredEvent(payload, projectID)
	case "build/finished":
		return s.handleFinishedEvent(payload, projectID)
	default:
		return nil, http.StatusOK, fmt.Errorf("handleEvent: unsupported bitrise event type %s", et)
	}
}

func (s *Stamper) handleTriggeredEvent(payload *HookPayload, redmineProject string) (*HookResponse, int, error) {
	if err := payload.ValidateInternal(); err != nil {
		return nil, http.StatusOK, err
	}

	iContainer, err := issues(s.settings, redmineProject)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("handleTriggeredEvent: wrong error from server: %s", err)
	}

	data, err := json.Marshal(iContainer)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("handleTriggeredEvent: can't serialize data to string: %s", err)
	}
	err = s.rdb.Set(payload.BuildSlug, data, 4*time.Hour).Err()
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("handleTriggeredEvent: can't write new cache with build: %+v\nerror: %s", payload, err)
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
			return nil, http.StatusBadRequest, fmt.Errorf("handleFinishedEvent: wrong error from server: %w", err)
		}
	} else {
		version += " cached"
		issuesList = new(IssuesContainer)
		_ = json.Unmarshal([]byte(cached), issuesList)
	}

	response := batchTransaction(RedmineDoneMarker{}, issuesList, s.settings, payload.BuildNumber)
	_ = sendMailgunNotification(response, s.settings.Host, payload.BuildNumber, issuesList.Issues, version)

	return response, http.StatusOK, nil
}

func (s *Stamper) readAndParsePayload(r *http.Request) (*HookPayload, error) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("readAndParsePayload: received wrong request data payload: %w", err)
	}

	payload := new(HookPayload)
	err = json.Unmarshal(data, payload)
	if err != nil {
		return nil, fmt.Errorf("readAndParsePayload: can't decode request payload json data: %w", err)
	}

	zerolog.Ctx(r.Context()).
		Debug().
		Interface("in json", payload).
		Msg("unmarshal event payload")

	return payload, nil
}
