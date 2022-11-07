package main

import (
	"errors"
	"testing"

	"github.com/alphatroya/ci-redmine-bindings/settings"
)

func TestBatchSuccessTransaction(t *testing.T) {
	m := MockDoneMarker{}
	s := &settings.Settings{}
	il := &IssuesContainer{
		[]*Issue{
			{},
			{},
			{},
			{},
			{},
		},
	}
	res := batchTransaction(m, il, s, 5)
	if len(res.Success) != len(il.Issues) {
		t.Errorf("Error during test expect: %d\nreceived: %d", len(il.Issues), len(res.Success))
	}
}

func TestBatchFailTransaction(t *testing.T) {
	m := MockDoneMarker{true}
	s := &settings.Settings{}
	il := &IssuesContainer{
		[]*Issue{
			{},
			{},
			{},
			{},
			{},
		},
	}
	res := batchTransaction(m, il, s, 5)
	if len(res.Failures) != len(il.Issues) {
		t.Errorf("Error during test expect: %d\nreceived: %d", len(il.Issues), len(res.Failures))
	}
}

type MockDoneMarker struct {
	failable bool
}

func (m MockDoneMarker) markAsDone(issue *Issue, settings *settings.Settings, buildNumber int) error {
	if m.failable {
		return errors.New("Fail")
	}
	return nil
}
