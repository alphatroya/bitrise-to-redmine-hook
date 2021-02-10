package main

func batchTransaction(rm DoneMarker, issues *IssuesContainer, settings *Settings, buildNumber int) *HookResponse {
	type Result struct {
		id  int
		err error
	}
	ch := make(chan Result)
	for _, issue := range issues.Issues {
		go func(issue *Issue) {
			err := rm.markAsDone(issue, settings, buildNumber)
			ch <- Result{issue.ID, err}
		}(issue)
	}

	response := NewResponse("Successful completed task")
	for range issues.Issues {
		res := <-ch
		if res.err != nil {
			response.Failures = append(response.Failures, res.id)
			continue
		}
		response.Success = append(response.Success, res.id)
	}

	return response
}
