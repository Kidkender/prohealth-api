package service

import (
	"context"

	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/model"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

type UserResult struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

func (s *UserService) GetUser(ctx context.Context, userId uint) (*UserResult, error) {
	var userDetail model.UserDetail
	var user model.User

	if err := s.db.WithContext(ctx).Where("id = ?", userId).First(&user).Error; err != nil {
		return nil, apperror.ErrUserNotFound
	}

	if err := s.db.WithContext(ctx).Where("user_id = ?", userId).First(&userDetail).Error; err != nil {
		return nil, apperror.ErrUserDetailNotFound
	}
	return &UserResult{
		ID:   userId,
		Name: userDetail.Fullname,
		Role: string(user.Role),
	}, nil
}
