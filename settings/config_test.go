package settings

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_Current(t *testing.T) {
	cases := []struct {
		name       string
		envs       map[string]string
		expected   *Config
		shouldFail bool
	}{
		{
			name:       "all empty",
			envs:       map[string]string{},
			shouldFail: true,
		},
		{
			name: "only redis url set",
			envs: map[string]string{
				"REDIS_URL": "redis",
			},
			shouldFail: true,
		},
		{
			name: "redis and redmine url set",
			envs: map[string]string{
				"REDIS_URL":    "redis",
				"REDMINE_HOST": "https://google.com",
			},
			shouldFail: true,
		},
		{
			name: "redis + redmine url and key set",
			envs: map[string]string{
				"REDIS_URL":       "redis",
				"REDMINE_HOST":    "https://google.com",
				"REDMINE_API_KEY": "11881",
			},
			shouldFail: true,
		},
		{
			name: "redis + redmine url and key + stamp set",
			envs: map[string]string{
				"REDIS_URL":                   "redis",
				"REDMINE_HOST":                "https://google.com",
				"REDMINE_API_KEY":             "11881",
				"STAMP_READY_TO_BUILD_STATUS": "1",
			},
			shouldFail: true,
		},
		{
			name: "redis + redmine url and key + stamp + wrong build number set",
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
			name: "redis + redmine url and key + stamp + build number set",
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
			name: "redis + redmine url and key + stamp + build number set + sentry",
			envs: map[string]string{
				"REDIS_URL":                   "redis",
				"REDMINE_HOST":                "https://google.com",
				"REDMINE_API_KEY":             "11881",
				"STAMP_READY_TO_BUILD_STATUS": "1",
				"STAMP_BUILD_CUSTOM_FIELD":    "1",
				"SENTRY_DSN":                  "sentry",
			},
			shouldFail: true,
		},
		{
			name: "all required settings set without port",
			envs: map[string]string{
				"REDIS_URL":                   "redis",
				"REDMINE_HOST":                "https://google.com",
				"REDMINE_API_KEY":             "11881",
				"STAMP_READY_TO_BUILD_STATUS": "1",
				"STAMP_BUILD_CUSTOM_FIELD":    "1",
				"STAMP_DONE_STATUS":           "1222",
				"SENTRY_DSN":                  "sentry",
			},
			expected: &Config{
				RedisURL:     "redis",
				Host:         "https://google.com",
				AuthToken:    "11881",
				RtbStatus:    "1",
				BuildFieldID: 1,
				DoneStatus:   "1222",
				Port:         "8080",
				SentryDSN:    "sentry",
			},
		},
		{
			name: "all required settings set with port",
			envs: map[string]string{
				"REDIS_URL":                   "redis",
				"REDMINE_HOST":                "https://google.com",
				"REDMINE_API_KEY":             "11881",
				"STAMP_READY_TO_BUILD_STATUS": "1",
				"STAMP_BUILD_CUSTOM_FIELD":    "1",
				"STAMP_DONE_STATUS":           "1222",
				"PORT":                        "8084",
				"SENTRY_DSN":                  "sentry",
			},
			expected: &Config{
				RedisURL:     "redis",
				Host:         "https://google.com",
				AuthToken:    "11881",
				RtbStatus:    "1",
				BuildFieldID: 1,
				DoneStatus:   "1222",
				Port:         "8084",
				SentryDSN:    "sentry",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
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
		})
	}
}
