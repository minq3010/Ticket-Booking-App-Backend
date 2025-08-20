package repositories

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/minq3010/Backend-React-Native-App/models"
	"gorm.io/gorm"
)

type EventRepository struct {
	db *gorm.DB
}

func (r *EventRepository) GetMany(ctx context.Context) ([]*models.Event, error) {
	events := []*models.Event{}

	res := r.db.Model(&models.Event{}).Order("updated_at desc").Find(&events)

	if res.Error != nil {
		return nil, res.Error
	}

	return events, nil
}

func (r *EventRepository) SearchByName(ctx context.Context, name string) ([]*models.Event, error) {
	var events []*models.Event

	err := r.db.WithContext(ctx).Where("name ILIKE ?", "%"+name+"%").Find(&events).Error

	return events, err
}

func (r *EventRepository) GetOne(ctx context.Context, eventId uint) (*models.Event, error) {
	event := &models.Event{}

	res := r.db.Model(event).Where("id = ?", eventId).First(event)

	if res.Error != nil {
		return nil, res.Error
	}

	return event, nil
}

func (r *EventRepository) CreateOne(ctx context.Context, event *models.Event) (*models.Event, error) {
	res := r.db.Model(event).Create(event)

	if res.Error != nil {
		return nil, res.Error
	}

	return event, nil
}

func (r *EventRepository) UpdateOne(ctx context.Context, eventId uint, updateData map[string]interface{}) (*models.Event, error) {
	event := &models.Event{}

	updateRes := r.db.Model(event).Where("id = ?", eventId).Updates(updateData)

	if updateRes.Error != nil {
		return nil, updateRes.Error
	}

	getRes := r.db.Model(event).Where("id = ?", eventId).First(event)

	if getRes.Error != nil {
		return nil, getRes.Error
	}

	return event, nil
}

func (r *EventRepository) DeleteOne(ctx context.Context, eventId uint) error {
	var event models.Event

	if err := r.db.Select("created_at").First(&event, eventId).Error; err != nil {
		return err
	}

	var count int64
	if err := r.db.Model(&models.Ticket{}).
		Where("event_id = ? AND entered = true", eventId).
		Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		if time.Since(event.CreatedAt) < 24*time.Hour {
			return errors.New("cannot delete event within 1 day if tickets have already been confirmed")
		}
	}

	if err := r.db.Delete(&models.Event{}, eventId).Error; err != nil {
		return err
	}

	return nil
}

func (r *EventRepository) DeleteExpiredEvents(ctx context.Context) error {
	cutoffTime := time.Now().AddDate(0, 0, -2)

	result := r.db.WithContext(ctx).Where("date < ?", cutoffTime).Delete(&models.Event{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		log.Printf("List Expired Events %d", result.RowsAffected)
	}
	return nil
}

func NewEventRepository(db *gorm.DB) models.EventRepository {
	return &EventRepository{
		db: db,
	}
}
