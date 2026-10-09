package metigan

import "time"

// ---- Email -----------------------------------------------------------------

// EmailType tells the API how to route a message.
type EmailType string

const (
	// EmailTypeTransactional is for codes, resets and receipts: realtime
	// queue, no List-Unsubscribe header, links not tracked by default.
	EmailTypeTransactional EmailType = "transactional"
	// EmailTypeMarketing is for newsletters and campaigns.
	EmailTypeMarketing EmailType = "marketing"
)

// Attachment is a file attached to an email. Content is the raw file; the
// SDK base64-encodes it.
type Attachment struct {
	Filename    string `json:"filename"`
	Content     []byte `json:"content"`
	ContentType string `json:"contentType,omitempty"`
}

// EmailOptions is an email sent with EmailService.SendEmail.
type EmailOptions struct {
	// From is "email@domain" or "Name <email@domain>". The domain must be
	// verified, except for the shared sender onboarding@metigan.io.
	From string `json:"from"`
	// Recipients: up to 1000 addresses; each one gets its own message.
	Recipients []string `json:"recipients"`
	Subject    string   `json:"subject,omitempty"`
	// Content is the HTML body.
	Content string `json:"content,omitempty"`
	// Text is the plain-text part (optional).
	Text string `json:"text,omitempty"`
	// CC and BCC: at most 49 addresses together.
	CC          []string     `json:"cc,omitempty"`
	BCC         []string     `json:"bcc,omitempty"`
	ReplyTo     string       `json:"replyTo,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	// Type is transactional or marketing. Empty: one recipient is
	// transactional, several are marketing.
	Type        EmailType         `json:"type,omitempty"`
	TrackOpens  *bool             `json:"trackOpens,omitempty"`
	TrackClicks *bool             `json:"trackClicks,omitempty"`
	Unsubscribe *bool             `json:"unsubscribe,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	// IdempotencyKey makes a send safe to repeat: the API sends at most once
	// per key. When empty the SDK generates one per call, so its own retries
	// never duplicate an email.
	IdempotencyKey string `json:"idempotencyKey,omitempty"`
}

// SentEmail is one recipient accepted by the API.
type SentEmail struct {
	Recipient  string `json:"recipient"`
	EmailID    string `json:"emailId"`
	TrackingID string `json:"trackingId"`
	JobID      string `json:"jobId"`
}

// FailedEmail is one recipient the API refused.
type FailedEmail struct {
	Recipient string `json:"recipient"`
	Error     string `json:"error"`
}

// Usage is the account's monthly sending quota.
type Usage struct {
	Sent        int64  `json:"sent"`
	Limit       int64  `json:"limit"`
	Remaining   int64  `json:"remaining"`
	PercentUsed string `json:"percentUsed"`
}

// SendEmailResponse is the result of SendEmail. Emails are queued, not yet
// delivered: use EmailService.Get with an EmailID to follow one.
type SendEmailResponse struct {
	Success          bool          `json:"success"`
	Message          string        `json:"message"`
	Type             string        `json:"type"`
	SuccessfulEmails []SentEmail   `json:"successfulEmails"`
	FailedEmails     []FailedEmail `json:"failedEmails"`
	RecipientCount   int           `json:"recipientCount"`
	HasAttachments   bool          `json:"hasAttachments"`
	AttachmentsCount int           `json:"attachmentsCount"`
	IsTemplate       bool          `json:"isTemplate"`
	Plan             string        `json:"plan"`
	EmailsRemaining  int64         `json:"emailsRemaining"`
	Usage            Usage         `json:"usage"`
}

// OTPOptions is a one-time code sent with EmailService.SendOTP. The API
// renders a localized code email unless TemplateID is set.
type OTPOptions struct {
	To   string `json:"to"`
	From string `json:"from"`
	// Code: up to 32 printable characters.
	Code             string `json:"code"`
	AppName          string `json:"appName,omitempty"`
	ExpiresInMinutes int    `json:"expiresInMinutes,omitempty"`
	// Locale of the built-in email, e.g. "en" or "pt" (default "pt").
	Locale  string `json:"locale,omitempty"`
	Subject string `json:"subject,omitempty"`
	ReplyTo string `json:"replyTo,omitempty"`
	// TemplateID renders one of your templates; Variables fill it (the code
	// is available as {{code}}).
	TemplateID     string            `json:"templateId,omitempty"`
	Variables      map[string]any    `json:"variables,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	IdempotencyKey string            `json:"idempotencyKey,omitempty"`
}

// TransactionalOptions is a single message sent with
// EmailService.SendTransactional.
type TransactionalOptions struct {
	To      string `json:"to"`
	From    string `json:"from"`
	Subject string `json:"subject,omitempty"`
	// HTML is the body (or use TemplateID).
	HTML           string            `json:"html,omitempty"`
	Text           string            `json:"text,omitempty"`
	ReplyTo        string            `json:"replyTo,omitempty"`
	TemplateID     string            `json:"templateId,omitempty"`
	Variables      map[string]any    `json:"variables,omitempty"`
	TrackOpens     *bool             `json:"trackOpens,omitempty"`
	TrackClicks    *bool             `json:"trackClicks,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	IdempotencyKey string            `json:"idempotencyKey,omitempty"`
}

// QueuedEmail is the result of SendOTP and SendTransactional.
type QueuedEmail struct {
	Success    bool   `json:"success"`
	Queued     bool   `json:"queued"`
	Status     string `json:"status"`
	EmailID    string `json:"emailId"`
	TrackingID string `json:"trackingId"`
}

// EmailStatus is the delivery state of one email (EmailService.Get).
type EmailStatus struct {
	EmailID    string `json:"emailId"`
	TrackingID string `json:"trackingId"`
	MessageID  string `json:"messageId"`
	// Status: queued, sending, sent, delivered, bounced, failed, ...
	Status        string     `json:"status"`
	Kind          string     `json:"kind"`
	Recipient     string     `json:"recipient"`
	Subject       string     `json:"subject"`
	QueuedAt      *time.Time `json:"queuedAt"`
	SentAt        *time.Time `json:"sentAt"`
	DeliveredAt   *time.Time `json:"deliveredAt"`
	FirstOpenedAt *time.Time `json:"firstOpenedAt"`
	OpenCount     int64      `json:"openCount"`
	ClickCount    int64      `json:"clickCount"`
	BounceReason  string     `json:"bounceReason"`
	FailureReason string     `json:"failureReason"`
	LastError     string     `json:"lastError"`
}

// ---- Contacts --------------------------------------------------------------

// ContactStatus is a contact's subscription state.
type ContactStatus string

const (
	ContactStatusSubscribed   ContactStatus = "subscribed"
	ContactStatusUnsubscribed ContactStatus = "unsubscribed"
	ContactStatusPending      ContactStatus = "pending"
	ContactStatusBounced      ContactStatus = "bounced"
	ContactStatusComplained   ContactStatus = "complained"
)

// Contact is a contact in an audience.
type Contact struct {
	ID           string         `json:"_id"`
	Email        string         `json:"email"`
	FirstName    string         `json:"firstName,omitempty"`
	LastName     string         `json:"lastName,omitempty"`
	Phone        string         `json:"phone,omitempty"`
	Status       ContactStatus  `json:"status"`
	AudienceID   string         `json:"audienceId"`
	Tags         []string       `json:"tags"`
	CustomFields map[string]any `json:"customFields,omitempty"`
	CreatedAt    *time.Time     `json:"createdAt,omitempty"`
	UpdatedAt    *time.Time     `json:"updatedAt,omitempty"`
}

// CreateContactOptions creates a contact. Creating an email that already
// exists in the audience returns the existing contact.
type CreateContactOptions struct {
	Email        string         `json:"email"`
	AudienceID   string         `json:"audienceId"`
	FirstName    string         `json:"firstName,omitempty"`
	LastName     string         `json:"lastName,omitempty"`
	Phone        string         `json:"phone,omitempty"`
	Tags         []string       `json:"tags,omitempty"`
	CustomFields map[string]any `json:"customFields,omitempty"`
	// Status defaults to subscribed.
	Status ContactStatus `json:"status,omitempty"`
}

// UpdateContactOptions changes a contact. Nil fields are left unchanged.
type UpdateContactOptions struct {
	Email        *string         `json:"email,omitempty"`
	FirstName    *string         `json:"firstName,omitempty"`
	LastName     *string         `json:"lastName,omitempty"`
	Phone        *string         `json:"phone,omitempty"`
	Status       *ContactStatus  `json:"status,omitempty"`
	Tags         *[]string       `json:"tags,omitempty"`
	CustomFields *map[string]any `json:"customFields,omitempty"`
}

// ContactListFilters filters ContactsService.List.
type ContactListFilters struct {
	AudienceID string
	Status     ContactStatus
	Tag        string
	Search     string
	Email      string
	Page       int
	Limit      int
}

// ContactListResponse is a page of contacts.
type ContactListResponse struct {
	Contacts   []Contact  `json:"contacts"`
	Pagination Pagination `json:"pagination"`
}

// ---- Audiences -------------------------------------------------------------

// Audience is a list of contacts. Count is its number of contacts.
type Audience struct {
	ID          string     `json:"_id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	Count       int64      `json:"count"`
	CreatedAt   *time.Time `json:"createdAt,omitempty"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

// AudienceWithContacts is an audience with its most recent contacts
// (AudiencesService.Get).
type AudienceWithContacts struct {
	Audience
	Contacts []Contact `json:"contacts"`
}

// CreateAudienceOptions creates an audience.
type CreateAudienceOptions struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// UpdateAudienceOptions changes an audience. Nil fields are left unchanged.
type UpdateAudienceOptions struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// AudienceListResponse is a page of audiences.
type AudienceListResponse struct {
	Audiences  []Audience `json:"audiences"`
	Pagination Pagination `json:"pagination"`
}

// AudienceStats counts an audience's contacts by status. GrowthRate is the
// percent added in the last 30 days; SubscriptionRate is a percent string
// with two decimals.
type AudienceStats struct {
	AudienceID       string  `json:"audienceId"`
	Name             string  `json:"name"`
	Total            int64   `json:"total"`
	Subscribed       int64   `json:"subscribed"`
	Unsubscribed     int64   `json:"unsubscribed"`
	Pending          int64   `json:"pending"`
	Bounced          int64   `json:"bounced"`
	Complained       int64   `json:"complained"`
	GrowthRate       float64 `json:"growthRate"`
	SubscriptionRate string  `json:"subscriptionRate"`
}

// ---- Pagination ------------------------------------------------------------

// Pagination describes a page of results.
type Pagination struct {
	Total int64 `json:"total"`
	Page  int64 `json:"page"`
	Limit int64 `json:"limit"`
	Pages int64 `json:"pages"`
}

// PaginationOptions selects a page (1-based). Zero values use the API's
// defaults.
type PaginationOptions struct {
	Page  int
	Limit int
}

// ---- Templates -------------------------------------------------------------

// EmailTemplate is a template built in the Metigan editor. Send it with
// EmailService.SendEmailWithTemplate.
type EmailTemplate struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	Subject    string           `json:"subject"`
	Components []map[string]any `json:"components"`
	Styles     map[string]any   `json:"styles,omitempty"`
	CreatedAt  *time.Time       `json:"createdAt,omitempty"`
	UpdatedAt  *time.Time       `json:"updatedAt,omitempty"`
}

// TemplateListResponse is a page of templates.
type TemplateListResponse struct {
	Templates  []EmailTemplate `json:"templates"`
	Pagination Pagination      `json:"pagination"`
}

// ---- Forms -----------------------------------------------------------------

// FormFieldType is the type of a form field.
type FormFieldType string

const (
	FormFieldTypeText        FormFieldType = "text"
	FormFieldTypeEmail       FormFieldType = "email"
	FormFieldTypeNumber      FormFieldType = "number"
	FormFieldTypeTextarea    FormFieldType = "textarea"
	FormFieldTypeSelect      FormFieldType = "select"
	FormFieldTypeCheckbox    FormFieldType = "checkbox"
	FormFieldTypeRadio       FormFieldType = "radio"
	FormFieldTypeDate        FormFieldType = "date"
	FormFieldTypePhone       FormFieldType = "phone"
	FormFieldTypeURL         FormFieldType = "url"
	FormFieldTypeFile        FormFieldType = "file"
	FormFieldTypeStep        FormFieldType = "step"
	FormFieldTypePassword    FormFieldType = "password"
	FormFieldTypeRating      FormFieldType = "rating"
	FormFieldTypeSlider      FormFieldType = "slider"
	FormFieldTypeHeading     FormFieldType = "heading"
	FormFieldTypeImageChoice FormFieldType = "image-choice"
	FormFieldTypeMatrix      FormFieldType = "matrix"
)

// FormField is one field of a form.
type FormField struct {
	ID          string         `json:"id"`
	Type        FormFieldType  `json:"type"`
	Label       string         `json:"label"`
	Placeholder string         `json:"placeholder,omitempty"`
	Required    bool           `json:"required"`
	Options     []string       `json:"options,omitempty"`
	HelpText    string         `json:"helpText,omitempty"`
	Validation  map[string]any `json:"validation,omitempty"`
}

// FormSettings are a form's behavior settings.
type FormSettings struct {
	RedirectURL              string `json:"redirectUrl,omitempty"`
	SuccessMessage           string `json:"successMessage"`
	ProcessingMessage        string `json:"processingMessage"`
	EnableCaptcha            bool   `json:"enableCaptcha"`
	StoreResponses           bool   `json:"storeResponses"`
	NotifyEmail              string `json:"notifyEmail,omitempty"`
	AllowMultipleSubmissions bool   `json:"allowMultipleSubmissions"`
	ShowProgressBar          bool   `json:"showProgressBar"`
}

// FormAnalytics are a form's counters.
type FormAnalytics struct {
	Views                 int64   `json:"views"`
	UniqueVisitors        int64   `json:"uniqueVisitors"`
	Submissions           int64   `json:"submissions"`
	AverageCompletionTime float64 `json:"averageCompletionTime"`
}

// Form is a form. ID is "form-<id>"; every method accepts it, the bare id
// or the slug.
type Form struct {
	ID           string        `json:"id"`
	Title        string        `json:"title"`
	Description  string        `json:"description,omitempty"`
	CoverImage   string        `json:"coverImage,omitempty"`
	Fields       []FormField   `json:"fields"`
	AudienceID   string        `json:"audienceId,omitempty"`
	Settings     FormSettings  `json:"settings"`
	Analytics    FormAnalytics `json:"analytics"`
	Published    bool          `json:"published"`
	PublishedURL string        `json:"publishedUrl,omitempty"`
	Slug         string        `json:"slug,omitempty"`
	CreatedAt    *time.Time    `json:"createdAt,omitempty"`
	UpdatedAt    *time.Time    `json:"updatedAt,omitempty"`
}

// FormListResponse is a page of forms.
type FormListResponse struct {
	Forms      []Form     `json:"forms"`
	Pagination Pagination `json:"pagination"`
}

// FormSubmissionData maps field IDs to values.
type FormSubmissionData map[string]any

// FormSubmissionOptions submits a response to a form.
type FormSubmissionOptions struct {
	// FormID is the form's ID, bare id or slug.
	FormID string             `json:"formId"`
	Data   FormSubmissionData `json:"data"`
}

// FormSubmissionResponse is the result of FormsService.Submit.
type FormSubmissionResponse struct {
	Success      bool       `json:"success"`
	Message      string     `json:"message"`
	SubmissionID string     `json:"submissionId"`
	FormID       string     `json:"formId"`
	SubmittedAt  *time.Time `json:"submittedAt,omitempty"`
}

// FormSubmission is a stored response.
type FormSubmission struct {
	ID          string         `json:"id"`
	FormID      string         `json:"formId"`
	Data        map[string]any `json:"data"`
	SubmittedAt *time.Time     `json:"submittedAt,omitempty"`
	IPAddress   string         `json:"ipAddress,omitempty"`
	UserAgent   string         `json:"userAgent,omitempty"`
}

// FormSubmissionListResponse is a page of submissions.
type FormSubmissionListResponse struct {
	Submissions []FormSubmission `json:"submissions"`
	Pagination  Pagination       `json:"pagination"`
}
