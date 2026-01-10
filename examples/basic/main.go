package main

import (
	"fmt"
	"log"
	"os"
	
	"github.com/metigan/go"
)

func main() {
	// Initialize the client
	apiKey := os.Getenv("METIGAN_API_KEY")
	if apiKey == "" {
		log.Fatal("METIGAN_API_KEY environment variable is required")
	}
	
	client := metigan.NewClient(metigan.Config{
		APIKey: apiKey,
	})
	
	// Example: Send an email
	result, err := client.Email().SendEmail(metigan.EmailOptions{
		From:       "Sender <sender@example.com>",
		Recipients: []string{"recipient@example.com"},
		Subject:    "Hello from Go!",
		Content:    "<h1>Hello!</h1><p>This email was sent from a Go application.</p>",
	})
	
	if err != nil {
		log.Fatalf("Failed to send email: %v", err)
	}
	
	if result.Success {
		fmt.Println("Email sent successfully!")
		fmt.Printf("Message: %s\n", result.Message)
		fmt.Printf("Recipients: %d\n", result.RecipientCount)
		fmt.Printf("Emails remaining: %d\n", result.EmailsRemaining)
	} else {
		fmt.Printf("Failed to send email: %s\n", result.Message)
	}
}

