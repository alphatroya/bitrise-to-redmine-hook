package main

import (
	"net/http"
	"testing"
)

type RWMock struct {
	header http.Header
}

func newRWMock() *RWMock {
	return &RWMock{make(http.Header)}
}

func (r *RWMock) Header() http.Header {
	return r.header
}

func (r *RWMock) Write([]byte) (int, error) {
	return 0, nil
}

func (r *RWMock) WriteHeader(statusCode int) {

}

func TestResponseHeaders(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "", nil)
	rw := newRWMock()
	handler(rw, req)
	if rw.Header().Get("Content-Type") != "application/json" {
		t.Error("response headers is not application/json")
	}
}
