package metigan

import (
	"context"
	"net/http"
	"net/url"
)

// AudiencesService manages audiences. Get it with Client.Audiences.
type AudiencesService struct {
	client *Client
}

func audiencePath(id, suffix string) string {
	return "/api/audiences/" + url.PathEscape(id) + suffix
}

// Create creates an audience.
func (s *AudiencesService) Create(ctx context.Context, options CreateAudienceOptions) (*Audience, error) {
	if err := required("name", options.Name); err != nil {
		return nil, err
	}
	var out Audience
	if err := s.client.do(ctx, http.MethodPost, "/api/audiences", nil, options, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns an audience with its most recent contacts.
func (s *AudiencesService) Get(ctx context.Context, audienceID string) (*AudienceWithContacts, error) {
	if err := required("audienceId", audienceID); err != nil {
		return nil, err
	}
	var out AudienceWithContacts
	if err := s.client.do(ctx, http.MethodGet, audiencePath(audienceID, ""), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns a page of audiences, newest first.
func (s *AudiencesService) List(ctx context.Context, options PaginationOptions) (*AudienceListResponse, error) {
	var out AudienceListResponse
	if err := s.client.do(ctx, http.MethodGet, "/api/audiences", pageQuery(options), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update changes an audience; nil fields are left unchanged.
func (s *AudiencesService) Update(ctx context.Context, audienceID string, options UpdateAudienceOptions) (*Audience, error) {
	if err := required("audienceId", audienceID); err != nil {
		return nil, err
	}
	var out Audience
	if err := s.client.do(ctx, http.MethodPatch, audiencePath(audienceID, ""), nil, options, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes an audience.
func (s *AudiencesService) Delete(ctx context.Context, audienceID string) error {
	if err := required("audienceId", audienceID); err != nil {
		return err
	}
	return s.client.do(ctx, http.MethodDelete, audiencePath(audienceID, ""), nil, nil, nil)
}

// GetStats counts an audience's contacts by status.
func (s *AudiencesService) GetStats(ctx context.Context, audienceID string) (*AudienceStats, error) {
	if err := required("audienceId", audienceID); err != nil {
		return nil, err
	}
	var out AudienceStats
	if err := s.client.do(ctx, http.MethodGet, audiencePath(audienceID, "/stats"), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
