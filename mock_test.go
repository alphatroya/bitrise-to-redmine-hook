package main

import (
	"bytes"
	"io"
)

type badBody struct{}

func (b badBody) Read(p []byte) (n int, err error) {
	return 0, bytes.ErrTooLarge
}

type mockBody struct {
	p string
}

func newMockBody(p string) mockBody {
	return mockBody{p}
}

func (m mockBody) Read(p []byte) (n int, err error) {
	for i, b := range []byte(m.p) {
		p[i] = b
	}
	return len(m.p), io.EOF
}
