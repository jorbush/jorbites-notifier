package push

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/jorbush/jorbites-notifier/config"
	"github.com/jorbush/jorbites-notifier/internal/database"
	"github.com/jorbush/jorbites-notifier/internal/models"
)

var allowedPushHosts = map[string]struct{}{
	"fcm.googleapis.com":                {},
	"android.googleapis.com":            {},
	"updates.push.services.mozilla.com": {},
}

var allowedPushHostSuffixes = []string{
	".push.apple.com",
	".notify.windows.com",
	".push.services.mozilla.com",
	".push.samsungosp.com",
}

// ValidatePushEndpoint ensures a Web Push subscription endpoint is HTTPS,
// has no credentials or custom port, is not an IP address, and belongs to a
// trusted Web Push provider domain to prevent SSRF.
func ValidatePushEndpoint(endpoint string) error {
	if endpoint == "" {
		return errors.New("push endpoint is empty")
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid push endpoint URL: %w", err)
	}

	if parsed.Scheme != "https" {
		return errors.New("push endpoint must use https")
	}

	if parsed.User != nil {
		return errors.New("push endpoint must not contain credentials")
	}

	if parsed.Port() != "" {
		return errors.New("push endpoint must not specify a custom port")
	}

	host := strings.ToLower(parsed.Hostname())
	if host == "" || net.ParseIP(host) != nil {
		return errors.New("push endpoint host must be a valid domain")
	}

	if _, ok := allowedPushHosts[host]; ok {
		return nil
	}

	for _, suffix := range allowedPushHostSuffixes {
		if strings.HasSuffix(host, suffix) && len(host) > len(suffix) {
			return nil
		}
	}

	return fmt.Errorf("push endpoint host %q is not an allowed push service", host)
}

type PushSender struct {
	config *config.Config
	db     *database.MongoDB
}

func NewPushSender(cfg *config.Config, db *database.MongoDB) *PushSender {
	return &PushSender{
		config: cfg,
		db:     db,
	}
}

func (p *PushSender) SendNotification(subscription models.PushSubscription, title, message, url string) error {
	if err := ValidatePushEndpoint(subscription.Endpoint); err != nil {
		slog.Warn("Rejected untrusted push subscription endpoint", "error", err)
		return err
	}

	s := &webpush.Subscription{
		Endpoint: subscription.Endpoint,
		Keys: webpush.Keys{
			P256dh: subscription.P256dh,
			Auth:   subscription.Auth,
		},
	}

	payload, err := json.Marshal(map[string]string{
		"title": title,
		"body":  message,
		"icon":  "/web-app-manifest-192x192.png",
		"url":   url,
	})
	if err != nil {
		return err
	}

	resp, err := webpush.SendNotification(payload, s, &webpush.Options{
		Subscriber:      p.config.VAPIDSubject,
		VAPIDPublicKey:  p.config.VAPIDPublicKey,
		VAPIDPrivateKey: p.config.VAPIDPrivateKey,
		TTL:             30,
	})

	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusGone || resp.StatusCode == http.StatusNotFound {
		slog.Info("Subscription expired or not found, deleting", "subscriptionId", subscription.ID.Hex())
		if err := p.db.DeletePushSubscription(context.Background(), subscription.ID.Hex()); err != nil {
			slog.Error("Error deleting subscription", "subscriptionId", subscription.ID.Hex(), "error", err)
		}
	}

	return nil
}
