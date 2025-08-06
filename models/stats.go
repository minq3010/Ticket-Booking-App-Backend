package models

import (
	"context"
	"time"
)

type Stats struct {
	ID                    uint      `json:"id" gorm:"primaryKey"`
	EventID               uint      `json:"eventId" gorm:"uniqueIndex"`
	TotalTicketsPurchased int64     `json:"totalTicketsSold"`
	TotalTicketsEntered   int64     `json:"totalTicketsEntered"`
	TotalTicketsDeleted   int64     `json:"totalTicketsDeleted"`
	Revenue               int64     `json:"revenue"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

type StatRepository interface {
	GetMany(ctx context.Context) ([]*Stats, error)
	GetOne(ctx context.Context, eventId uint) (*Stats, error)
	UpdateStat(ctx context.Context, eventId uint) error
	UpdateAllStats(ctx context.Context) error
}