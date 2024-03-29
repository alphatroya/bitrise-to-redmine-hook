package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/alphatroya/ci-redmine-bindings/settings"
)

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

func issues(settings *settings.Config, project string) (*IssuesContainer, error) {
	request, err := http.NewRequest("GET", settings.Host+"/issues.json?status_id="+settings.RtbStatus+"&project_id="+project, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("X-Redmine-API-Key", settings.AuthToken)
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("Received wrong status code %d", response.StatusCode)
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	result := new(IssuesContainer)
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
