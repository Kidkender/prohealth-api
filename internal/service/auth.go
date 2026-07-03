package service

import (
	"context"
	"log/slog"
	"time"

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
	UserId       uint   `json:"user_id"`
	Token        string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
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

		return nil
	})
}

func (s *AuthService) Login(
	ctx context.Context,
	req *dto.LoginInput,
	ip string,
	userAgent string,
) (*AuthResult, error) {
	var user model.User

	err := s.db.Where("email = ?", req.Email).First(&user).Error
	if err != nil {
		return nil, apperror.ErrUserNotFound
	}
	slog.Info("user found", "email", req.Email)

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, apperror.ErrInvalidCredentials
	}

	slog.Info("system role", "role", user.SystemRole)
	claims := security.NewClaims(user.ID, user.SystemRole)
	accessToken, err := security.GenerateToken(claims)

	if err != nil {
		return nil, err
	}

	refreshToken, err := security.GenerateRandomToken()
	if err != nil {
		return nil, err
	}

	session := model.Session{
		UserID:     user.ID,
		TokenHash:  security.HashToken(refreshToken),
		ExpiresAt:  time.Now().Add(7 * time.Hour),
		IP:         ip,
		UserAgent:  userAgent,
		LastUsedAt: time.Now(),
	}
	if err := s.db.WithContext(ctx).Create(&session).Error; err != nil {
		return nil, err
	}

	return &AuthResult{
		UserId:       user.ID,
		Token:        accessToken,
		RefreshToken: refreshToken,
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

func (s *AuthService) RefreshToken(
	ctx context.Context,
	refreshToken string,
) (*AuthResult, error) {
	var session model.Session

	if err := s.db.WithContext(ctx).Where("token_hash = ?", security.HashToken(refreshToken)).First(&session).Error; err != nil {
		return nil, apperror.ErrInvalidToken
	}

	if session.Revoked {
		s.db.WithContext(ctx).Where("user_id = ?", session.UserID).Update("revoked", true)
		return nil, apperror.ErrTokenReused
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, apperror.ErrTokenExpired
	}

	var user model.User
	if err := s.db.First(&user, session.UserID).Error; err != nil {
		return nil, err
	}
	s.db.Model(&session).Update("revoked", true)
	newRawToken, _ := security.GenerateRandomToken()
	newSession := model.Session{
		UserID:     session.UserID,
		TokenHash:  security.HashToken(newRawToken),
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
		IP:         session.IP,
		UserAgent:  session.UserAgent,
		LastUsedAt: time.Now(),
	}
	s.db.Create(&newSession)
	claims := security.NewClaims(user.ID, user.SystemRole)
	accessToken, _ := security.GenerateToken(claims)

	return &AuthResult{
		UserId:       user.ID,
		Token:        accessToken,
		RefreshToken: newRawToken,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	hash := security.HashToken(rawToken)
	return s.db.WithContext(ctx).Where("token_hash = ?", hash).Update("revoked", true).Error
}
