package metigan

import (
	"encoding/base64"
	"fmt"
)

// EmailService handles email operations
type EmailService struct {
	client *Client
}

// SendEmail sends an email
func (s *EmailService) SendEmail(options EmailOptions) (*EmailSuccessResponse, error) {
	// Validate required fields
	if options.From == "" {
		return nil, &ValidationError{Message: "from field is required", Field: "from"}
	}
	if len(options.Recipients) == 0 {
		return nil, &ValidationError{Message: "at least one recipient is required", Field: "recipients"}
	}
	if options.Subject == "" {
		return nil, &ValidationError{Message: "subject is required", Field: "subject"}
	}
	if options.Content == "" {
		return nil, &ValidationError{Message: "content is required", Field: "content"}
	}

	// Process attachments - encode to base64
	processedAttachments := make([]map[string]interface{}, len(options.Attachments))
	for i, att := range options.Attachments {
		processedAttachments[i] = map[string]interface{}{
			"content":     base64.StdEncoding.EncodeToString(att.Content),
			"filename":    att.Filename,
			"contentType": att.ContentType,
		}
	}

	// Prepare request body
	requestBody := map[string]interface{}{
		"from":       options.From,
		"recipients": options.Recipients,
		"subject":    options.Subject,
		"content":    options.Content,
	}

	if len(processedAttachments) > 0 {
		requestBody["attachments"] = processedAttachments
	}
	if len(options.CC) > 0 {
		requestBody["cc"] = options.CC
	}
	if len(options.BCC) > 0 {
		requestBody["bcc"] = options.BCC
	}
	if options.ReplyTo != "" {
		requestBody["replyTo"] = options.ReplyTo
	}
	if options.TrackingID != "" {
		requestBody["trackingId"] = options.TrackingID
	}

	// Make request
	resp, err := s.client.makeRequest("POST", "/emails", requestBody)
	if err != nil {
		return nil, err
	}

	var result EmailSuccessResponse
	if err := s.client.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SendEmailWithTemplate sends an email using a template
func (s *EmailService) SendEmailWithTemplate(templateID string, variables map[string]interface{}, options EmailOptions) (*EmailSuccessResponse, error) {
	if templateID == "" {
		return nil, &ValidationError{Message: "template ID is required", Field: "templateId"}
	}

	requestBody := map[string]interface{}{
		"templateId": templateID,
		"variables":  variables,
		"from":       options.From,
		"recipients": options.Recipients,
	}

	if options.Subject != "" {
		requestBody["subject"] = options.Subject
	}
	if options.ReplyTo != "" {
		requestBody["replyTo"] = options.ReplyTo
	}

	resp, err := s.client.makeRequest("POST", fmt.Sprintf("/emails/templates/%s", templateID), requestBody)
	if err != nil {
		return nil, err
	}

	var result EmailSuccessResponse
	if err := s.client.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

