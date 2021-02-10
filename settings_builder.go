package main

import (
	"os"
	"strconv"
)

type SettingsBuilder interface {
	build() (*Settings, error)
}

type EnvSettingsBuilder struct {
}

func (e *EnvSettingsBuilder) build() (*Settings, error) {
	host, err := getEnvVar("REDMINE_HOST")
	if err != nil {
		return nil, err
	}

	authToken, err := getEnvVar("REDMINE_API_KEY")
	if err != nil {
		return nil, err
	}

	rtbStatus, err := getEnvVar("STAMP_READY_TO_BUILD_STATUS")
	if err != nil {
		return nil, err
	}

	buildFieldIDString, err := getEnvVar("STAMP_BUILD_CUSTOM_FIELD")
	if err != nil {
		return nil, err
	}
	buildFieldID, parseErr := strconv.ParseInt(buildFieldIDString, 10, 32)
	if parseErr != nil {
		return nil, &HookErrorResponse{Message: "Failed to parse STAMP_BUILD_CUSTOM_FIELD parameter to int"}
	}

	nextStatus, err := getEnvVar("STAMP_DONE_STATUS")
	if err != nil {
		return nil, err
	}

	return &Settings{
		host:         host,
		authToken:    authToken,
		rtbStatus:    rtbStatus,
		buildFieldID: buildFieldID,
		doneStatus:   nextStatus,
	}, nil
}

func getEnvVar(key string) (string, *HookErrorResponse) {
	val := os.Getenv(key)
	if len(val) == 0 {
		resp := NewErrorResponse(key + " ENV variable is not set")
		return "", resp
	}
	return val, nil
}
