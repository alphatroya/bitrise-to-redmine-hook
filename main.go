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
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	host, err := getEnvVar(w, "REDMINE_HOST")
	if err != nil {
		return
	}

	authToken, err := getEnvVar(w, "REDMINE_API_KEY")
	if err != nil {
		return
	}

	rtbStatus, err := getEnvVar(w, "STAMP_READY_TO_BUILD_STATUS")
	if err != nil {
		return
	}

	// nextStatus, err := getEnvVar(w, "STAMP_DONE_STATUS")
	// if err != nil {
	// 	return
	// }

	redmineProject := r.Header.Get("REDMINE_PROJECT")
	if len(redmineProject) == 0 {
		errJSON := new(HookErrorResponse)
		errJSON.Message = "REDMINE_PROJECT header is not set"
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errJSON)
		return
	}

	issues, err := issues(host, rtbStatus, redmineProject, authToken)
	if err != nil {
		errJSON := new(HookErrorResponse)
		errJSON.Message = fmt.Sprintf("Wrong error from server: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errJSON)
		return
	}
	var issuesID []int
	for _, issue := range issues.Issues {
		issuesID = append(issuesID, issue.ID)
	}
	fmt.Fprintf(w, "%#v", issuesID)
}

func getEnvVar(w http.ResponseWriter, key string) (string, error) {
	authToken := os.Getenv(key)
	if len(authToken) == 0 {
		errJSON := new(HookErrorResponse)
		errJSON.Message = key + " ENV variable is not set"
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errJSON)
		return "", errJSON
	}
	return authToken, nil
}

func markAsDone(issue *Issue, host string, nextStatus string, token string) (*Issue, error) {

	type Payload struct {
		Issue struct {
			assignedToId string `json:"assigned_to_id"`
			statusId     string `json:"status_id"`
		} `json:"issue"`
	}

	requestBody := Payload{}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(body)
	// assignee=$(http GET "$REDMINE_HOST/issues/$issue_id.json" "X-Redmine-API-Key":"$REDMINE_API_KEY" | jq '.issue.author.id')
	// echo "{ \"issue\": { \"assigned_to_id\": $assignee, \"status_id\": $next_status_id, \"custom_fields\": [{\"id\": 32, \"value\": \"$build_number\" }] }}" |
	//     http PUT "$REDMINE_HOST/issues/$issue_id.json" "X-Redmine-API-Key":"$REDMINE_API_KEY" >/dev/null

	request, err := http.NewRequest("PUT", host+fmt.Sprintf("/issues/%d.json", id), buffer)
	if err != nil {
		return nil, err
	}
	request.Header.Set("X-Redmine-API-Key", token)
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
	result := new(Issue)
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func issues(host string, status string, project string, token string) (*IssuesList, error) {
	request, err := http.NewRequest("GET", host+"/issues.json?status_id="+status+"&project_id="+project, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("X-Redmine-API-Key", token)
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
