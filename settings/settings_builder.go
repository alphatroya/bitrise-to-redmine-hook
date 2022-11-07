package settings

import (
	"errors"
	"os"
	"strconv"
)

const (
	redisURLEnvKey = "REDIS_URL"
)

func Current() (*Settings, error) {
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
		return nil, errors.New("Failed to parse STAMP_BUILD_CUSTOM_FIELD parameter to int")
	}

	nextStatus, err := getEnvVar("STAMP_DONE_STATUS")
	if err != nil {
		return nil, err
	}

	return &Settings{
		RedisURL:     redis,
		Host:         host,
		AuthToken:    authToken,
		RtbStatus:    rtbStatus,
		BuildFieldID: buildFieldID,
		DoneStatus:   nextStatus,
	}, nil
}

func getEnvVar(key string) (string, error) {
	val := os.Getenv(key)
	if len(val) == 0 {
		return "", errors.New(key + " ENV variable is not set")
	}
	return val, nil
}
