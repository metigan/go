package metigan

import (
	"context"
	"net/http"
	"net/url"
)

// FormsService reads forms and submits responses. Get it with
// Client.Forms.
type FormsService struct {
	client *Client
}

// Submit stores a response to a form (POST /api/submissions).
func (s *FormsService) Submit(ctx context.Context, options FormSubmissionOptions) (*FormSubmissionResponse, error) {
	if err := required("formId", options.FormID); err != nil {
		return nil, err
	}
	if options.Data == nil {
		return nil, &ValidationError{Field: "data", Message: "is required"}
	}
	var out FormSubmissionResponse
	if err := s.client.do(ctx, http.MethodPost, "/api/submissions", nil, options, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns a form by ID ("form-<id>" or the bare id) or slug.
func (s *FormsService) Get(ctx context.Context, formIDOrSlug string) (*Form, error) {
	if err := required("formIdOrSlug", formIDOrSlug); err != nil {
		return nil, err
	}
	var out Form
	if err := s.client.do(ctx, http.MethodGet, "/api/forms/"+url.PathEscape(formIDOrSlug), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns a page of forms.
func (s *FormsService) List(ctx context.Context, options PaginationOptions) (*FormListResponse, error) {
	var out FormListResponse
	if err := s.client.do(ctx, http.MethodGet, "/api/forms", pageQuery(options), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListSubmissions returns a page of a form's responses.
func (s *FormsService) ListSubmissions(ctx context.Context, formIDOrSlug string, options PaginationOptions) (*FormSubmissionListResponse, error) {
	if err := required("formIdOrSlug", formIDOrSlug); err != nil {
		return nil, err
	}
	q := pageQuery(options)
	q.Set("formId", formIDOrSlug)
	var out FormSubmissionListResponse
	if err := s.client.do(ctx, http.MethodGet, "/api/submissions", q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
