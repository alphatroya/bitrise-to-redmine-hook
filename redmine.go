package main

// IssuesContainer represents get issues response data
type IssuesContainer struct {
	Issues []*Issue `json:"issues"`
}

// Issue represents single Redmine issue
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
