package service

import (
	"context"
	"errors"

	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/dto"
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrUserNotFound
		}
		return nil, err
	}

	if err := s.db.WithContext(ctx).Where("user_id = ?", userId).First(&userDetail).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrUserDetailNotFound
		}
		return nil, err
	}
	return &UserResult{
		ID:   userId,
		Name: userDetail.Fullname,
		Role: string(user.Role),
	}, nil
}

func (s *UserService) AssignRole(ctx context.Context, payload dto.AssignRoleRequest) error {
	var user model.User
	if err := s.db.WithContext(ctx).Where("id = ?", payload.UserID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrUserNotFound
		}
		return err
	}
	user.Role = payload.Role
	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return err
	}
	return nil
}
