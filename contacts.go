package metigan

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// ContactsService manages contacts. Get it with Client.Contacts.
type ContactsService struct {
	client *Client
}

func contactPath(id string, suffix string) string {
	return "/api/contacts/" + url.PathEscape(id) + suffix
}

func (s *ContactsService) contact(ctx context.Context, method, path string, query url.Values, body any) (*Contact, error) {
	var c Contact
	if err := s.client.do(ctx, method, path, query, body, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// Create adds a contact to an audience. If the email is already in the
// audience, the existing contact is returned.
func (s *ContactsService) Create(ctx context.Context, options CreateContactOptions) (*Contact, error) {
	if err := required("email", options.Email); err != nil {
		return nil, err
	}
	if err := required("audienceId", options.AudienceID); err != nil {
		return nil, err
	}
	return s.contact(ctx, http.MethodPost, "/api/contacts", nil, options)
}

// Get returns a contact by ID.
func (s *ContactsService) Get(ctx context.Context, contactID string) (*Contact, error) {
	if err := required("contactId", contactID); err != nil {
		return nil, err
	}
	return s.contact(ctx, http.MethodGet, contactPath(contactID, ""), nil, nil)
}

// GetByEmail returns the contact with this email in an audience.
func (s *ContactsService) GetByEmail(ctx context.Context, email, audienceID string) (*Contact, error) {
	if err := required("email", email); err != nil {
		return nil, err
	}
	if err := required("audienceId", audienceID); err != nil {
		return nil, err
	}
	return s.contact(ctx, http.MethodGet, "/api/contacts/email/"+url.PathEscape(email),
		url.Values{"audienceId": {audienceID}}, nil)
}

// Update changes a contact; nil fields are left unchanged.
func (s *ContactsService) Update(ctx context.Context, contactID string, options UpdateContactOptions) (*Contact, error) {
	if err := required("contactId", contactID); err != nil {
		return nil, err
	}
	return s.contact(ctx, http.MethodPatch, contactPath(contactID, ""), nil, options)
}

// List returns a page of contacts matching filters.
func (s *ContactsService) List(ctx context.Context, filters ContactListFilters) (*ContactListResponse, error) {
	q := pageQuery(PaginationOptions{Page: filters.Page, Limit: filters.Limit})
	for k, v := range map[string]string{
		"audienceId": filters.AudienceID, "status": string(filters.Status), "tag": filters.Tag,
		"search": filters.Search, "email": filters.Email,
	} {
		if v != "" {
			q.Set(k, v)
		}
	}
	var out ContactListResponse
	if err := s.client.do(ctx, http.MethodGet, "/api/contacts", q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Search finds contacts whose email or name matches query, optionally in
// one audience.
func (s *ContactsService) Search(ctx context.Context, query, audienceID string) ([]Contact, error) {
	if err := required("query", query); err != nil {
		return nil, err
	}
	q := url.Values{"q": {query}}
	if audienceID != "" {
		q.Set("audienceId", audienceID)
	}
	var out struct {
		Contacts []Contact `json:"contacts"`
	}
	if err := s.client.do(ctx, http.MethodGet, "/api/contacts/search", q, nil, &out); err != nil {
		return nil, err
	}
	return out.Contacts, nil
}

// Delete removes a contact.
func (s *ContactsService) Delete(ctx context.Context, contactID string) error {
	if err := required("contactId", contactID); err != nil {
		return err
	}
	return s.client.do(ctx, http.MethodDelete, contactPath(contactID, ""), nil, nil, nil)
}

// Subscribe marks a contact as subscribed.
func (s *ContactsService) Subscribe(ctx context.Context, contactID string) (*Contact, error) {
	if err := required("contactId", contactID); err != nil {
		return nil, err
	}
	return s.contact(ctx, http.MethodPost, contactPath(contactID, "/subscribe"), nil, nil)
}

// Unsubscribe marks a contact as unsubscribed.
func (s *ContactsService) Unsubscribe(ctx context.Context, contactID string) (*Contact, error) {
	if err := required("contactId", contactID); err != nil {
		return nil, err
	}
	return s.contact(ctx, http.MethodPost, contactPath(contactID, "/unsubscribe"), nil, nil)
}

// AddTags adds tags to a contact.
func (s *ContactsService) AddTags(ctx context.Context, contactID string, tags []string) (*Contact, error) {
	if err := checkTags(contactID, tags); err != nil {
		return nil, err
	}
	return s.contact(ctx, http.MethodPost, contactPath(contactID, "/tags"), nil, map[string][]string{"tags": tags})
}

// RemoveTags removes tags from a contact.
func (s *ContactsService) RemoveTags(ctx context.Context, contactID string, tags []string) (*Contact, error) {
	if err := checkTags(contactID, tags); err != nil {
		return nil, err
	}
	return s.contact(ctx, http.MethodDelete, contactPath(contactID, "/tags"), nil, map[string][]string{"tags": tags})
}

func checkTags(contactID string, tags []string) error {
	if err := required("contactId", contactID); err != nil {
		return err
	}
	for _, t := range tags {
		if strings.TrimSpace(t) != "" {
			return nil
		}
	}
	return &ValidationError{Field: "tags", Message: fmt.Sprintf("at least one tag is required (got %d blank)", len(tags))}
}
