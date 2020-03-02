package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

func sendMailgunNotification(response *HookResponse, redmineHost string) error {
	token := os.Getenv("MAILGUN_API")
	domain := os.Getenv("MAILGUN_DOMAIN")
	rec := os.Getenv("MAILGUN_RECIPIENT")
	sender := os.Getenv("MAILGUN_SENDER")
	if token == "" || domain == "" || rec == "" || sender == "" {
		return errors.New("one of parameters for MAILGUN integration is not set")
	}

	mg := mailgun.NewMailgun(domain, token)

	subject := "Redmine Hooks Results"
	var body string
	body += "Success:\n"
	for _, success := range response.Success {
		body += redmineHost + "/issues/" + fmt.Sprintf("%d", success) + "\n"
	}
	body += "Failures:\n"
	for _, failure := range response.Failures {
		body += redmineHost + "/issues/" + fmt.Sprintf("%d", failure) + "\n"
	}

	message := mg.NewMessage(sender, subject, body, rec)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, _, err := mg.Send(ctx, message)

	if err != nil {
		return err
	}

	return nil
}
