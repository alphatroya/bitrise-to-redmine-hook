package settings

import (
	"os"
	"testing"
)

func TestSettingsBuilderFailures(t *testing.T) {
	cases := []struct {
		envs     map[string]string
		expected string
	}{
		{
			map[string]string{},
			"REDIS_URL ENV variable is not set",
		},
		{
			map[string]string{
				redisURLEnvKey: "redis",
			},
			"REDMINE_HOST ENV variable is not set",
		},
		{
			map[string]string{
				redisURLEnvKey: "redis",
				"REDMINE_HOST": "https://google.com",
			},
			"REDMINE_API_KEY ENV variable is not set",
		},
		{
			map[string]string{
				redisURLEnvKey:    "redis",
				"REDMINE_HOST":    "https://google.com",
				"REDMINE_API_KEY": "11881",
			},
			"STAMP_READY_TO_BUILD_STATUS ENV variable is not set",
		},
		{
			map[string]string{
				redisURLEnvKey:                "redis",
				"REDMINE_HOST":                "https://google.com",
				"REDMINE_API_KEY":             "11881",
				"STAMP_READY_TO_BUILD_STATUS": "1",
			},
			"STAMP_BUILD_CUSTOM_FIELD ENV variable is not set",
		},
		{
			map[string]string{
				redisURLEnvKey:                "redis",
				"REDMINE_HOST":                "https://google.com",
				"REDMINE_API_KEY":             "11881",
				"STAMP_READY_TO_BUILD_STATUS": "1",
				"STAMP_BUILD_CUSTOM_FIELD":    "b",
			},
			"Failed to parse STAMP_BUILD_CUSTOM_FIELD parameter to int",
		},
		{
			map[string]string{
				redisURLEnvKey:                "redis",
				"REDMINE_HOST":                "https://google.com",
				"REDMINE_API_KEY":             "11881",
				"STAMP_READY_TO_BUILD_STATUS": "1",
				"STAMP_BUILD_CUSTOM_FIELD":    "1",
			},
			"STAMP_DONE_STATUS ENV variable is not set",
		},
	}

	for _, tc := range cases {
		os.Clearenv()
		for key, value := range tc.envs {
			_ = os.Setenv(key, value)
		}
		_, err := Current()
		if err == nil {
			t.Errorf("Build settings should fail if envs %v set", tc.envs)
		}
		if err.Error() != tc.expected {
			t.Errorf("Wrong error received\ngot %s\nexpected %s", err.Error(), tc.expected)
		}
	}
}

func TestSettingsBuilderSuccess(t *testing.T) {
	cases := []struct {
		envs     map[string]string
		expected *Settings
	}{
		{
			map[string]string{
				redisURLEnvKey:                "redis",
				"REDMINE_HOST":                "https://google.com",
				"REDMINE_API_KEY":             "11881",
				"STAMP_READY_TO_BUILD_STATUS": "1",
				"STAMP_BUILD_CUSTOM_FIELD":    "1",
				"STAMP_DONE_STATUS":           "1222",
			},
			&Settings{
				"redis",
				"https://google.com",
				"11881",
				"1",
				1,
				"1222",
			},
		},
	}

	for _, tc := range cases {
		os.Clearenv()
		for key, value := range tc.envs {
			_ = os.Setenv(key, value)
		}
		received, err := Current()
		if err != nil {
			t.Errorf("Build settings should succeed, received error: %s", err)
		}
		// TODO: replace by reflection
		if tc.expected.RedisURL != received.RedisURL ||
			tc.expected.Host != received.Host ||
			tc.expected.AuthToken != received.AuthToken ||
			tc.expected.RtbStatus != received.RtbStatus ||
			tc.expected.BuildFieldID != received.BuildFieldID ||
			tc.expected.DoneStatus != received.DoneStatus {
			t.Errorf("Wrong settings received\ngot %+v\nexpected %+v", received, tc.expected)
		}
	}
}
