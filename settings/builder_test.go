package settings

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_Current(t *testing.T) {
	cases := []struct {
		envs       map[string]string
		expected   *Config
		shouldFail bool
	}{
		{
			envs:       map[string]string{},
			shouldFail: true,
		},
		{
			envs: map[string]string{
				"REDIS_URL": "redis",
			},
			shouldFail: true,
		},
		{
			envs: map[string]string{
				"REDIS_URL":    "redis",
				"REDMINE_HOST": "https://google.com",
			},
			shouldFail: true,
		},
		{
			envs: map[string]string{
				"REDIS_URL":       "redis",
				"REDMINE_HOST":    "https://google.com",
				"REDMINE_API_KEY": "11881",
			},
			shouldFail: true,
		},
		{
			envs: map[string]string{
				"REDIS_URL":                   "redis",
				"REDMINE_HOST":                "https://google.com",
				"REDMINE_API_KEY":             "11881",
				"STAMP_READY_TO_BUILD_STATUS": "1",
			},
			shouldFail: true,
		},
		{
			envs: map[string]string{
				"REDIS_URL":                   "redis",
				"REDMINE_HOST":                "https://google.com",
				"REDMINE_API_KEY":             "11881",
				"STAMP_READY_TO_BUILD_STATUS": "1",
				"STAMP_BUILD_CUSTOM_FIELD":    "b",
			},
			shouldFail: true,
		},
		{
			envs: map[string]string{
				"REDIS_URL":                   "redis",
				"REDMINE_HOST":                "https://google.com",
				"REDMINE_API_KEY":             "11881",
				"STAMP_READY_TO_BUILD_STATUS": "1",
				"STAMP_BUILD_CUSTOM_FIELD":    "1",
			},
			shouldFail: true,
		},
		{
			envs: map[string]string{
				"REDIS_URL":                   "redis",
				"REDMINE_HOST":                "https://google.com",
				"REDMINE_API_KEY":             "11881",
				"STAMP_READY_TO_BUILD_STATUS": "1",
				"STAMP_BUILD_CUSTOM_FIELD":    "1",
				"STAMP_DONE_STATUS":           "1222",
			},
			expected: &Config{
				RedisURL:     "redis",
				Host:         "https://google.com",
				AuthToken:    "11881",
				RtbStatus:    "1",
				BuildFieldID: 1,
				DoneStatus:   "1222",
			},
		},
	}

	for _, tt := range cases {
		os.Clearenv()
		for key, value := range tt.envs {
			_ = os.Setenv(key, value)
		}

		received, err := Current()
		if (err != nil) != tt.shouldFail {
			t.Fatalf("Build settings should succeed, received error: %s", err)
		}
		if diff := cmp.Diff(received, tt.expected); diff != "" {
			t.Fatalf("TestSettingsBuilderSuccess fail, diff: %s", diff)
		}
	}
}
