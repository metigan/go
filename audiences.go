package metigan

import (
	"fmt"
)

// AudiencesService handles audience operations
type AudiencesService struct {
	client *Client
}

// Create creates a new audience
func (s *AudiencesService) Create(options CreateAudienceOptions) (*Audience, error) {
	if options.Name == "" {
		return nil, &ValidationError{Message: "name is required", Field: "name"}
	}

	resp, err := s.client.makeRequest("POST", "/audiences", options)
	if err != nil {
		return nil, err
	}

	var audience Audience
	if err := s.client.parseResponse(resp, &audience); err != nil {
		return nil, err
	}

	return &audience, nil
}

// Get retrieves an audience by ID
func (s *AudiencesService) Get(audienceID string) (*Audience, error) {
	if audienceID == "" {
		return nil, &ValidationError{Message: "audience ID is required", Field: "audienceId"}
	}

	resp, err := s.client.makeRequest("GET", fmt.Sprintf("/audiences/%s", audienceID), nil)
	if err != nil {
		return nil, err
	}

	var audience Audience
	if err := s.client.parseResponse(resp, &audience); err != nil {
		return nil, err
	}

	return &audience, nil
}

// List lists audiences with optional pagination
func (s *AudiencesService) List(options PaginationOptions) (*AudienceListResponse, error) {
	query := "?"
	if options.Page > 0 {
		query += fmt.Sprintf("page=%d&", options.Page)
	}
	if options.Limit > 0 {
		query += fmt.Sprintf("limit=%d&", options.Limit)
	}

	endpoint := "/audiences"
	if query != "?" {
		endpoint += query[:len(query)-1]
	}

	resp, err := s.client.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result AudienceListResponse
	if err := s.client.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetStats retrieves statistics for an audience
func (s *AudiencesService) GetStats(audienceID string) (*AudienceStats, error) {
	if audienceID == "" {
		return nil, &ValidationError{Message: "audience ID is required", Field: "audienceId"}
	}

	resp, err := s.client.makeRequest("GET", fmt.Sprintf("/audiences/%s/stats", audienceID), nil)
	if err != nil {
		return nil, err
	}

	var stats AudienceStats
	if err := s.client.parseResponse(resp, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

// Delete deletes an audience
func (s *AudiencesService) Delete(audienceID string) error {
	if audienceID == "" {
		return &ValidationError{Message: "audience ID is required", Field: "audienceId"}
	}

	resp, err := s.client.makeRequest("DELETE", fmt.Sprintf("/audiences/%s", audienceID), nil)
	if err != nil {
		return err
	}

	return s.client.parseResponse(resp, nil)
}

