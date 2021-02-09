package main

import (
	"errors"
	"testing"
)

func TestValidateInternalHookFailedValidation(t *testing.T) {
	cases := []struct {
		sut HookPayload
		err error
	}{
		{
			HookPayload{
				BuildSlug:              "",
				BuildNumber:            0,
				BuildStatus:            0,
				BuildTriggeredWorkflow: "test",
			},
			errors.New("Skipping done transition: build workflow is not internal"),
		},
	}

	for i, tc := range cases {
		err := tc.sut.ValidateInternal()
		if err == nil {
			t.Errorf("Test case #%d should fail", i)
			break
		}
		if err.Error() != tc.err.Error() {
			t.Errorf("Hook payload was wrong, expected: %s\nreceived: %s", tc.err, err)
		}
	}
}

func TestValidateInternalHookSuccessValidation(t *testing.T) {
	cases := []struct {
		sut HookPayload
	}{
		{
			HookPayload{
				BuildSlug:              "",
				BuildNumber:            0,
				BuildStatus:            0,
				BuildTriggeredWorkflow: "internal",
			},
		},
	}

	for i, tc := range cases {
		err := tc.sut.ValidateInternal()
		if err != nil {
			t.Errorf("Test case #%d should be succeed", i)
		}
	}
}

func TestValidateInternalAndSuccessHookFailedValidation(t *testing.T) {
	cases := []struct {
		sut HookPayload
		err error
	}{
		{
			HookPayload{
				BuildSlug:              "",
				BuildNumber:            0,
				BuildStatus:            0,
				BuildTriggeredWorkflow: "test",
			},
			errors.New("Skipping done transition: build workflow is not internal"),
		},
		{
			HookPayload{
				BuildSlug:              "",
				BuildNumber:            0,
				BuildStatus:            0,
				BuildTriggeredWorkflow: "internal",
			},
			errors.New("Skipping done transition: build status is not success"),
		},
	}

	for i, tc := range cases {
		err := tc.sut.ValidateInternalAndSuccess()
		if err == nil {
			t.Errorf("Test case #%d should fail", i)
			break
		}
		if err.Error() != tc.err.Error() {
			t.Errorf("Hook payload was wrong, expected: %s\nreceived: %s", tc.err, err)
		}
	}
}

func TestValidateInternalAndSuccessHookSuccessValidation(t *testing.T) {
	cases := []struct {
		sut HookPayload
	}{
		{
			HookPayload{
				BuildSlug:              "",
				BuildNumber:            0,
				BuildStatus:            1,
				BuildTriggeredWorkflow: "internal",
			},
		},
	}

	for i, tc := range cases {
		err := tc.sut.ValidateInternalAndSuccess()
		if err != nil {
			t.Errorf("Test case #%d should be succeed", i)
		}
	}
}
