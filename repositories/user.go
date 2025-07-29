package repositories

import (
	"context"

	"github.com/minq3010/Backend-React-Native-App/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func (r *UserRepository) GetUserInfo(ctx context.Context, userId uint) (*models.User, error) {
	user := &models.User{}

	res := r.db.Model(user).Where("id = ?", userId).First(user)

	if res.Error != nil {
		return nil, res.Error
	}

	return user, nil
}

func (r *UserRepository) UpdateUserInfo(ctx context.Context, userId uint, updateData map[string]interface{}) (*models.User, error) {
	user := &models.User{}

	res := r.db.Model(user).Where("id = ?", userId).Updates(updateData)

	if res.Error != nil {
		return nil, res.Error
	}

	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return user, nil
}

func NewUserRepository(db *gorm.DB) models.UserRepository {
	return &UserRepository{
		db: db,
	}
}