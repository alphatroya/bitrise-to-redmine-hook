package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStamperRequestRedmineProjectKeyCheckFailure(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "", nil)
	rw := httptest.NewRecorder()
	handler := NewStamper(nil, nil)
	handler.ServeHTTP(rw, req)
	if rw.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Response status code should be 400 on failure, received %d", rw.Result().StatusCode)
	}
}

func TestStamperRequestBadPayloadCheckFailure(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "", badBody{})
	req.Header.Set("REDMINE_PROJECT", "11")
	rw := httptest.NewRecorder()
	handler := NewStamper(nil, nil)
	handler.ServeHTTP(rw, req)
	if rw.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Response status code should be 400 on bad payload, received %d", rw.Result().StatusCode)
	}
}

func TestStamperRequestRedmineProjectKeyCheckSuccess(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "", newMockBody(`{"build_triggered_workflow":"internal", "build_status":1, "build_number":12}`))
	req.Header.Set("REDMINE_PROJECT", "11")
	rw := httptest.NewRecorder()
	handler := NewStamper(nil, nil)
	handler.ServeHTTP(rw, req)
	if rw.Result().StatusCode != http.StatusOK {
		t.Errorf("Response status code should be 200 on success, received %d", rw.Result().StatusCode)
	}
}
