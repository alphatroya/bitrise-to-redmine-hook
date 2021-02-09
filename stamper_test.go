package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStamperResponseJSONHeader(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "", nil)
	rw := httptest.NewRecorder()
	handler := NewStamper(nil, nil)
	handler.ServeHTTP(rw, req)
	if rw.Header().Get("Content-Type") != "application/json" {
		t.Error("Response headers is not application/json")
	}
}

func TestStamperRequestRedmineProjectKeyCheckFailure(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "", nil)
	rw := httptest.NewRecorder()
	handler := NewStamper(nil, nil)
	handler.ServeHTTP(rw, req)
	if rw.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Response status code should be 400 on failure, received %d", rw.Result().StatusCode)
	}
}

func TestStamperRequestRedmineProjectKeyCheckSuccess(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "", nil)
	req.Header.Set("REDMINE_PROJECT", "11")
	rw := httptest.NewRecorder()
	handler := NewStamper(nil, nil)
	handler.ServeHTTP(rw, req)
	if rw.Result().StatusCode != http.StatusOK {
		t.Errorf("Response status code should be 200 on success, received %d", rw.Result().StatusCode)
	}
}
