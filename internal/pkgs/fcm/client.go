package fcm

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"KANA-SPACE-BACKEND/internal/modules/notification"
)

type client struct {
	messaging *messaging.Client
}

func NewClient(app *firebase.App) (notification.IPushClient, error) {
	m, err := app.Messaging(context.Background())
	if err != nil {
		return nil, fmt.Errorf("gagal inisialisasi FCM messaging client: %w", err)
	}
	return &client{messaging: m}, nil
}

func (c *client) SendToTokens(ctx context.Context, tokens []string, title, body string, data map[string]string) ([]string, error) {
	if len(tokens) == 0 {
		return nil, nil
	}

	message := &messaging.MulticastMessage{
		Tokens:       tokens,
		Notification: &messaging.Notification{Title: title, Body: body},
		Data:         data,
	}

	resp, err := c.messaging.SendEachForMulticast(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("fcm send gagal: %w", err)
	}

	var invalidTokens []string
	if resp.FailureCount > 0 {
		for i, r := range resp.Responses {
			if !r.Success && messaging.IsRegistrationTokenNotRegistered(r.Error) {
				invalidTokens = append(invalidTokens, tokens[i])
			}
		}
	}
	return invalidTokens, nil
}
