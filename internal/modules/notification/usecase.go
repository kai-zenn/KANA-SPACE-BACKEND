package notification

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotificationNotFound = errors.New("notifikasi tidak ditemukan")
	ErrNotNotificationOwner = errors.New("bukan pemilik notifikasi ini")
	ErrEmptyToken           = errors.New("token FCM tidak boleh kosong")
	ErrInvalidPlatform      = errors.New("platform harus android, ios, atau web")
	ErrInvalidOffset        = errors.New("offset tidak valid")
	ErrInvalidLimit         = errors.New("limit tidak valid")
	ErrNotOwner             = errors.New("anda bukan pemilik notifikasi ini")
)

type INotificationUseCase interface {
	Notify(ctx context.Context, input NotifyInput) error
	NotifyBatch(ctx context.Context, inputs []NotifyInput) error
	RegisterDevice(ctx context.Context, userID uuid.UUID, token, platform string) error
	ListMyNotifications(ctx context.Context, userID uuid.UUID, onlyUnread bool, limit, offset int) ([]NotificationResponse, error)
	MarkAsRead(ctx context.Context, id, userID uuid.UUID) error
}

type NotificationUseCase struct {
	notifRepo  INotificationRepository
	deviceRepo IUserDeviceRepository
	pushClient IPushClient
}

type IPushClient interface {
	SendToTokens(ctx context.Context, tokens []string, title, body string, data map[string]string) ([]string, error)
}

func NewNotificationUseCase(
	notifRepo INotificationRepository,
	deviceRepo IUserDeviceRepository,
	pushClient IPushClient,
) INotificationUseCase {
	return &NotificationUseCase{
		notifRepo:  notifRepo,
		deviceRepo: deviceRepo,
		pushClient: pushClient,
	}
}

func (uc *NotificationUseCase) Notify(ctx context.Context, input NotifyInput) error {
	refType := input.ReferenceType
	notif := &Notification{
		UserID:        input.UserID,
		Type:          input.Type,
		Title:         input.Title,
		Body:          input.Body,
		ReferenceType: &refType,
		ReferenceID:   input.ReferenceID,
	}
	if err := uc.notifRepo.Create(ctx, notif); err != nil {
		return fmt.Errorf("gagal simpan notifikasi: %w", err)
	}

	go uc.pushSafely(input.UserID, input.Title, input.Body, input.Data)

	return nil
}

func (uc *NotificationUseCase) NotifyBatch(ctx context.Context, inputs []NotifyInput) error {
	if len(inputs) == 0 {
		return nil
	}

	notifs := make([]Notification, 0, len(inputs))
	for _, in := range inputs {
		refType := in.ReferenceType
		notifs = append(notifs, Notification{
			UserID:        in.UserID,
			Type:          in.Type,
			Title:         in.Title,
			Body:          in.Body,
			ReferenceType: &refType,
			ReferenceID:   in.ReferenceID,
		})
	}
	if err := uc.notifRepo.CreateInBatches(ctx, notifs, 100); err != nil {
		return fmt.Errorf("gagal simpan notifikasi batch: %w", err)
	}

	for _, in := range inputs {
		go uc.pushSafely(in.UserID, in.Title, in.Body, in.Data)
	}
	return nil
}

func (uc *NotificationUseCase) pushSafely(userID uuid.UUID, title, body string, data map[string]string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Notification] panic saat push FCM: %v", r)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tokens, err := uc.deviceRepo.FindTokensByUserID(ctx, userID)
	if err != nil || len(tokens) == 0 {
		return
	}

	invalidTokens, err := uc.pushClient.SendToTokens(ctx, tokens, title, body, data)
	if err != nil {
		log.Printf("[Notification] gagal push FCM ke user %s: %v", userID, err)
		return
	}

	for _, invalidToken := range invalidTokens {
		if err := uc.deviceRepo.DeleteByToken(context.Background(), invalidToken); err != nil {
			log.Printf("[Notification] gagal hapus token invalid: %v", err)
		}
	}
}

func (uc *NotificationUseCase) RegisterDevice(ctx context.Context, userID uuid.UUID, token, platform string) error {
	return uc.deviceRepo.UpsertToken(ctx, userID, token, platform)
}

func (uc *NotificationUseCase) ListMyNotifications(ctx context.Context, userID uuid.UUID, onlyUnread bool, limit, offset int) ([]NotificationResponse, error) {
	notifs, err := uc.notifRepo.ListByUser(ctx, userID, onlyUnread, limit, offset)
	if err != nil {
		return nil, err
	}

	resp := make([]NotificationResponse, len(notifs))
	for i, n := range notifs {
		resp[i] = NotificationResponse{
			ID:            n.ID,
			Type:          n.Type,
			Title:         n.Title,
			Body:          n.Body,
			ReferenceType: n.ReferenceType,
			ReferenceID:   n.ReferenceID,
			IsRead:        n.IsRead,
			CreatedAt:     n.CreatedAt,
		}
	}
	return resp, nil
}

func (uc *NotificationUseCase) MarkAsRead(ctx context.Context, id, userID uuid.UUID) error {
	return uc.notifRepo.MarkAsRead(ctx, id, userID)
}
