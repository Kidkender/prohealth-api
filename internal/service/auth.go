package service

import (
	"context"

	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/dto"
	"github.com/kidkender/prohealth-api/internal/model"
	"github.com/kidkender/prohealth-api/internal/security"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

type AuthResult struct {
	UserId uint   `json:"user_id"`
	Token  string `json:"token"`
}

func (s *AuthService) Register(
	ctx context.Context,
	req *dto.RegisterRequest,
) error {
	var count int64
	err := s.db.Model(&model.User{}).
		Where("email = ?", req.Email).
		Count(&count).Error
	if err != nil {
		return err
	}

	if count > 0 {
		return apperror.ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {

		hospital := model.Hospital{}
		if err = tx.First(&hospital, req.HospitalID).Error; err != nil {
			return err
		}

		user := model.User{
			Email:    req.Email,
			Password: string(hash),
		}
		if err = tx.Create(&user).Error; err != nil {
			return err
		}

		profile := model.UserDetail{
			UserId:   user.ID,
			Fullname: req.Fullname,
		}
		if err = tx.Create(&profile).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.HospitalMember{
			HospitalID: req.HospitalID,
			UserID:     user.ID,
			Role:       model.RoleDoctor,
		}).Error; err != nil {
			return err
		}

		return nil

	})
}

func (s *AuthService) Login(
	ctx context.Context,
	req *dto.LoginInput,
) (*AuthResult, error) {
	var user model.User
	var member model.HospitalMember

	err := s.db.Where("email = ?", req.Email).First(&user).Error
	if err != nil {
		return nil, apperror.ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, apperror.ErrInvalidCredentials
	}

	if err := s.db.Where("hospital_id = ? AND user_id = ?", req.HospitalID, user.ID).
		First(&member).Error; err != nil {
		return nil, apperror.ErrUnauthorized
	}

	claims := security.NewClaims(user.ID, member.Role)
	token, err := security.GenerateToken(claims)

	if err != nil {
		return nil, err
	}

	return &AuthResult{
		UserId: user.ID,
		Token:  token,
	}, nil
}

func (s *AuthService) ChangePassword(
	ctx context.Context,
	userID uint,
	req *dto.ChangePasswordInput,
) error {
	var user model.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return apperror.ErrInvalidCredentials
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hash)
	if err := s.db.Save(&user).Error; err != nil {
		return err
	}

	return nil
}
