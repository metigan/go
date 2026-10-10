# Metigan Go SDK

Official Go client for the [Metigan](https://metigan.io) API: send emails, OTP codes and transactional messages, and manage contacts, audiences, templates and forms. Standard library only, no dependencies.

## Installation

```bash
go get github.com/metigan/go@latest
```

Requires Go 1.21+. The package name is `metigan`:

```go
import metigan "github.com/metigan/go"
```

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	metigan "github.com/metigan/go"
)

func main() {
	client := metigan.NewClient(metigan.Config{APIKey: os.Getenv("METIGAN_API_KEY")})

	res, err := client.Email().SendEmail(context.Background(), metigan.EmailOptions{
		From:       "My App <onboarding@metigan.io>",
		Recipients: []string{"customer@example.com"},
		Subject:    "Welcome!",
		Content:    "<h1>Hello!</h1><p>Thanks for signing up.</p>",
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, e := range res.SuccessfulEmails {
		fmt.Println("queued", e.Recipient, e.EmailID)
	}
}
```

`onboarding@metigan.io` works on every account without verifying a domain. To send from your own address, verify its domain in the dashboard first.

## Configuration

```go
client := metigan.NewClient(metigan.Config{
	APIKey:        "sk_...",            // required, sent as X-Api-Key
	BaseURL:       "",                  // default: $METIGAN_API_URL, then https://api.metigan.io
	Timeout:       30 * time.Second,    // per attempt
	RetryCount:    3,                   // retries on network errors and 5xx; negative disables
	RetryDelay:    time.Second,         // multiplied by the attempt number
	HTTPClient:    nil,                 // your own *http.Client
	WebhookSecret: "whsec_...",         // for client.Webhooks().Verify
})
```

Every call takes a `context.Context` for cancellation and deadlines. The client is safe for concurrent use.

## Email

### Send

```go
res, err := client.Email().SendEmail(ctx, metigan.EmailOptions{
	From:       "Acme <hello@acme.com>",
	Recipients: []string{"a@example.com", "b@example.com"}, // up to 1000, one message each
	Subject:    "Your invoice",
	Content:    "<p>HTML body</p>",
	Text:       "Plain-text body",              // optional
	CC:         []string{"finance@acme.com"},   // CC+BCC: at most 49
	ReplyTo:    "support@acme.com",
	Type:       metigan.EmailTypeTransactional, // or EmailTypeMarketing
	Attachments: []metigan.Attachment{
		{Filename: "invoice.pdf", Content: pdfBytes, ContentType: "application/pdf"},
	},
})
```

Transactional sends accept at most 5 recipients (`metigan.MaxTransactionalRecipients`); for more people use `EmailTypeMarketing`. Each account's first 500 transactional emails a day skip the sending-IP warm-up, the rest are paced like marketing.

A successful response means the emails were **queued**. `res.SuccessfulEmails` has one `EmailID` per recipient, `res.FailedEmails` the ones refused, and `res.Usage` your monthly quota.

Sends are idempotent: the SDK attaches a random `IdempotencyKey` to each call, so its retries never send twice. Set your own key (for example an order ID) to make your application's retries safe too.

### Send a template

```go
res, err := client.Email().SendEmailWithTemplate(ctx, "TEMPLATE_ID",
	map[string]any{"name": "Ana", "plan": "Pro"}, // fills {{name}}, {{plan}}
	metigan.EmailOptions{
		From:       "Acme <hello@acme.com>",
		Recipients: []string{"ana@example.com"},
	})
```

Subject and body come from the template; set `Subject` to override it.

### OTP codes

```go
q, err := client.Email().SendOTP(ctx, metigan.OTPOptions{
	To:               "user@example.com",
	From:             "Acme <security@acme.com>",
	Code:             "482913",
	AppName:          "Acme",
	ExpiresInMinutes: 10,
	Locale:           "pt", // built-in localized email
})
```

### Transactional message

```go
q, err := client.Email().SendTransactional(ctx, metigan.TransactionalOptions{
	To:      "user@example.com",
	From:    "Acme <hello@acme.com>",
	Subject: "Password changed",
	HTML:    "<p>Your password was changed.</p>",
})
```

### Delivery status

```go
st, err := client.Email().Get(ctx, emailID)
fmt.Println(st.Status, st.DeliveredAt, st.OpenCount, st.BounceReason)
```

## Contacts

```go
c, err := client.Contacts().Create(ctx, metigan.CreateContactOptions{
	Email:        "ana@example.com",
	AudienceID:   "AUDIENCE_ID",
	FirstName:    "Ana",
	Tags:         []string{"customer"},
	CustomFields: map[string]any{"plan": "pro"},
})

c, err = client.Contacts().Get(ctx, c.ID)
c, err = client.Contacts().GetByEmail(ctx, "ana@example.com", "AUDIENCE_ID")

name := "Ana Maria"
c, err = client.Contacts().Update(ctx, c.ID, metigan.UpdateContactOptions{FirstName: &name})

page, err := client.Contacts().List(ctx, metigan.ContactListFilters{
	AudienceID: "AUDIENCE_ID", Status: metigan.ContactStatusSubscribed, Page: 1, Limit: 50,
})
found, err := client.Contacts().Search(ctx, "ana", "AUDIENCE_ID")

c, err = client.Contacts().AddTags(ctx, c.ID, []string{"vip"})
c, err = client.Contacts().RemoveTags(ctx, c.ID, []string{"vip"})
c, err = client.Contacts().Unsubscribe(ctx, c.ID)
c, err = client.Contacts().Subscribe(ctx, c.ID)
err = client.Contacts().Delete(ctx, c.ID)
```

## Audiences

```go
a, err := client.Audiences().Create(ctx, metigan.CreateAudienceOptions{Name: "Newsletter"})
full, err := client.Audiences().Get(ctx, a.ID) // audience + recent contacts
list, err := client.Audiences().List(ctx, metigan.PaginationOptions{Page: 1, Limit: 20})
stats, err := client.Audiences().GetStats(ctx, a.ID)

name := "Weekly newsletter"
a, err = client.Audiences().Update(ctx, a.ID, metigan.UpdateAudienceOptions{Name: &name})
err = client.Audiences().Delete(ctx, a.ID)
```

## Templates

```go
list, err := client.Templates().List(ctx, metigan.PaginationOptions{})
tpl, err := client.Templates().Get(ctx, "TEMPLATE_ID")
```

## Forms

```go
form, err := client.Forms().Get(ctx, "form-id-or-slug")
forms, err := client.Forms().List(ctx, metigan.PaginationOptions{})

sub, err := client.Forms().Submit(ctx, metigan.FormSubmissionOptions{
	FormID: form.ID,
	Data:   metigan.FormSubmissionData{"field-email": "ana@example.com"},
})
subs, err := client.Forms().ListSubmissions(ctx, form.ID, metigan.PaginationOptions{})
```

## Webhooks

Verify every delivery before trusting it. Read the **raw** body first: re-encoding parsed JSON changes the bytes and breaks the signature.

```go
http.HandleFunc("/webhooks/metigan", func(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}
	event, err := client.Webhooks().Verify(body, r.Header) // uses Config.WebhookSecret
	if err != nil {
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}
	switch event.Event {
	case metigan.EventEmailBounced:
		log.Println("bounced:", event.Data["recipient"])
	}
	w.WriteHeader(http.StatusNoContent)
})
```

Without a client: `metigan.VerifyWebhook(body, r.Header, secret, metigan.DefaultWebhookTolerance)`.

## Errors

```go
_, err := client.Email().SendEmail(ctx, opts)

var apiErr *metigan.APIError
var valErr *metigan.ValidationError
switch {
case errors.As(err, &valErr):
	// rejected by the SDK before calling the API (missing field)
	log.Println(valErr.Field, valErr.Message)
case errors.As(err, &apiErr):
	// non-2xx from the API: 400 invalid input, 403 bad key or domain, 429 limit...
	log.Println(apiErr.StatusCode, apiErr.Err, apiErr.Message)
case err != nil:
	// network error or context cancelled
}
```

## License

MIT
