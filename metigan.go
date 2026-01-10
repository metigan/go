package metigan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// baseURL is the base URL for the Metigan API
	baseURL = "https://api.metigan.com"
	// apiVersion is the API version path
	apiVersion = "/v1"
)

// Config holds the configuration for the Metigan client
type Config struct {
	APIKey     string
	Timeout    time.Duration
	RetryCount int
	RetryDelay time.Duration
	Debug      bool
}

// Client is the main Metigan client
type Client struct {
	config     Config
	httpClient *http.Client
	email      *EmailService
	contacts   *ContactsService
	audiences  *AudiencesService
	templates  *TemplatesService
	forms      *FormsService
}

// NewClient creates a new Metigan client with the provided configuration
func NewClient(config Config) *Client {
	// Set defaults
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.RetryCount == 0 {
		config.RetryCount = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 2 * time.Second
	}

	client := &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}

	// Initialize services
	client.email = &EmailService{client: client}
	client.contacts = &ContactsService{client: client}
	client.audiences = &AudiencesService{client: client}
	client.templates = &TemplatesService{client: client}
	client.forms = &FormsService{client: client}

	return client
}

// Email returns the email service
func (c *Client) Email() *EmailService {
	return c.email
}

// Contacts returns the contacts service
func (c *Client) Contacts() *ContactsService {
	return c.contacts
}

// Audiences returns the audiences service
func (c *Client) Audiences() *AudiencesService {
	return c.audiences
}

// Templates returns the templates service
func (c *Client) Templates() *TemplatesService {
	return c.templates
}

// Forms returns the forms service
func (c *Client) Forms() *FormsService {
	return c.forms
}


// makeRequest performs an HTTP request with retries
func (c *Client) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	url := baseURL + apiVersion + endpoint
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("User-Agent", "Metigan-Go-SDK/1.0")

	// Perform request with retries
	var resp *http.Response
	for attempt := 0; attempt <= c.config.RetryCount; attempt++ {
		resp, err = c.httpClient.Do(req)
		if err == nil {
			if resp.StatusCode < 500 {
				return resp, nil
			}
			// Close response body if status is 5xx and we'll retry
			if attempt < c.config.RetryCount {
				resp.Body.Close()
			}
		}

		if attempt < c.config.RetryCount {
			time.Sleep(c.config.RetryDelay)
			continue
		}
	}

	if err != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", c.config.RetryCount, err)
	}

	return resp, nil
}

// parseResponse parses the HTTP response into the target type
func (c *Client) parseResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var apiError APIError
		if err := json.NewDecoder(resp.Body).Decode(&apiError); err != nil {
			return &APIError{
				StatusCode: resp.StatusCode,
				Message:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status),
			}
		}
		apiError.StatusCode = resp.StatusCode
		return &apiError
	}

	if target != nil {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// APIError represents an API error response
type APIError struct {
	StatusCode int    `json:"-"`
	Err        string `json:"error,omitempty"`
	Message    string `json:"message,omitempty"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
	}
	if e.Err != "" {
		return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Err)
	}
	return fmt.Sprintf("API error %d", e.StatusCode)
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
	Field   string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error in field %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

