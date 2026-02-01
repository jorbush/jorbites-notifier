package push

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/jorbush/jorbites-notifier/config"
	"github.com/jorbush/jorbites-notifier/internal/database"
	"github.com/jorbush/jorbites-notifier/internal/models"
)

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
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusGone || resp.StatusCode == http.StatusNotFound {
		log.Printf("Subscription expired or not found, deleting... %s", subscription.ID)
		if err := p.db.DeletePushSubscription(context.Background(), subscription.ID); err != nil {
			log.Printf("Error deleting subscription %s: %v", subscription.ID, err)
		}
	}

	return nil
}
