package metigan

import (
	"fmt"
)

// FormsService handles form operations
type FormsService struct {
	client *Client
}

// Submit submits a form
func (s *FormsService) Submit(options FormSubmissionOptions) (*FormSubmissionResponse, error) {
	if options.FormID == "" {
		return nil, &ValidationError{Message: "form ID is required", Field: "formId"}
	}
	if options.Data == nil {
		return nil, &ValidationError{Message: "form data is required", Field: "data"}
	}

	resp, err := s.client.makeRequest("POST", fmt.Sprintf("/forms/%s/submit", options.FormID), options)
	if err != nil {
		return nil, err
	}

	var result FormSubmissionResponse
	if err := s.client.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Get retrieves a form by ID or slug
func (s *FormsService) Get(formIDOrSlug string) (*Form, error) {
	if formIDOrSlug == "" {
		return nil, &ValidationError{Message: "form ID or slug is required", Field: "formIdOrSlug"}
	}

	resp, err := s.client.makeRequest("GET", fmt.Sprintf("/forms/%s", formIDOrSlug), nil)
	if err != nil {
		return nil, err
	}

	var form Form
	if err := s.client.parseResponse(resp, &form); err != nil {
		return nil, err
	}

	return &form, nil
}

// List lists all forms
func (s *FormsService) List(options PaginationOptions) (*FormListResponse, error) {
	query := "?"
	if options.Page > 0 {
		query += fmt.Sprintf("page=%d&", options.Page)
	}
	if options.Limit > 0 {
		query += fmt.Sprintf("limit=%d&", options.Limit)
	}

	endpoint := "/forms"
	if query != "?" {
		endpoint += query[:len(query)-1]
	}

	resp, err := s.client.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result FormListResponse
	if err := s.client.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

