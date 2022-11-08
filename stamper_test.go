package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

var logger = log.New(ioutil.Discard, "", 0)

func TestStamperRequestRedmineProjectKeyCheckFailure(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "", nil)
	rw := httptest.NewRecorder()
	handler := NewStamper(nil, nil, logger)
	handler.ServeHTTP(rw, req)
	if rw.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Response status code should be 400 on failure, received %d", rw.Result().StatusCode)
	}
}

func TestStamperRequestBadPayloadCheckFailure(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "", badBody{})
	req.Header.Set("REDMINE_PROJECT", "11")
	rw := httptest.NewRecorder()
	handler := NewStamper(nil, nil, logger)
	handler.ServeHTTP(rw, req)
	if rw.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Response status code should be 400 on bad payload, received %d", rw.Result().StatusCode)
	}
}

func TestStamperRequestRedmineProjectKeyCheckSuccess(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "", newMockBody(`{"build_triggered_workflow":"internal", "build_status":1, "build_number":12}`))
	req.Header.Set("REDMINE_PROJECT", "11")
	rw := httptest.NewRecorder()
	handler := NewStamper(nil, nil, logger)
	handler.ServeHTTP(rw, req)
	if rw.Result().StatusCode != http.StatusOK {
		t.Errorf("Response status code should be 200 on success, received %d", rw.Result().StatusCode)
	}
}

func TestStamperRequestTriggeredEventNonInternal(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "", newMockBody(`{"build_triggered_workflow":"test", "build_status":1, "build_number":12}`))
	req.Header.Set("REDMINE_PROJECT", "11")
	req.Header.Set("Bitrise-Event-Type", "build/triggered")
	rw := httptest.NewRecorder()
	handler := NewStamper(nil, nil, logger)
	handler.ServeHTTP(rw, req)
	if rw.Result().StatusCode != http.StatusOK {
		t.Errorf("Response status code should be 200 on success, received %d", rw.Result().StatusCode)
	}

	resp := rw.Body.String()
	expected := "Skipping done transition: build workflow is not internal\n"
	if resp != expected {
		t.Errorf("Response body message wrong\nreceived: %q\nexpected: %q", resp, expected)
	}
}
