package metigan

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// EmailService sends emails. Get it with Client.Email.
type EmailService struct {
	client *Client
}

// SendEmail queues an email to every recipient (POST /api/email/send).
func (s *EmailService) SendEmail(ctx context.Context, options EmailOptions) (*SendEmailResponse, error) {
	if err := required("from", options.From); err != nil {
		return nil, err
	}
	if len(options.Recipients) == 0 {
		return nil, &ValidationError{Field: "recipients", Message: "at least one recipient is required"}
	}
	if err := required("subject", options.Subject); err != nil {
		return nil, err
	}
	if err := required("content", options.Content); err != nil {
		return nil, err
	}
	return s.send(ctx, options, nil)
}

// SendEmailWithTemplate queues one of your templates to every recipient.
// Variables fill {{name}} placeholders; options.Subject and
// options.Content may be left empty (the template provides them).
func (s *EmailService) SendEmailWithTemplate(ctx context.Context, templateID string, variables map[string]any, options EmailOptions) (*SendEmailResponse, error) {
	if err := required("templateId", templateID); err != nil {
		return nil, err
	}
	if err := required("from", options.From); err != nil {
		return nil, err
	}
	if len(options.Recipients) == 0 {
		return nil, &ValidationError{Field: "recipients", Message: "at least one recipient is required"}
	}
	return s.send(ctx, options, &templateFields{UseTemplate: "true", TemplateID: templateID, Variables: variables})
}

type templateFields struct {
	UseTemplate string         `json:"useTemplate"`
	TemplateID  string         `json:"templateId"`
	Variables   map[string]any `json:"variables,omitempty"`
}

func (s *EmailService) send(ctx context.Context, options EmailOptions, tpl *templateFields) (*SendEmailResponse, error) {
	if options.Type == EmailTypeTransactional && len(options.Recipients) > MaxTransactionalRecipients {
		return nil, &ValidationError{Field: "recipients", Message: fmt.Sprintf(
			"a transactional send accepts at most %d recipients; use EmailTypeMarketing", MaxTransactionalRecipients)}
	}
	if options.IdempotencyKey == "" {
		options.IdempotencyKey = newIdempotencyKey()
	}
	var body any = options
	if tpl != nil {
		body = struct {
			EmailOptions
			templateFields
		}{options, *tpl}
	}
	var out SendEmailResponse
	if err := s.client.do(ctx, http.MethodPost, "/api/email/send", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SendOTP sends a one-time code (POST /api/otp/send).
func (s *EmailService) SendOTP(ctx context.Context, options OTPOptions) (*QueuedEmail, error) {
	if err := required("to", options.To); err != nil {
		return nil, err
	}
	if err := required("from", options.From); err != nil {
		return nil, err
	}
	if err := required("code", options.Code); err != nil {
		return nil, err
	}
	if options.IdempotencyKey == "" {
		options.IdempotencyKey = newIdempotencyKey()
	}
	var out QueuedEmail
	if err := s.client.do(ctx, http.MethodPost, "/api/otp/send", nil, options, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SendTransactional sends a single transactional message
// (POST /api/transactional/send).
func (s *EmailService) SendTransactional(ctx context.Context, options TransactionalOptions) (*QueuedEmail, error) {
	if err := required("to", options.To); err != nil {
		return nil, err
	}
	if err := required("from", options.From); err != nil {
		return nil, err
	}
	if options.TemplateID == "" {
		if err := required("subject", options.Subject); err != nil {
			return nil, err
		}
		if err := required("html", options.HTML); err != nil {
			return nil, err
		}
	}
	if options.IdempotencyKey == "" {
		options.IdempotencyKey = newIdempotencyKey()
	}
	var out QueuedEmail
	if err := s.client.do(ctx, http.MethodPost, "/api/transactional/send", nil, options, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns the delivery status of an email, by the EmailID a send
// returned (GET /api/email/{emailId}).
func (s *EmailService) Get(ctx context.Context, emailID string) (*EmailStatus, error) {
	if err := required("emailId", emailID); err != nil {
		return nil, err
	}
	var out struct {
		Data EmailStatus `json:"data"`
	}
	if err := s.client.do(ctx, http.MethodGet, "/api/email/"+url.PathEscape(emailID), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}
