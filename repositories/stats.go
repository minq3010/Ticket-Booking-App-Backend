package repositories

import (
    "gorm.io/gorm"
)

type Statistics struct {
    TotalEvents   int64
    TotalTickets  int64
    TotalRevenue  float64
}

func GetStatistics(db *gorm.DB) (Statistics, error) {
    var stats Statistics

    // Đếm tổng số sự kiện
    if err := db.Table("events").Count(&stats.TotalEvents).Error; err != nil {
        return stats, err
    }

    // Đếm tổng số vé đã bán (tất cả vé trong bảng tickets)
    if err := db.Table("tickets").Count(&stats.TotalTickets).Error; err != nil {
        return stats, err
    }

    // Tính tổng doanh thu: SUM(events.price) cho mỗi ticket đã bán
    if err := db.Table("tickets").
        Joins("JOIN events ON tickets.event_id = events.id").
        Select("COALESCE(SUM(events.price),0)").Scan(&stats.TotalRevenue).Error; err != nil {
        return stats, err
    }

    return stats, nil
}
