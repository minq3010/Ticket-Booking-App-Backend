package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/minq3010/Backend-React-Native-App/models"
	"gorm.io/gorm"
)

type StatRepository struct {
	db *gorm.DB
}

func (r *StatRepository) GetMany(ctx context.Context) ([]*models.Stats, error) {
	stats := []*models.Stats{}

	res := r.db.Model(&models.Stats{}).Order("updated_at desc").Find(&stats)

	if res.Error != nil {
		return nil, res.Error
	}
	return stats, nil
}

func (r *StatRepository) GetOne(ctx context.Context, eventId uint) (*models.Stats, error) {
	stat := &models.Stats{}

	res := r.db.Model(stat).Where("id = ?", eventId).First(stat)

	if res.Error != nil {
		return nil, res.Error
	}
	return stat, nil
}

func (r *StatRepository) UpdateStat(ctx context.Context, eventId uint) error {
	var totalTickets, totalEntered, totalDeleted, revenue int64

	// Count total tickets purchased
	if err := r.db.Model(&models.Ticket{}).
		Where("event_id = ?", eventId).
		Where("deleted_at IS NULL").
		Count(&totalTickets).Error; err != nil {
		// log error
		fmt.Printf("Error counting total tickets for event %d: %v\n", eventId, err)
		return err
	}

	// Count total tickets entered
	if err := r.db.Model(&models.Ticket{}).
		Where("event_id = ? AND entered = true", eventId).
		Where("deleted_at IS NULL").
		Count(&totalEntered).Error; err != nil {
		fmt.Printf("Error counting entered tickets for event %d: %v\n", eventId, err)
		return err
	}

	// Count total tickets deleted
	if err := r.db.Unscoped().Model(&models.Ticket{}).
		Where("event_id = ?", eventId).
		Where("deleted_at IS NOT NULL").
		Count(&totalDeleted).Error; err != nil {
		fmt.Printf("Error counting deleted tickets for event %d: %v\n", eventId, err)
		return err
	}

	// Calculate revenue
	if err := r.db.Table("tickets").
		Select("COALESCE(SUM(events.price), 0)").
		Joins("JOIN events ON tickets.event_id = events.id").
		Where("tickets.event_id = ?", eventId).
		Where("tickets.deleted_at IS NULL").
		Scan(&revenue).Error; err != nil {
		fmt.Printf("Error calculating revenue for event %d: %v\n", eventId, err)
		return err
	}

	var stats models.Stats
	err := r.db.WithContext(ctx).
		Where("event_id = ?", eventId).
		First(&stats).Error

	stats.EventID = eventId
	stats.TotalTicketsPurchased = totalTickets
	stats.TotalTicketsEntered = totalEntered
	stats.TotalTicketsDeleted = totalDeleted
	stats.Revenue = revenue
	stats.UpdatedAt = time.Now()

	switch err {
	case nil:
		// Đã có, update
		return r.db.WithContext(ctx).Save(&stats).Error
	case gorm.ErrRecordNotFound:
		// Chưa có, tạo mới
		return r.db.WithContext(ctx).Create(&stats).Error
	default:
		return err
	}
}

func (r *StatRepository) UpdateAllStats(ctx context.Context) error {
	var events []models.Event

	if err := r.db.Find(&events).Error; err != nil {
		return err
	}

	for _, event := range events {
		if err := r.UpdateStat(ctx, event.ID); err != nil {
			fmt.Printf("Error updating stats for event %d: %v\n", event.ID, err)
		}
	}
	return nil
}

func NewStatRepository(db *gorm.DB) models.StatRepository {
	return &StatRepository{
		db: db,
	}
}

// func GetStatistics(db *gorm.DB) (Statistics, error) {
//     var stats Statistics

//     // Đếm tổng số sự kiện
//     if err := db.Table("events").Count(&stats.TotalEvents).Error; err != nil {
//         return stats, err
//     }

//     // Đếm tổng số vé đã bán (tất cả vé trong bảng tickets)
//     if err := db.Table("tickets").Count(&stats.TotalTickets).Error; err != nil {
//         return stats, err
//     }

//     // Tính tổng doanh thu: SUM(events.price) cho mỗi ticket đã bán
//     if err := db.Table("tickets").
//         Joins("JOIN events ON tickets.event_id = events.id").
//         Select("COALESCE(SUM(events.price),0)").Scan(&stats.TotalRevenue).Error; err != nil {
//         return stats, err
//     }

//     return stats, nil
// }
