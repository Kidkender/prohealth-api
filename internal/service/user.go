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
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Role       string `json:"role"`
	HospitalID uint   `json:"hospital_id"`
}

func (s *UserService) GetUser(ctx context.Context, userId uint, hospitalID uint) (*UserResult, error) {
	var userDetail model.UserDetail

	if err := s.db.WithContext(ctx).Where("user_id = ?", userId).First(&userDetail).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrUserDetailNotFound
		}
		return nil, err
	}

	if hospitalID == 0 {
		return &UserResult{
			ID:   userId,
			Name: userDetail.Fullname,
		}, nil
	}

	var member model.HospitalMember
	if err := s.db.WithContext(ctx).Where("user_id = ? AND hospital_id = ?", userId, hospitalID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrHospitalMemberNotFound
		}
		return nil, err
	}

	return &UserResult{
		ID:         userId,
		Name:       userDetail.Fullname,
		Role:       string(member.Role),
		HospitalID: member.HospitalID,
	}, nil
}

func (s *UserService) AssignRole(ctx context.Context, hospitalID uint, payload dto.AssignRoleRequest) error {
	var member model.HospitalMember
	if err := s.db.WithContext(ctx).Where("user_id = ? AND hospital_id = ?", payload.UserID, hospitalID).
		First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrHospitalMemberNotFound
		}
		return err
	}

	member.Role = payload.Role
	if err := s.db.WithContext(ctx).Save(&member).Error; err != nil {
		return err
	}
	return nil
}
