package metigan

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// DefaultWebhookTolerance is how far a webhook's timestamp may be from now
// before Verify rejects it (replay protection).
const DefaultWebhookTolerance = 5 * time.Minute

// Webhook event names.
const (
	EventEmailSent         = "email.sent"
	EventEmailDelivered    = "email.delivered"
	EventEmailOpened       = "email.opened"
	EventEmailClicked      = "email.clicked"
	EventEmailBounced      = "email.bounced"
	EventEmailComplained   = "email.complained"
	EventEmailUnsubscribed = "email.unsubscribed"
	EventEmailFailed       = "email.failed"
	EventContactCreated    = "contact.created"
	EventContactDeleted    = "contact.deleted"
	EventAudienceCreated   = "audience.created"
	EventAudienceUpdated   = "audience.updated"
	EventAudienceDeleted   = "audience.deleted"
)

// WebhookEvent is a verified webhook body. Timestamp is in milliseconds
// since the epoch; Data depends on Event.
type WebhookEvent struct {
	Event     string         `json:"event"`
	MessageID string         `json:"messageId"`
	Data      map[string]any `json:"data"`
	Timestamp int64          `json:"timestamp"`
}

// Webhook verification errors. Reject the request when Verify returns any
// error.
var (
	ErrWebhookMissingSecret    = errors.New("metigan: webhook signing secret is required")
	ErrWebhookMissingSignature = errors.New("metigan: missing X-Webhook-Signature header")
	ErrWebhookInvalidSignature = errors.New("metigan: malformed X-Webhook-Signature header")
	ErrWebhookTimestamp        = errors.New("metigan: webhook timestamp outside the tolerance window")
	ErrWebhookNoMatch          = errors.New("metigan: webhook signature does not match")
	ErrWebhookInvalidPayload   = errors.New("metigan: webhook body is not a Metigan event")
)

// WebhooksService verifies webhook deliveries. Get it with
// Client.Webhooks; the secret comes from Config.WebhookSecret.
type WebhooksService struct {
	secret    string
	tolerance time.Duration
}

// Verify checks the signature of a delivery and returns its event. rawBody
// must be the exact bytes received (read r.Body before any JSON decoding).
func (s *WebhooksService) Verify(rawBody []byte, header http.Header) (*WebhookEvent, error) {
	return VerifyWebhook(rawBody, header, s.secret, s.tolerance)
}

// VerifyWebhook checks a delivery signed with secret (whsec_...).
//
// Metigan sends X-Webhook-Signature: t=<unix>,v1=<hex HMAC-SHA256(secret,
// "<t>.<body>")>. tolerance bounds the clock skew; 0 disables the check.
//
//	body, _ := io.ReadAll(r.Body)
//	event, err := metigan.VerifyWebhook(body, r.Header, secret, metigan.DefaultWebhookTolerance)
//	if err != nil {
//		http.Error(w, "invalid signature", http.StatusBadRequest)
//		return
//	}
func VerifyWebhook(rawBody []byte, header http.Header, secret string, tolerance time.Duration) (*WebhookEvent, error) {
	return verifyWebhookAt(rawBody, header, secret, tolerance, time.Now())
}

func verifyWebhookAt(rawBody []byte, header http.Header, secret string, tolerance time.Duration, now time.Time) (*WebhookEvent, error) {
	if secret == "" {
		return nil, ErrWebhookMissingSecret
	}
	sigHeader := strings.TrimSpace(header.Get("X-Webhook-Signature"))
	if sigHeader == "" {
		return nil, ErrWebhookMissingSignature
	}
	ts, sigs := parseSignature(sigHeader)
	if len(sigs) == 0 {
		return nil, ErrWebhookInvalidSignature
	}
	if ts < 0 {
		if v, err := strconv.ParseInt(strings.TrimSpace(header.Get("X-Webhook-Timestamp")), 10, 64); err == nil {
			ts = v
		}
	}
	if ts < 0 {
		return nil, ErrWebhookInvalidSignature
	}
	if tolerance > 0 {
		skew := now.Sub(time.Unix(ts, 0))
		if skew < 0 {
			skew = -skew
		}
		if skew > tolerance {
			return nil, ErrWebhookTimestamp
		}
	}

	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(strconv.FormatInt(ts, 10) + "."))
	m.Write(rawBody)
	expected := m.Sum(nil)
	matched := false
	for _, s := range sigs {
		if got, err := hex.DecodeString(s); err == nil && hmac.Equal(got, expected) {
			matched = true
		}
	}
	if !matched {
		return nil, ErrWebhookNoMatch
	}

	var ev WebhookEvent
	if err := json.Unmarshal(rawBody, &ev); err != nil || ev.Event == "" {
		return nil, ErrWebhookInvalidPayload
	}
	return &ev, nil
}

// parseSignature reads "t=<n>,v1=<hex>[,v1=<hex>]"; ts is -1 when absent.
// A bare hex value is accepted as a v1 signature.
func parseSignature(h string) (ts int64, sigs []string) {
	ts = -1
	scheme := false
	for _, tok := range strings.FieldsFunc(h, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' }) {
		k, v, ok := strings.Cut(tok, "=")
		if !ok {
			continue
		}
		switch k {
		case "t":
			scheme = true
			if n, err := strconv.ParseInt(v, 10, 64); err == nil {
				ts = n
			}
		case "v1":
			scheme = true
			if v != "" {
				sigs = append(sigs, strings.ToLower(v))
			}
		}
	}
	if !scheme {
		if _, err := hex.DecodeString(h); err == nil {
			sigs = append(sigs, strings.ToLower(h))
		}
	}
	return ts, sigs
}
