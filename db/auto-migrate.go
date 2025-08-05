package db

import (
	"github.com/minq3010/Backend-React-Native-App/models"
	"gorm.io/gorm"
)

func DBMigrator(db *gorm.DB) error {
	return db.AutoMigrate(&models.Event{}, &models.Ticket{}, &models.User{}, &models.Payment{}, &models.Stats{})
}