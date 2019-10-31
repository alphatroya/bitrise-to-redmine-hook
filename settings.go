package main

import (
	"os"
	"strconv"
)

type Settings struct {
	host         string
	authToken    string
	rtbStatus    string
	buildFieldID int64
	doneStatus   string
}

func NewSettings() (*Settings, *HookErrorResponse) {
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
	authToken := os.Getenv(key)
	if len(authToken) == 0 {
		errJSON := NewErrorResponse(key + " ENV variable is not set")
		return "", errJSON
	}
	return authToken, nil
}
