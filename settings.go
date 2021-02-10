package main

// Settings combine all ENV setting needed for app working
type Settings struct {
	redisURL     string
	host         string
	authToken    string
	rtbStatus    string
	buildFieldID int64
	doneStatus   string
}
