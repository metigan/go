package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	metigan "github.com/metigan/go"
)

func main() {
	apiKey := os.Getenv("METIGAN_API_KEY")
	if apiKey == "" {
		log.Fatal("METIGAN_API_KEY environment variable is required")
	}
	to := os.Getenv("METIGAN_TO")
	if to == "" {
		log.Fatal("METIGAN_TO (the recipient) is required")
	}

	client := metigan.NewClient(metigan.Config{APIKey: apiKey})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// onboarding@metigan.io works on every account without verifying a domain.
	res, err := client.Email().SendEmail(ctx, metigan.EmailOptions{
		From:       "Metigan Go SDK <onboarding@metigan.io>",
		Recipients: []string{to},
		Subject:    "Hello from Go!",
		Content:    "<h1>Hello!</h1><p>This email was sent from a Go application.</p>",
	})
	var apiErr *metigan.APIError
	if errors.As(err, &apiErr) {
		log.Fatalf("API refused the email (%d): %s", apiErr.StatusCode, apiErr.Message)
	}
	if err != nil {
		log.Fatalf("send failed: %v", err)
	}

	fmt.Printf("%s (%d recipient(s), %d emails left this month)\n", res.Message, res.RecipientCount, res.EmailsRemaining)
	for _, e := range res.SuccessfulEmails {
		status, err := client.Email().Get(ctx, e.EmailID)
		if err != nil {
			log.Fatalf("status of %s: %v", e.EmailID, err)
		}
		fmt.Printf("%s -> %s (%s)\n", e.Recipient, status.Status, e.EmailID)
	}
	for _, f := range res.FailedEmails {
		fmt.Printf("%s failed: %s\n", f.Recipient, f.Error)
	}
}
