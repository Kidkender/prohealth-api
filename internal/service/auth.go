package service

import (
	"context"
	"errors"

	"github.com/kidkender/prohealth-api/internal/dto"
	"github.com/kidkender/prohealth-api/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrUserExists = errors.New("error.user_exists")

type AuthService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
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
		return ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {

		user := model.User{
			Email:    req.Email,
			Password: string(hash),
			Role:     model.RoleUser,
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
		return nil

	})
}

func (s *AuthService) Login(
	ctx context.Context,
	req *dto.LoginRequest,
) (*model.User, error) {
	var user model.User
	err := s.db.Where("email = ?", req.Email).First(&user).Error
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, err
	}

	return &user, nil
}
