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
		t.Error("response headers is not application/json")
	}
}
