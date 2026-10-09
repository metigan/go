package metigan

import (
	"context"
	"net/http"
	"net/url"
)

// TemplatesService reads email templates. Get it with Client.Templates.
type TemplatesService struct {
	client *Client
}

// Get returns a template by ID.
func (s *TemplatesService) Get(ctx context.Context, templateID string) (*EmailTemplate, error) {
	if err := required("templateId", templateID); err != nil {
		return nil, err
	}
	var out EmailTemplate
	if err := s.client.do(ctx, http.MethodGet, "/api/templates/"+url.PathEscape(templateID), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns a page of templates (default 20 per page, at most 100).
func (s *TemplatesService) List(ctx context.Context, options PaginationOptions) (*TemplateListResponse, error) {
	var out TemplateListResponse
	if err := s.client.do(ctx, http.MethodGet, "/api/templates", pageQuery(options), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
