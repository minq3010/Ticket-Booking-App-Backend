package db

import (
	"fmt"

	"github.com/gofiber/fiber/v2/log"
	"github.com/minq3010/Backend-React-Native-App/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var dbInstance *gorm.DB 

func Init(config *config.EnvConfig, DBMigrator func(*gorm.DB) error) *gorm.DB {
	uri := fmt.Sprintf(`
		host=%s user=%s dbname=%s password=%s sslmode=%s port=%s`,
		config.DBHost, config.DBUser, config.DBName, config.DBPassword, config.DBSSLMode, config.DBPort,
	)

	db, err := gorm.Open(postgres.Open(uri), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	log.Info("Connected to the database")

	if err := DBMigrator(db); err != nil {
		log.Fatalf("Unable to migrate: %v", err)
	}

	dbInstance = db 
	return db
}


func GetDB() *gorm.DB {
	return dbInstance
}