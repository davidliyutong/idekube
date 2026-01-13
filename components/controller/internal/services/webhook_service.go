package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/repository"
)

// WebhookService handles webhook operations
type WebhookService struct {
	webhookRepo *repository.WebhookRepository
	httpClient  *http.Client
}

// NewWebhookService creates a new webhook service
func NewWebhookService(webhookRepo *repository.WebhookRepository) *WebhookService {
	return &WebhookService{
		webhookRepo: webhookRepo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CreateWebhook creates a new webhook
func (s *WebhookService) CreateWebhook(ctx context.Context, userID int64, webhook *models.Webhook) error {
	webhook.UserID = userID
	webhook.Enabled = true
	webhook.CreatedAt = time.Now()
	webhook.UpdatedAt = time.Now()

	return s.webhookRepo.Create(ctx, webhook)
}

// GetWebhook retrieves a webhook by ID
func (s *WebhookService) GetWebhook(ctx context.Context, id, userID int64) (*models.Webhook, error) {
	webhook, err := s.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if webhook.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	return webhook, nil
}

// ListWebhooks lists all webhooks for a user
func (s *WebhookService) ListWebhooks(ctx context.Context, userID int64, events []string) ([]*models.Webhook, error) {
	if len(events) > 0 {
		return s.webhookRepo.ListByUserAndEvents(ctx, userID, events)
	}
	return s.webhookRepo.ListByUser(ctx, userID)
}

// UpdateWebhook updates a webhook
func (s *WebhookService) UpdateWebhook(ctx context.Context, id, userID int64, updates map[string]interface{}) error {
	webhook, err := s.GetWebhook(ctx, id, userID)
	if err != nil {
		return err
	}

	webhook.UpdatedAt = time.Now()
	return s.webhookRepo.Update(ctx, webhook, updates)
}

// DeleteWebhook deletes a webhook
func (s *WebhookService) DeleteWebhook(ctx context.Context, id, userID int64) error {
	webhook, err := s.GetWebhook(ctx, id, userID)
	if err != nil {
		return err
	}

	return s.webhookRepo.Delete(ctx, webhook.ID)
}

// TriggerWebhooks triggers webhooks for an event
func (s *WebhookService) TriggerWebhooks(ctx context.Context, userID int64, event string, payload map[string]interface{}) error {
	webhooks, err := s.webhookRepo.ListByUserAndEvents(ctx, userID, []string{event})
	if err != nil {
		return fmt.Errorf("failed to list webhooks: %w", err)
	}

	for _, webhook := range webhooks {
		if !webhook.Enabled {
			continue
		}

		go s.deliverWebhook(webhook, event, payload)
	}

	return nil
}

// deliverWebhook delivers a webhook with retries
func (s *WebhookService) deliverWebhook(webhook *models.Webhook, event string, payload map[string]interface{}) {
	maxRetries := 3
	retryDelay := time.Second

	webhookPayload := map[string]interface{}{
		"event":     event,
		"timestamp": time.Now().Unix(),
		"data":      payload,
	}

	body, err := json.Marshal(webhookPayload)
	if err != nil {
		return
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		}

		req, err := http.NewRequest("POST", webhook.URL, bytes.NewBuffer(body))
		if err != nil {
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Webhook-Event", event)
		if webhook.Secret != nil {
			// TODO: Add HMAC signature
			req.Header.Set("X-Webhook-Signature", *webhook.Secret)
		}

		resp, err := s.httpClient.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Success
			webhook.LastTriggeredAt = &[]time.Time{time.Now()}[0]
			_ = s.webhookRepo.Update(context.Background(), webhook, map[string]interface{}{
				"last_triggered_at": webhook.LastTriggeredAt,
			})
			return
		}
	}

	// All retries failed
	// TODO: Log failed delivery
}

// TestWebhook tests a webhook by sending a test event
func (s *WebhookService) TestWebhook(ctx context.Context, id, userID int64) error {
	webhook, err := s.GetWebhook(ctx, id, userID)
	if err != nil {
		return err
	}

	testPayload := map[string]interface{}{
		"test":    true,
		"message": "This is a test webhook",
	}

	go s.deliverWebhook(webhook, "webhook.test", testPayload)

	return nil
}
