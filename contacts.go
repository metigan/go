package metigan

import (
	"fmt"
)

// ContactsService handles contact operations
type ContactsService struct {
	client *Client
}

// Create creates a new contact
func (s *ContactsService) Create(options CreateContactOptions) (*Contact, error) {
	if options.Email == "" {
		return nil, &ValidationError{Message: "email is required", Field: "email"}
	}
	if options.AudienceID == "" {
		return nil, &ValidationError{Message: "audience ID is required", Field: "audienceId"}
	}

	// Set default status if not provided
	if options.Status == "" {
		options.Status = ContactStatusSubscribed
	}

	resp, err := s.client.makeRequest("POST", "/contacts", options)
	if err != nil {
		return nil, err
	}

	var contact Contact
	if err := s.client.parseResponse(resp, &contact); err != nil {
		return nil, err
	}

	return &contact, nil
}

// Get retrieves a contact by ID
func (s *ContactsService) Get(contactID string) (*Contact, error) {
	if contactID == "" {
		return nil, &ValidationError{Message: "contact ID is required", Field: "contactId"}
	}

	resp, err := s.client.makeRequest("GET", fmt.Sprintf("/contacts/%s", contactID), nil)
	if err != nil {
		return nil, err
	}

	var contact Contact
	if err := s.client.parseResponse(resp, &contact); err != nil {
		return nil, err
	}

	return &contact, nil
}

// GetByEmail retrieves a contact by email address
func (s *ContactsService) GetByEmail(email, audienceID string) (*Contact, error) {
	if email == "" {
		return nil, &ValidationError{Message: "email is required", Field: "email"}
	}
	if audienceID == "" {
		return nil, &ValidationError{Message: "audience ID is required", Field: "audienceId"}
	}

	resp, err := s.client.makeRequest("GET", fmt.Sprintf("/contacts/email/%s?audienceId=%s", email, audienceID), nil)
	if err != nil {
		return nil, err
	}

	var contact Contact
	if err := s.client.parseResponse(resp, &contact); err != nil {
		return nil, err
	}

	return &contact, nil
}

// Update updates an existing contact
func (s *ContactsService) Update(contactID string, options UpdateContactOptions) (*Contact, error) {
	if contactID == "" {
		return nil, &ValidationError{Message: "contact ID is required", Field: "contactId"}
	}

	resp, err := s.client.makeRequest("PATCH", fmt.Sprintf("/contacts/%s", contactID), options)
	if err != nil {
		return nil, err
	}

	var contact Contact
	if err := s.client.parseResponse(resp, &contact); err != nil {
		return nil, err
	}

	return &contact, nil
}

// List lists contacts with optional filters
func (s *ContactsService) List(filters ContactListFilters) (*ContactListResponse, error) {
	// Build query string
	query := "?"
	if filters.AudienceID != "" {
		query += fmt.Sprintf("audienceId=%s&", filters.AudienceID)
	}
	if filters.Status != "" {
		query += fmt.Sprintf("status=%s&", filters.Status)
	}
	if filters.Tag != "" {
		query += fmt.Sprintf("tag=%s&", filters.Tag)
	}
	if filters.Search != "" {
		query += fmt.Sprintf("search=%s&", filters.Search)
	}
	if filters.Page > 0 {
		query += fmt.Sprintf("page=%d&", filters.Page)
	}
	if filters.Limit > 0 {
		query += fmt.Sprintf("limit=%d&", filters.Limit)
	}

	endpoint := "/contacts"
	if query != "?" {
		endpoint += query[:len(query)-1] // Remove trailing &
	}

	resp, err := s.client.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result ContactListResponse
	if err := s.client.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Delete deletes a contact
func (s *ContactsService) Delete(contactID string) error {
	if contactID == "" {
		return &ValidationError{Message: "contact ID is required", Field: "contactId"}
	}

	resp, err := s.client.makeRequest("DELETE", fmt.Sprintf("/contacts/%s", contactID), nil)
	if err != nil {
		return err
	}

	return s.client.parseResponse(resp, nil)
}

// Subscribe subscribes a contact
func (s *ContactsService) Subscribe(contactID string) error {
	if contactID == "" {
		return &ValidationError{Message: "contact ID is required", Field: "contactId"}
	}

	resp, err := s.client.makeRequest("POST", fmt.Sprintf("/contacts/%s/subscribe", contactID), nil)
	if err != nil {
		return err
	}

	return s.client.parseResponse(resp, nil)
}

// Unsubscribe unsubscribes a contact
func (s *ContactsService) Unsubscribe(contactID string) error {
	if contactID == "" {
		return &ValidationError{Message: "contact ID is required", Field: "contactId"}
	}

	resp, err := s.client.makeRequest("POST", fmt.Sprintf("/contacts/%s/unsubscribe", contactID), nil)
	if err != nil {
		return err
	}

	return s.client.parseResponse(resp, nil)
}

// AddTags adds tags to a contact
func (s *ContactsService) AddTags(contactID string, tags []string) error {
	if contactID == "" {
		return &ValidationError{Message: "contact ID is required", Field: "contactId"}
	}
	if len(tags) == 0 {
		return &ValidationError{Message: "at least one tag is required", Field: "tags"}
	}

	resp, err := s.client.makeRequest("POST", fmt.Sprintf("/contacts/%s/tags", contactID), map[string]interface{}{
		"tags": tags,
	})
	if err != nil {
		return err
	}

	return s.client.parseResponse(resp, nil)
}

// RemoveTags removes tags from a contact
func (s *ContactsService) RemoveTags(contactID string, tags []string) error {
	if contactID == "" {
		return &ValidationError{Message: "contact ID is required", Field: "contactId"}
	}
	if len(tags) == 0 {
		return &ValidationError{Message: "at least one tag is required", Field: "tags"}
	}

	resp, err := s.client.makeRequest("DELETE", fmt.Sprintf("/contacts/%s/tags", contactID), map[string]interface{}{
		"tags": tags,
	})
	if err != nil {
		return err
	}

	return s.client.parseResponse(resp, nil)
}

