package models

import (
	"context"
	"time"
)

type Payment struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	OrderID       string    `json:"order_id" gorm:"unique;not null"`         // Mã giao dịch duy nhất
	TicketID      string    `json:"ticket_id" gorm:"default:null"`            // Có thể null nếu chưa tạo vé
	EventID       uint      `json:"event_id" gorm:"not null"`                // Sự kiện liên quan
	UserID        uint      `json:"user_id" gorm:"not null"`                // Người thực hiện thanh toán
	Amount        int       `json:"amount" gorm:"not null"`                // Số tiền thanh toán
	Status        string    `json:"status" gorm:"default:'pending'"`       // pending / success / fail
	PaymentMethod string    `json:"method" gorm:"default:'vnpay'"`         // vnpay / momo / ...
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type PaymentRepository interface {
	Create(ctx context.Context, payment *Payment) error
	GetByOrderID(ctx context.Context, orderID string) (*Payment, error)
	UpdateStatus(ctx context.Context, orderID string, status string) error
	UpdateTicketID(ctx context.Context, orderID string, ticketID string) error
}
