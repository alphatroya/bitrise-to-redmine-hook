package main

type IssuesList struct {
	Issues []*Issue `json:"issues"`
}

type Issue struct {
	Author struct {
		ID int `json:"id"`
	} `json:"author"`
	ID      int `json:"id"`
	Project struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"project"`
}
