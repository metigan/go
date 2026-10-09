// Package metigan is the official Go client for the Metigan API
// (https://api.metigan.io): send emails, OTP codes and transactional
// messages, and manage contacts, audiences, templates and forms.
//
//	client := metigan.NewClient(metigan.Config{APIKey: os.Getenv("METIGAN_API_KEY")})
//	res, err := client.Email().SendEmail(ctx, metigan.EmailOptions{
//		From:       "My App <onboarding@metigan.io>",
//		Recipients: []string{"user@example.com"},
//		Subject:    "Welcome!",
//		Content:    "<h1>Hello!</h1>",
//	})
package metigan

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Version is the SDK version, sent in the User-Agent header.
const Version = "0.1.0"

// DefaultBaseURL is the Metigan API used when neither Config.BaseURL nor the
// METIGAN_API_URL environment variable is set.
const DefaultBaseURL = "https://api.metigan.io"

// Config configures a Client. Only APIKey is required.
type Config struct {
	// APIKey is your Metigan API key, sent as the X-Api-Key header.
	APIKey string
	// BaseURL overrides the API address. Defaults to METIGAN_API_URL, then
	// DefaultBaseURL.
	BaseURL string
	// Timeout is the per-attempt timeout (default 30s). Ignored when
	// HTTPClient is set.
	Timeout time.Duration
	// RetryCount is how many times a request is retried after a network
	// error or a 5xx response (default 3). A negative value disables retries.
	RetryCount int
	// RetryDelay is the base delay between retries, multiplied by the
	// attempt number (default 1s).
	RetryDelay time.Duration
	// HTTPClient replaces the default http.Client.
	HTTPClient *http.Client
	// WebhookSecret is the signing secret (whsec_...) used by
	// Client.Webhooks().Verify.
	WebhookSecret string
}

// Client is the Metigan API client. It is safe for concurrent use.
type Client struct {
	apiKey     string
	baseURL    string
	retryCount int
	retryDelay time.Duration
	httpClient *http.Client

	email     *EmailService
	contacts  *ContactsService
	audiences *AudiencesService
	templates *TemplatesService
	forms     *FormsService
	webhooks  *WebhooksService
}

// NewClient returns a client configured by config.
func NewClient(config Config) *Client {
	c := &Client{
		apiKey:     config.APIKey,
		baseURL:    config.BaseURL,
		retryCount: config.RetryCount,
		retryDelay: config.RetryDelay,
		httpClient: config.HTTPClient,
	}
	if c.baseURL == "" {
		c.baseURL = os.Getenv("METIGAN_API_URL")
	}
	if c.baseURL == "" {
		c.baseURL = DefaultBaseURL
	}
	c.baseURL = strings.TrimRight(c.baseURL, "/")
	if c.retryCount == 0 {
		c.retryCount = 3
	} else if c.retryCount < 0 {
		c.retryCount = 0
	}
	if c.retryDelay == 0 {
		c.retryDelay = time.Second
	}
	if c.httpClient == nil {
		timeout := config.Timeout
		if timeout == 0 {
			timeout = 30 * time.Second
		}
		c.httpClient = &http.Client{Timeout: timeout}
	}

	c.email = &EmailService{client: c}
	c.contacts = &ContactsService{client: c}
	c.audiences = &AudiencesService{client: c}
	c.templates = &TemplatesService{client: c}
	c.forms = &FormsService{client: c}
	c.webhooks = &WebhooksService{secret: config.WebhookSecret, tolerance: DefaultWebhookTolerance}
	return c
}

// Email sends emails, OTP codes and transactional messages.
func (c *Client) Email() *EmailService { return c.email }

// Contacts manages contacts.
func (c *Client) Contacts() *ContactsService { return c.contacts }

// Audiences manages audiences.
func (c *Client) Audiences() *AudiencesService { return c.audiences }

// Templates reads email templates.
func (c *Client) Templates() *TemplatesService { return c.templates }

// Forms reads forms and submits responses.
func (c *Client) Forms() *FormsService { return c.forms }

// Webhooks verifies webhook signatures.
func (c *Client) Webhooks() *WebhooksService { return c.webhooks }

// do sends a JSON request and decodes the response into out (if non-nil).
// The body is rebuilt on every attempt, so retries send the full payload.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	if c.apiKey == "" {
		return &ValidationError{Field: "apiKey", Message: "API key is required"}
	}
	var payload []byte
	if body != nil {
		var err error
		if payload, err = json.Marshal(body); err != nil {
			return fmt.Errorf("metigan: encode request: %w", err)
		}
	}
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var lastErr error
	for attempt := 0; attempt <= c.retryCount; attempt++ {
		if attempt > 0 {
			t := time.NewTimer(c.retryDelay * time.Duration(attempt))
			select {
			case <-ctx.Done():
				t.Stop()
				return ctx.Err()
			case <-t.C:
			}
		}
		var r io.Reader
		if payload != nil {
			r = bytes.NewReader(payload)
		}
		req, err := http.NewRequestWithContext(ctx, method, u, r)
		if err != nil {
			return fmt.Errorf("metigan: build request: %w", err)
		}
		req.Header.Set("X-Api-Key", c.apiKey)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "Metigan-Go-SDK/"+Version)
		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			lastErr = err
			continue
		}
		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode >= 500 {
			lastErr = newAPIError(resp.StatusCode, data)
			continue
		}
		if resp.StatusCode >= 400 {
			return newAPIError(resp.StatusCode, data)
		}
		if out != nil && len(bytes.TrimSpace(data)) > 0 {
			if err := json.Unmarshal(data, out); err != nil {
				return fmt.Errorf("metigan: decode response: %w", err)
			}
		}
		return nil
	}
	var apiErr *APIError
	if errors.As(lastErr, &apiErr) {
		return apiErr
	}
	return fmt.Errorf("metigan: request failed after %d attempts: %w", c.retryCount+1, lastErr)
}

// newIdempotencyKey returns a random key, so retries of one call are
// deduplicated by the API instead of sending twice.
func newIdempotencyKey() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return "go-sdk-" + hex.EncodeToString(b[:])
}

// APIError is a non-2xx response from the API.
type APIError struct {
	StatusCode int
	// Err is the short error ("error" field), e.g. "Missing required fields".
	Err string `json:"error"`
	// Message is the detail ("message" field).
	Message string `json:"message"`
	// Code is a machine-readable code, when the API sends one.
	Code string `json:"code"`
	// Body is the raw response body.
	Body []byte `json:"-"`
}

func newAPIError(status int, body []byte) *APIError {
	e := &APIError{StatusCode: status, Body: body}
	_ = json.Unmarshal(body, e)
	return e
}

func (e *APIError) Error() string {
	switch {
	case e.Err != "" && e.Message != "" && e.Err != e.Message:
		return fmt.Sprintf("metigan: API error %d: %s: %s", e.StatusCode, e.Err, e.Message)
	case e.Message != "":
		return fmt.Sprintf("metigan: API error %d: %s", e.StatusCode, e.Message)
	case e.Err != "":
		return fmt.Sprintf("metigan: API error %d: %s", e.StatusCode, e.Err)
	}
	return fmt.Sprintf("metigan: API error %d", e.StatusCode)
}

// ValidationError is a request the SDK rejects before calling the API.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("metigan: invalid %s: %s", e.Field, e.Message)
	}
	return "metigan: " + e.Message
}

func required(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{Field: field, Message: "is required"}
	}
	return nil
}

func pageQuery(p PaginationOptions) url.Values {
	q := url.Values{}
	if p.Page > 0 {
		q.Set("page", fmt.Sprint(p.Page))
	}
	if p.Limit > 0 {
		q.Set("limit", fmt.Sprint(p.Limit))
	}
	return q
}
