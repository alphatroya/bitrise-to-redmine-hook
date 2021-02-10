package main

import (
	"os"
	"strconv"
)

const (
	redisURLEnvKey = "REDIS_URL"
)

type SettingsBuilder interface {
	build() (*Settings, error)
}

type EnvSettingsBuilder struct {
}

func (e *EnvSettingsBuilder) build() (*Settings, error) {
	redis, err := getEnvVar(redisURLEnvKey)
	if err != nil {
		return nil, err
	}

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
	buildFieldID, err := strconv.ParseInt(buildFieldIDString, 10, 32)
	if err != nil {
		return nil, &HookErrorResponse{Message: "Failed to parse STAMP_BUILD_CUSTOM_FIELD parameter to int"}
	}

	nextStatus, err := getEnvVar("STAMP_DONE_STATUS")
	if err != nil {
		return nil, err
	}

	return &Settings{
		redisURL:     redis,
		host:         host,
		authToken:    authToken,
		rtbStatus:    rtbStatus,
		buildFieldID: buildFieldID,
		doneStatus:   nextStatus,
	}, nil
}

func getEnvVar(key string) (string, error) {
	val := os.Getenv(key)
	if len(val) == 0 {
		resp := NewErrorResponse(key + " ENV variable is not set")
		return "", resp
	}
	return val, nil
}
