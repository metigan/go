package metigan

import (
	"fmt"
)

// TemplatesService handles template operations
type TemplatesService struct {
	client *Client
}

// Get retrieves a template by ID
func (s *TemplatesService) Get(templateID string) (*EmailTemplate, error) {
	if templateID == "" {
		return nil, &ValidationError{Message: "template ID is required", Field: "templateId"}
	}

	resp, err := s.client.makeRequest("GET", fmt.Sprintf("/templates/%s", templateID), nil)
	if err != nil {
		return nil, err
	}

	var template EmailTemplate
	if err := s.client.parseResponse(resp, &template); err != nil {
		return nil, err
	}

	return &template, nil
}

// List lists all templates
func (s *TemplatesService) List(options PaginationOptions) ([]EmailTemplate, error) {
	query := "?"
	if options.Page > 0 {
		query += fmt.Sprintf("page=%d&", options.Page)
	}
	if options.Limit > 0 {
		query += fmt.Sprintf("limit=%d&", options.Limit)
	}

	endpoint := "/templates"
	if query != "?" {
		endpoint += query[:len(query)-1]
	}

	resp, err := s.client.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Templates []EmailTemplate `json:"templates"`
	}
	if err := s.client.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Templates, nil
}

