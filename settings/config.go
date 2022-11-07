package settings

// Config combine all ENV setting needed for app working
type Config struct {
	RedisURL     string `env:"REDIS_URL"                   env-required:"true"`
	Host         string `env:"REDMINE_HOST"                env-required:"true"`
	AuthToken    string `env:"REDMINE_API_KEY"             env-required:"true"`
	RtbStatus    string `env:"STAMP_READY_TO_BUILD_STATUS" env-required:"true"`
	BuildFieldID int64  `env:"STAMP_BUILD_CUSTOM_FIELD"    env-required:"true"`
	DoneStatus   string `env:"STAMP_DONE_STATUS"           env-required:"true"`
}
