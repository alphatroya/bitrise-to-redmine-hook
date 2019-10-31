package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/bitrise", handler)
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("Bitrise-Event-Type") != "build/finished" {
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errJSON := new(HookErrorResponse)
		errJSON.Message = "Received wrong request data payload"
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errJSON)
		return
	}
	payload := new(HookPayload)
	err = json.Unmarshal(data, payload)
	if err != nil {
		errJSON := new(HookErrorResponse)
		errJSON.Message = "Can't decode request payload json data"
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errJSON)
		return
	}

	if payload.BuildStatus != 0 {
		return
	}

	settings, errorResponse := NewSettings()
	if errorResponse != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	redmineProject := r.Header.Get("REDMINE_PROJECT")
	if len(redmineProject) == 0 {
		errJSON := new(HookErrorResponse)
		errJSON.Message = "REDMINE_PROJECT header is not set"
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errJSON)
		return
	}

	issues, err := issues(settings, redmineProject)
	if err != nil {
		errJSON := new(HookErrorResponse)
		errJSON.Message = fmt.Sprintf("Wrong error from server: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errJSON)
		return
	}

	for _, issue := range issues.Issues {
		err = markAsDone(issue, settings, payload.BuildNumber)
		if err != nil {
			errJSON := new(HookErrorResponse)
			errJSON.Message = fmt.Sprintf("Wrong error from server: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errJSON)
		}
	}
}

func markAsDone(issue *Issue, settings *Settings, buildNumber int) error {
	type PayloadCustomField struct {
		ID    int64  `json:"id"`
		Value string `json:"value"`
	}

	type PayloadIssue struct {
		AssignedToId string                `json:"assigned_to_id"`
		StatusId     string                `json:"status_id"`
		CustomFields []*PayloadCustomField `json:"custom_fields"`
	}

	type Payload struct {
		Issue *PayloadIssue `json:"issue"`
	}

	requestBody := Payload{
		Issue: &PayloadIssue{
			AssignedToId: fmt.Sprintf("%d", issue.Author.ID),
			StatusId:     settings.doneStatus,
			CustomFields: []*PayloadCustomField{
				{settings.buildFieldID, fmt.Sprintf("%d", buildNumber)},
			},
		},
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	buffer := bytes.NewBuffer(body)

	request, err := http.NewRequest("PUT", settings.host+fmt.Sprintf("/issues/%d.json", issue.ID), buffer)
	if err != nil {
		return err
	}
	request.Header.Set("X-Redmine-API-Key", settings.authToken)
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode >= 400 {
		return fmt.Errorf("Received wrong status code %d", response.StatusCode)
	}
	return nil
}

func issues(settings *Settings, project string) (*IssuesList, error) {
	request, err := http.NewRequest("GET", settings.host+"/issues.json?status_id="+settings.rtbStatus+"&project_id="+project, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("X-Redmine-API-Key", settings.authToken)
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode >= 400 {
		return nil, fmt.Errorf("Received wrong status code %d", response.StatusCode)
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	result := new(IssuesList)
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type HookErrorResponse struct {
	Message string `json:"message"`
}

func (h *HookErrorResponse) Error() string {
	return h.Message
}

type HookPayload struct {
	BuildSlug              string `json:"build_slug"`
	BuildNumber            int    `json:"build_number"`
	AppSlug                string `json:"app_slug"`
	BuildStatus            int    `json:"build_status"`
	BuildTriggeredWorkflow string `json:"build_triggered_workflow"`
	Git                    struct {
		Provider      string      `json:"provider"`
		SrcBranch     string      `json:"src_branch"`
		DstBranch     string      `json:"dst_branch"`
		PullRequestID int         `json:"pull_request_id"`
		Tag           interface{} `json:"tag"`
	} `json:"git"`
}

type IssuesList struct {
	Issues     []*Issue `json:"issues"`
	TotalCount int      `json:"total_count"`
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
}

type IssueContainer struct {
	Issue *Issue `json:"issue"`
}

type Issue struct {
	AssignedTo struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"assigned_to"`
	Author struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"author"`
	CreatedOn   string `json:"created_on"`
	Description string `json:"description"`
	DoneRatio   int    `json:"done_ratio"`
	DueDate     string `json:"due_date"`
	ID          int    `json:"id"`
	Priority    struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"priority"`
	Project struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"project"`
	SpentHours float32 `json:"spent_hours"`
	Status     struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"status"`
	Subject string `json:"subject"`
	Tracker struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"tracker"`
	UpdatedOn string `json:"updated_on"`
}
