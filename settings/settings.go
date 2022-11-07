package settings

// Settings combine all ENV setting needed for app working
type Settings struct {
	RedisURL     string
	Host         string
	AuthToken    string
	RtbStatus    string
	BuildFieldID int64
	DoneStatus   string
}
