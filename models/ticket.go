package models

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Ticket struct {
	ID      uint  `json:"id" gorm:"primarykey"`
	EventID uint  `json:"eventId"`
	UserID  uint  `json:"userId" gorm:"foreignkey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Event   Event `json:"event" gorm:"foreignkey:EventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Entered bool  `json:"entered" default:"false"`
	// Price		int			`json:"price"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

type TicketRepository interface {
	GetMany(ctx context.Context, userId uint) ([]*Ticket, error)
	GetOne(ctx context.Context, userId uint, ticketId uint) (*Ticket, error)
	CreateOne(ctx context.Context, userId uint, ticket *Ticket) (*Ticket, error)
	UpdateOne(ctx context.Context, userId uint, ticketId uint, updateData map[string]interface{}) (*Ticket, error)
	DeleteOne(ctx context.Context, ticketId uint) error
}

type ValidateTicket struct {
	TicketId uint `json:"ticketId"`
	OwnerId  uint `json:"ownerId"`
}
