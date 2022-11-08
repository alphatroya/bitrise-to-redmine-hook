package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

func sendMailgunNotification(response *HookResponse, redmineHost string, buildNumber int, issues []*Issue, version string) error {
	if len(response.Success) == 0 && len(response.Failures) == 0 {
		return errors.New("response object not contain neither success or failures")
	}

	token := os.Getenv("MAILGUN_API")
	domain := os.Getenv("MAILGUN_DOMAIN")
	rec := os.Getenv("MAILGUN_RECIPIENT")
	sender := os.Getenv("MAILGUN_SENDER")
	if token == "" || domain == "" || rec == "" || sender == "" {
		return errors.New("one of parameters for MAILGUN integration is not set")
	}

	mg := mailgun.NewMailgun(domain, token)

	subject := "Redmine Hooks Results"
	if len(issues) != 0 {
		subject += fmt.Sprintf(" (%s)", issues[0].Project.Name)
	}
	body := fmt.Sprintf("Build number: %d\n", buildNumber)
	if len(response.Success) != 0 {
		body += fmt.Sprintf("Success(%d):\n", len(response.Success))
		for _, success := range response.Success {
			body += redmineHost + "/issues/" + fmt.Sprintf("%d", success) + "\n"
		}
	}
	if len(response.Failures) != 0 {
		body += "Failures:\n"
		for _, failure := range response.Failures {
			body += redmineHost + "/issues/" + fmt.Sprintf("%d", failure) + "\n"
		}
	}

	body += "\n"
	body += version + "\n"
	message := mg.NewMessage(sender, subject, body, rec)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if _, _, err := mg.Send(ctx, message); err != nil {
		return err
	}

	return nil
}
