package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/audit"
	"github.com/kidkender/prohealth-api/internal/model"
	"github.com/kidkender/prohealth-api/internal/security"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const invitationTTL = 7 * 24 * time.Hour

type InvitationService struct {
	db          *gorm.DB
	mailService *MailService
	frontendURL string
}

func NewInvitationService(db *gorm.DB, mailService *MailService, frontendURL string) *InvitationService {
	return &InvitationService{db: db, mailService: mailService, frontendURL: frontendURL}
}

func (s *InvitationService) Invite(ctx context.Context, actorID, hospitalID uint, email string, role model.UserRole) error {
	var existingUser model.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&existingUser).Error; err == nil {
		var memberCount int64
		if err := s.db.WithContext(ctx).Model(&model.HospitalMember{}).
			Where("hospital_id = ? AND user_id = ?", hospitalID, existingUser.ID).
			Count(&memberCount).Error; err != nil {
			return err
		}
		if memberCount > 0 {
			return apperror.ErrHospitalMemberAlreadyExists
		}
	}

	var pendingCount int64
	if err := s.db.WithContext(ctx).Model(&model.HospitalInvitation{}).
		Where("hospital_id = ? AND email = ? AND status = ?", hospitalID, email, model.InvitationPending).
		Count(&pendingCount).Error; err != nil {
		return err
	}
	if pendingCount > 0 {
		return apperror.ErrInvitationAlreadyPending
	}

	rawToken, err := security.GenerateRandomToken()
	if err != nil {
		return err
	}

	invitation := model.HospitalInvitation{
		HospitalID: hospitalID,
		Email:      email,
		Role:       role,
		TokenHash:  security.HashToken(rawToken),
		Status:     model.InvitationPending,
		InvitedBy:  actorID,
		ExpiresAt:  time.Now().Add(invitationTTL),
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&invitation).Error; err != nil {
			return err
		}
		return audit.Record(tx, hospitalID, &actorID, model.ActionCreate, "hospital_invitation", invitation.ID)
	}); err != nil {
		return err
	}

	link := fmt.Sprintf("%s/invitations/accept?token=%s", s.frontendURL, rawToken)
	body := fmt.Sprintf(
		`<p>Bạn được mời tham gia hệ thống ProHealth với vai trò <b>%s</b>.</p>
		 <p><a href="%s">Nhấn vào đây để chấp nhận lời mời</a></p>
		 <p>Lời mời sẽ hết hạn sau 7 ngày.</p>`,
		role, link,
	)
	if err := s.mailService.Send(ctx, email, "Lời mời tham gia ProHealth", body); err != nil {
		return fmt.Errorf("send invitation email: %w", err)
	}

	return nil
}

func (s *InvitationService) ListPending(ctx context.Context, hospitalID uint) ([]model.HospitalInvitation, error) {
	var invitations []model.HospitalInvitation
	err := s.db.WithContext(ctx).
		Where("hospital_id = ? AND status = ?", hospitalID, model.InvitationPending).
		Order("created_at DESC").
		Find(&invitations).Error
	return invitations, err
}

func (s *InvitationService) Revoke(ctx context.Context, actorID, hospitalID, invitationID uint) error {
	var invitation model.HospitalInvitation
	if err := s.db.WithContext(ctx).
		Where("id = ? AND hospital_id = ?", invitationID, hospitalID).
		First(&invitation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrInvitationNotFound
		}
		return err
	}
	if invitation.Status != model.InvitationPending {
		return apperror.ErrInvitationAlreadyProcessed
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&invitation).Update("status", model.InvitationRevoked).Error; err != nil {
			return err
		}
		return audit.Record(tx, hospitalID, &actorID, model.ActionDelete, "hospital_invitation", invitation.ID)
	})
}

func (s *InvitationService) AcceptInvitation(ctx context.Context, token, password, fullname string) error {
	var invitation model.HospitalInvitation
	if err := s.db.WithContext(ctx).
		Where("token_hash = ?", security.HashToken(token)).
		First(&invitation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrInvitationNotFound
		}
		return err
	}

	if invitation.Status != model.InvitationPending {
		return apperror.ErrInvitationAlreadyProcessed
	}
	if time.Now().After(invitation.ExpiresAt) {
		return apperror.ErrInvitationExpired
	}

	var existingCount int64
	if err := s.db.WithContext(ctx).Model(&model.User{}).
		Where("email = ?", invitation.Email).
		Count(&existingCount).Error; err != nil {
		return err
	}
	if existingCount > 0 {
		return apperror.ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user := model.User{Email: invitation.Email, Password: string(hash)}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		profile := model.UserDetail{UserId: user.ID, Fullname: fullname}
		if err := tx.Create(&profile).Error; err != nil {
			return err
		}

		member := model.HospitalMember{HospitalID: invitation.HospitalID, UserID: user.ID, Role: invitation.Role}
		if err := tx.Create(&member).Error; err != nil {
			return err
		}

		if err := tx.Model(&invitation).Update("status", model.InvitationAccepted).Error; err != nil {
			return err
		}

		return audit.Record(tx, invitation.HospitalID, &user.ID, model.ActionCreate, "hospital_member", user.ID)
	})
}
