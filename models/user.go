package models

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	Manager  UserRole = "manager"
	Attendee UserRole = "attendee"
)

type User struct {
	ID        uint      `json:"id"        gorm:"primarykey"`
	Email     string    `json:"email"     gorm:"text;not null"`
	Role      UserRole  `json:"role"      gorm:"text;default:attendee"`
	Avatar    string    `json:"avatar"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type UserRepository interface {
	GetUserInfo(ctx context.Context, userId uint) (*User, error)
	UpdateUserInfo(
		ctx context.Context,
		userId uint,
		updateData map[string]interface{},
	) (*User, error)
	SearchUserAccountByEmail(ctx context.Context, email string) ([]*User, error)
	GetAllUsers(ctx context.Context) ([]*User, error)
	DeleteUser(ctx context.Context, userId uint) error  
}

func (u *User) AfterCreate(db *gorm.DB) (err error) {
	if u.ID == 1 {
		db.Model(u).Update("role", Manager)
	}
	return
}
