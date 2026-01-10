package metigan

import "time"

// EmailAttachment represents an email attachment
type Attachment struct {
	Content     []byte `json:"content"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
}

// EmailOptions represents options for sending an email
type EmailOptions struct {
	From       string       `json:"from"`
	Recipients []string     `json:"recipients"`
	Subject    string       `json:"subject"`
	Content    string       `json:"content"`
	Attachments []Attachment `json:"attachments,omitempty"`
	CC         []string     `json:"cc,omitempty"`
	BCC        []string     `json:"bcc,omitempty"`
	ReplyTo    string       `json:"replyTo,omitempty"`
	TrackingID string       `json:"trackingId,omitempty"`
}

// EmailSuccessResponse represents a successful email response
type EmailSuccessResponse struct {
	Success          bool   `json:"success"`
	Message          string `json:"message"`
	SuccessfulEmails []struct {
		Recipient  string `json:"recipient"`
		TrackingID string `json:"trackingId"`
	} `json:"successfulEmails"`
	FailedEmails    []struct {
		Recipient string `json:"recipient"`
		Error     string `json:"error"`
	} `json:"failedEmails"`
	RecipientCount  int `json:"recipientCount"`
	EmailsRemaining int `json:"emailsRemaining"`
}

// ContactStatus represents the status of a contact
type ContactStatus string

const (
	ContactStatusSubscribed   ContactStatus = "subscribed"
	ContactStatusUnsubscribed ContactStatus = "unsubscribed"
	ContactStatusPending      ContactStatus = "pending"
	ContactStatusBounced      ContactStatus = "bounced"
	ContactStatusComplained   ContactStatus = "complained"
)

// Contact represents a contact
type Contact struct {
	ID          string                 `json:"id,omitempty"`
	Email       string                 `json:"email"`
	FirstName   string                 `json:"firstName,omitempty"`
	LastName    string                 `json:"lastName,omitempty"`
	Phone       string                 `json:"phone,omitempty"`
	Status      ContactStatus          `json:"status"`
	AudienceID  string                 `json:"audienceId"`
	Tags        []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"customFields,omitempty"`
	CreatedAt   *time.Time             `json:"createdAt,omitempty"`
	UpdatedAt   *time.Time             `json:"updatedAt,omitempty"`
}

// CreateContactOptions represents options for creating a contact
type CreateContactOptions struct {
	Email       string                 `json:"email"`
	FirstName   string                 `json:"firstName,omitempty"`
	LastName    string                 `json:"lastName,omitempty"`
	Phone       string                 `json:"phone,omitempty"`
	AudienceID  string                 `json:"audienceId"`
	Tags        []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"customFields,omitempty"`
	Status      ContactStatus          `json:"status,omitempty"`
}

// UpdateContactOptions represents options for updating a contact
type UpdateContactOptions struct {
	FirstName   string                 `json:"firstName,omitempty"`
	LastName    string                 `json:"lastName,omitempty"`
	Phone       string                 `json:"phone,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"customFields,omitempty"`
	Status      ContactStatus          `json:"status,omitempty"`
}

// ContactListFilters represents filters for listing contacts
type ContactListFilters struct {
	AudienceID string        `json:"audienceId,omitempty"`
	Status     ContactStatus `json:"status,omitempty"`
	Tag        string        `json:"tag,omitempty"`
	Search     string        `json:"search,omitempty"`
	Page       int           `json:"page,omitempty"`
	Limit      int           `json:"limit,omitempty"`
}

// ContactListResponse represents a paginated list of contacts
type ContactListResponse struct {
	Contacts []Contact `json:"contacts"`
	Pagination Pagination `json:"pagination"`
}

// Audience represents an audience
type Audience struct {
	ID          string     `json:"id,omitempty"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Count       int        `json:"count"`
	CreatedAt   *time.Time `json:"createdAt,omitempty"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

// CreateAudienceOptions represents options for creating an audience
type CreateAudienceOptions struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// AudienceListResponse represents a paginated list of audiences
type AudienceListResponse struct {
	Audiences []Audience `json:"audiences"`
	Pagination Pagination `json:"pagination"`
}

// AudienceStats represents statistics for an audience
type AudienceStats struct {
	Total        int `json:"total"`
	Subscribed   int `json:"subscribed"`
	Unsubscribed int `json:"unsubscribed"`
	Pending      int `json:"pending"`
	Bounced      int `json:"bounced"`
	Complained   int `json:"complained"`
}

// Pagination represents pagination information
type Pagination struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Pages int `json:"pages"`
}

// PaginationOptions represents options for pagination
type PaginationOptions struct {
	Page  int `json:"page,omitempty"`
	Limit int `json:"limit,omitempty"`
}

// EmailTemplate represents an email template
type EmailTemplate struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Subject   string     `json:"subject"`
	Content   string     `json:"content,omitempty"`
	Variables map[string]interface{} `json:"variables,omitempty"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

// FormFieldType represents the type of a form field
type FormFieldType string

const (
	FormFieldTypeText     FormFieldType = "text"
	FormFieldTypeEmail    FormFieldType = "email"
	FormFieldTypeNumber   FormFieldType = "number"
	FormFieldTypeTextarea FormFieldType = "textarea"
	FormFieldTypeSelect   FormFieldType = "select"
	FormFieldTypeCheckbox FormFieldType = "checkbox"
	FormFieldTypeRadio    FormFieldType = "radio"
	FormFieldTypeDate     FormFieldType = "date"
	FormFieldTypePhone    FormFieldType = "phone"
	FormFieldTypeURL      FormFieldType = "url"
	FormFieldTypeFile     FormFieldType = "file"
)

// FormField represents a form field configuration
type FormField struct {
	ID          string        `json:"id"`
	Type        FormFieldType `json:"type"`
	Label       string        `json:"label"`
	Placeholder string        `json:"placeholder,omitempty"`
	Required    bool          `json:"required,omitempty"`
	HelpText    string        `json:"helpText,omitempty"`
	Options     []string      `json:"options,omitempty"`
}

// Form represents a form configuration
type Form struct {
	ID          string     `json:"id,omitempty"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Fields      []FormField `json:"fields"`
	Settings    FormSettings `json:"settings,omitempty"`
	Published   bool       `json:"published,omitempty"`
	PublishedURL string    `json:"publishedUrl,omitempty"`
	Slug        string     `json:"slug,omitempty"`
	CreatedAt   *time.Time `json:"createdAt,omitempty"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

// FormSettings represents form settings
type FormSettings struct {
	SuccessMessage           string `json:"successMessage,omitempty"`
	ProcessingMessage        string `json:"processingMessage,omitempty"`
	RedirectURL              string `json:"redirectUrl,omitempty"`
	NotifyEmail              string `json:"notifyEmail,omitempty"`
	EnableCaptcha            bool   `json:"enableCaptcha,omitempty"`
	StoreResponses           bool   `json:"storeResponses,omitempty"`
	AllowMultipleSubmissions bool   `json:"allowMultipleSubmissions,omitempty"`
}

// FormSubmissionData represents form submission data
type FormSubmissionData map[string]interface{}

// FormSubmissionOptions represents options for submitting a form
type FormSubmissionOptions struct {
	FormID string            `json:"formId"`
	Data   FormSubmissionData `json:"data"`
}

// FormSubmissionResponse represents a form submission response
type FormSubmissionResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	SubmissionID string `json:"submissionId,omitempty"`
}

// FormListResponse represents a paginated list of forms
type FormListResponse struct {
	Forms []Form `json:"forms"`
	Pagination Pagination `json:"pagination"`
}

