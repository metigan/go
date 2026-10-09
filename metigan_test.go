package metigan

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"
)

type recorded struct {
	Method, Path, RawQuery string
	Header                 http.Header
	Body                   map[string]any
}

// server answers each request with the next (status, body) in replies and
// records what it received.
func server(t *testing.T, replies ...[2]string) (*Client, *[]recorded) {
	t.Helper()
	var mu sync.Mutex
	var got []recorded
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		raw, _ := io.ReadAll(r.Body)
		rec := recorded{Method: r.Method, Path: r.URL.EscapedPath(), RawQuery: r.URL.RawQuery, Header: r.Header}
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &rec.Body); err != nil {
				t.Errorf("request body is not JSON: %s", raw)
			}
		}
		got = append(got, rec)
		reply := replies[len(replies)-1]
		if len(got) <= len(replies) {
			reply = replies[len(got)-1]
		}
		status, _ := strconv.Atoi(reply[0])
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, reply[1])
	}))
	t.Cleanup(srv.Close)
	c := NewClient(Config{APIKey: "sk_test", BaseURL: srv.URL + "/", RetryDelay: time.Millisecond})
	return c, &got
}

func TestSendEmail(t *testing.T) {
	c, got := server(t, [2]string{"200", `{"success":true,"message":"Emails queued for processing","recipientCount":1,
		"successfulEmails":[{"success":true,"recipient":"a@b.co","emailId":"em-1","trackingId":"tr-1","jobId":"j"}],
		"failedEmails":[],"emailsRemaining":99,"usage":{"sent":1,"limit":100,"remaining":99,"percentUsed":"1.00%"}}`})
	res, err := c.Email().SendEmail(context.Background(), EmailOptions{
		From: "App <onboarding@metigan.io>", Recipients: []string{"a@b.co"}, Subject: "Hi", Content: "<p>x</p>",
		Attachments: []Attachment{{Filename: "a.txt", Content: []byte("hello"), ContentType: "text/plain"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.SuccessfulEmails[0].EmailID != "em-1" || res.EmailsRemaining != 99 || res.Usage.Limit != 100 {
		t.Fatalf("response not decoded: %+v", res)
	}
	r := (*got)[0]
	if r.Method != "POST" || r.Path != "/api/email/send" {
		t.Fatalf("request = %s %s", r.Method, r.Path)
	}
	if r.Header.Get("X-Api-Key") != "sk_test" || r.Header.Get("Authorization") != "" {
		t.Fatalf("auth headers = %v", r.Header)
	}
	if r.Body["idempotencyKey"] == "" || r.Body["idempotencyKey"] == nil {
		t.Fatal("no idempotency key generated")
	}
	att := r.Body["attachments"].([]any)[0].(map[string]any)
	if att["content"] != base64.StdEncoding.EncodeToString([]byte("hello")) {
		t.Fatalf("attachment content = %v", att["content"])
	}
	if _, ok := r.Body["useTemplate"]; ok {
		t.Fatal("plain send must not set useTemplate")
	}
}

func TestSendEmailWithTemplate(t *testing.T) {
	c, got := server(t, [2]string{"200", `{"success":true}`})
	_, err := c.Email().SendEmailWithTemplate(context.Background(), "tpl1", map[string]any{"name": "Ana"},
		EmailOptions{From: "a@metigan.io", Recipients: []string{"x@y.co"}})
	if err != nil {
		t.Fatal(err)
	}
	b := (*got)[0].Body
	if b["useTemplate"] != "true" || b["templateId"] != "tpl1" || b["variables"].(map[string]any)["name"] != "Ana" {
		t.Fatalf("template fields = %v", b)
	}
	if _, ok := b["subject"]; ok {
		t.Fatal("empty subject must be omitted")
	}
}

func TestRetryResendsBodyWithSameKey(t *testing.T) {
	c, got := server(t, [2]string{"502", `{"error":"bad gateway"}`}, [2]string{"200", `{"success":true}`})
	_, err := c.Email().SendEmail(context.Background(), EmailOptions{From: "a@metigan.io", Recipients: []string{"x@y.co"}, Subject: "s", Content: "c"})
	if err != nil {
		t.Fatal(err)
	}
	if len(*got) != 2 {
		t.Fatalf("attempts = %d, want 2", len(*got))
	}
	first, second := (*got)[0].Body, (*got)[1].Body
	if second["subject"] != "s" || first["idempotencyKey"] != second["idempotencyKey"] {
		t.Fatalf("retry body differs: %v vs %v", first, second)
	}
}

func TestAPIError(t *testing.T) {
	c, got := server(t, [2]string{"400", `{"error":"Invalid sender","message":"Invalid sender email format"}`})
	_, err := c.Email().SendEmail(context.Background(), EmailOptions{From: "bad", Recipients: []string{"x@y.co"}, Subject: "s", Content: "c"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 400 || apiErr.Message != "Invalid sender email format" {
		t.Fatalf("err = %v", err)
	}
	if len(*got) != 1 {
		t.Fatalf("4xx must not be retried, attempts = %d", len(*got))
	}
}

func TestServerErrorAfterRetries(t *testing.T) {
	c, got := server(t, [2]string{"503", `{"error":"down"}`})
	_, err := c.Templates().List(context.Background(), PaginationOptions{})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 503 {
		t.Fatalf("err = %v", err)
	}
	if len(*got) != 4 {
		t.Fatalf("attempts = %d, want 4", len(*got))
	}
}

func TestValidation(t *testing.T) {
	c := NewClient(Config{APIKey: "k"})
	_, err := c.Email().SendEmail(context.Background(), EmailOptions{Recipients: []string{"x@y.co"}})
	var v *ValidationError
	if !errors.As(err, &v) || v.Field != "from" {
		t.Fatalf("err = %v", err)
	}
	_, err = NewClient(Config{}).Templates().Get(context.Background(), "x")
	if !errors.As(err, &v) || v.Field != "apiKey" {
		t.Fatalf("missing key err = %v", err)
	}
}

func TestEmailStatus(t *testing.T) {
	c, got := server(t, [2]string{"200", `{"success":true,"data":{"emailId":"em/1","status":"delivered",
		"queuedAt":"2026-10-09T10:00:00Z","deliveredAt":"2026-10-09T10:00:02.5Z","openCount":2}}`})
	st, err := c.Email().Get(context.Background(), "em/1")
	if err != nil {
		t.Fatal(err)
	}
	if st.Status != "delivered" || st.OpenCount != 2 || st.DeliveredAt == nil {
		t.Fatalf("status = %+v", st)
	}
	if p := (*got)[0].Path; p != "/api/email/em%2F1" {
		t.Fatalf("path = %s", p)
	}
}

func TestOTPAndTransactional(t *testing.T) {
	c, got := server(t, [2]string{"200", `{"success":true,"queued":true,"status":"queued","emailId":"otp-1","trackingId":"t"}`})
	q, err := c.Email().SendOTP(context.Background(), OTPOptions{To: "x@y.co", From: "a@metigan.io", Code: "123456", ExpiresInMinutes: 5})
	if err != nil || q.EmailID != "otp-1" {
		t.Fatalf("otp = %+v, %v", q, err)
	}
	if _, err := c.Email().SendTransactional(context.Background(), TransactionalOptions{To: "x@y.co", From: "a@metigan.io", Subject: "s", HTML: "<p>h</p>"}); err != nil {
		t.Fatal(err)
	}
	otp, tx := (*got)[0], (*got)[1]
	if otp.Path != "/api/otp/send" || otp.Body["code"] != "123456" || otp.Body["expiresInMinutes"] != float64(5) {
		t.Fatalf("otp request = %+v", otp)
	}
	if tx.Path != "/api/transactional/send" || tx.Body["html"] != "<p>h</p>" {
		t.Fatalf("transactional request = %+v", tx)
	}
}

func TestContacts(t *testing.T) {
	contact := `{"_id":"64b000000000000000000001","email":"a+b@x.co","status":"subscribed","audienceId":"aud","tags":["vip"],
		"createdAt":"2026-10-09T10:00:00.000Z","__v":0,"userId":"u"}`
	c, got := server(t, [2]string{"200", contact})
	ctx := context.Background()
	ct, err := c.Contacts().GetByEmail(ctx, "a+b@x.co", "aud")
	if err != nil {
		t.Fatal(err)
	}
	if ct.ID != "64b000000000000000000001" || ct.Tags[0] != "vip" || ct.CreatedAt == nil {
		t.Fatalf("contact = %+v", ct)
	}
	if r := (*got)[0]; r.Path != "/api/contacts/email/a+b@x.co" || r.RawQuery != "audienceId=aud" {
		t.Fatalf("getByEmail request = %s ? %s", r.Path, r.RawQuery)
	}

	if _, err := c.Contacts().RemoveTags(ctx, "id1", []string{"vip"}); err != nil {
		t.Fatal(err)
	}
	if r := (*got)[1]; r.Method != "DELETE" || r.Path != "/api/contacts/id1/tags" || r.Body["tags"].([]any)[0] != "vip" {
		t.Fatalf("removeTags request = %+v", r)
	}

	name := "Ana"
	if _, err := c.Contacts().Update(ctx, "id1", UpdateContactOptions{FirstName: &name}); err != nil {
		t.Fatal(err)
	}
	if r := (*got)[2]; r.Method != "PATCH" || len(r.Body) != 1 || r.Body["firstName"] != "Ana" {
		t.Fatalf("update request = %+v", r)
	}

	if _, err := c.Contacts().List(ctx, ContactListFilters{AudienceID: "aud", Status: ContactStatusSubscribed, Page: 2}); err != nil {
		t.Fatal(err)
	}
	if r := (*got)[3]; r.RawQuery != "audienceId=aud&page=2&status=subscribed" {
		t.Fatalf("list query = %s", r.RawQuery)
	}
}

func TestAudiencesAndForms(t *testing.T) {
	c, got := server(t,
		[2]string{"200", `{"_id":"a1","name":"News","count":3,"createdAt":"2026-10-09T10:00:00.000Z","contacts":[{"_id":"c1","email":"x@y.co","status":"subscribed"}]}`},
		[2]string{"200", `{"audienceId":"a1","name":"News","total":3,"subscribed":2,"growthRate":12.5,"subscriptionRate":"66.67"}`},
		[2]string{"201", `{"id":"sub1","formId":"form-f1","success":true,"message":"ok","submissionId":"sub1","submittedAt":"2026-10-09T10:00:00.000Z"}`},
	)
	ctx := context.Background()
	a, err := c.Audiences().Get(ctx, "a1")
	if err != nil || a.ID != "a1" || a.Count != 3 || a.Contacts[0].ID != "c1" {
		t.Fatalf("audience = %+v, %v", a, err)
	}
	st, err := c.Audiences().GetStats(ctx, "a1")
	if err != nil || st.Subscribed != 2 || st.SubscriptionRate != "66.67" {
		t.Fatalf("stats = %+v, %v", st, err)
	}
	sub, err := c.Forms().Submit(ctx, FormSubmissionOptions{FormID: "form-f1", Data: FormSubmissionData{"email": "x@y.co"}})
	if err != nil || sub.SubmissionID != "sub1" {
		t.Fatalf("submit = %+v, %v", sub, err)
	}
	if r := (*got)[2]; r.Path != "/api/submissions" || r.Body["formId"] != "form-f1" {
		t.Fatalf("submit request = %+v", r)
	}
}

func sign(secret string, ts int64, body []byte) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(strconv.FormatInt(ts, 10) + "."))
	m.Write(body)
	return "t=" + strconv.FormatInt(ts, 10) + ",v1=" + hex.EncodeToString(m.Sum(nil))
}

func TestVerifyWebhook(t *testing.T) {
	secret := "whsec_" + hex.EncodeToString(make([]byte, 32))
	body := []byte(`{"event":"email.bounced","messageId":"m1","data":{"recipient":"x@y.co"},"timestamp":1760000000000}`)
	now := time.Unix(1760000000, 0)
	h := http.Header{}
	h.Set("X-Webhook-Signature", sign(secret, now.Unix(), body))

	ev, err := verifyWebhookAt(body, h, secret, DefaultWebhookTolerance, now)
	if err != nil || ev.Event != EventEmailBounced || ev.Data["recipient"] != "x@y.co" {
		t.Fatalf("event = %+v, %v", ev, err)
	}
	if _, err := verifyWebhookAt(append([]byte(" "), body...), h, secret, DefaultWebhookTolerance, now); !errors.Is(err, ErrWebhookNoMatch) {
		t.Fatalf("tampered body err = %v", err)
	}
	if _, err := verifyWebhookAt(body, h, secret, DefaultWebhookTolerance, now.Add(10*time.Minute)); !errors.Is(err, ErrWebhookTimestamp) {
		t.Fatalf("stale err = %v", err)
	}
	if _, err := verifyWebhookAt(body, http.Header{}, secret, 0, now); !errors.Is(err, ErrWebhookMissingSignature) {
		t.Fatalf("missing header err = %v", err)
	}
	if _, err := NewClient(Config{}).Webhooks().Verify(body, h); !errors.Is(err, ErrWebhookMissingSecret) {
		t.Fatalf("no secret err = %v", err)
	}
}
