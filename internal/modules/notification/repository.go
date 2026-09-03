package notification

import (
	"context"
	"errors"
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type INotificationRepository interface {
	Create(ctx context.Context, n *Notification) error
	CreateInBatches(ctx context.Context, ns []Notification, batchSize int) error
	ListByUser(ctx context.Context, userID uuid.UUID, onlyUnread bool, limit, offset int) ([]Notification, error)
	MarkAsRead(ctx context.Context, id, userID uuid.UUID) error
}

type IUserDeviceRepository interface {
	UpsertToken(ctx context.Context, userID uuid.UUID, token, platform string) error
	FindTokensByUserID(ctx context.Context, userID uuid.UUID) ([]string, error)
	DeleteByToken(ctx context.Context, token string) error
}


// == Notification Respository ==
type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) INotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, n *Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *notificationRepository) CreateInBatches(ctx context.Context, ns []Notification, batchSize int) error {
	if len(ns) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(ns, batchSize).Error
}

func (r *notificationRepository) ListByUser(ctx context.Context, userID uuid.UUID, onlyUnread bool, limit, offset int) ([]Notification, error) {
	var notifs []Notification
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if onlyUnread {
		query = query.Where("is_read = ?", false)
	}
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&notifs).Error
	return notifs, err
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id, userID uuid.UUID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&Notification{}).
		Where("id = ? AND user_id = ? AND is_read = ?", id, userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("notification not found or already read")
	}
	return nil
}

// == User Device Repository ==
type userDeviceRepository struct {
	db *gorm.DB
}

func NewUserDeviceRepository(db *gorm.DB) IUserDeviceRepository {
	return &userDeviceRepository{db: db}
}

func (r *userDeviceRepository) UpsertToken(ctx context.Context, userID uuid.UUID, token, platform string) error {
	device := UserDevice{
		UserID:   userID,
		FCMToken: token,
		Platform: platform,
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "fcm_token"}},
			DoUpdates: clause.AssignmentColumns([]string{"user_id", "platform", "updated_at"}),
		}).
		Create(&device).Error
}

func (r *userDeviceRepository) FindTokensByUserID(ctx context.Context, userID uuid.UUID) ([]string, error) {
	var tokens []string
	err := r.db.WithContext(ctx).Model(&UserDevice{}).
		Where("user_id = ?", userID).
		Pluck("fcm_token", &tokens).Error
	return tokens, err
}

func (r *userDeviceRepository) DeleteByToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("fcm_token = ?", token).Delete(&UserDevice{}).Error
}
