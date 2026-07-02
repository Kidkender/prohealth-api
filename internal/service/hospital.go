package service

import (
	"context"
	"time"

	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/cache"
	"github.com/kidkender/prohealth-api/internal/dto"
	"github.com/kidkender/prohealth-api/internal/model"
	"github.com/kidkender/prohealth-api/internal/slug"
	"gorm.io/gorm"
)

type HospitalService struct {
	db    *gorm.DB
	cache *cache.Cache
}

type HospitalWithRole struct {
	model.Hospital
	Role model.UserRole `gorm:"column:role"`
}

func NewHospitalService(db *gorm.DB, cache *cache.Cache) *HospitalService {
	return &HospitalService{db: db, cache: cache}
}

func (s *HospitalService) CreateHospital(ctx context.Context, req dto.CreateHospitalRequest) error {
	generatedSlug := slug.From(req.Name)
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.Hospital{}).
		Where("slug = ?", generatedSlug).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return apperror.ErrHospitalSlugAlreadyExists
	}

	hospital := model.Hospital{
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		Slug:        generatedSlug,
	}
	return s.db.WithContext(ctx).Create(&hospital).Error
}

func (s *HospitalService) GetHospitalBySlug(ctx context.Context, hospitalSlug string) (*model.Hospital, error) {
	var hospital model.Hospital
	if err := s.db.WithContext(ctx).Where("slug = ?", hospitalSlug).First(&hospital).Error; err != nil {
		return nil, err
	}
	return &hospital, nil
}

func (s *HospitalService) GetHospitalById(ctx context.Context, id uint) (*model.Hospital, error) {
	var hospital model.Hospital
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&hospital).Error; err != nil {
		return nil, err
	}
	return &hospital, nil
}

func (s *HospitalService) GetHospitals(ctx context.Context) ([]model.Hospital, error) {
	var hospitals []model.Hospital
	key := cache.ScheduleAllKey()
	if hit, _ := s.cache.Get(ctx, key, &hospitals); hit {
		return hospitals, nil
	}
	if err := s.db.WithContext(ctx).Find(&hospitals).Error; err != nil {
		return nil, err
	}
	if err := s.cache.Set(ctx, key, hospitals, time.Minute); err != nil {
		return nil, err
	}
	return hospitals, nil
}

func (s *HospitalService) UpdateHospital(ctx context.Context, id uint, req dto.UpdateHospitalRequest) error {
	hospital, err := s.GetHospitalById(ctx, id)
	if err != nil {
		return err
	}
	if req.Description != "" {
		hospital.Description = req.Description
	}
	if req.Address != "" {
		hospital.Address = req.Address
	}
	return s.db.WithContext(ctx).Save(hospital).Error
}

type HospitalDoctor struct {
	UserID   uint   `json:"user_id"`
	Fullname string `json:"fullname"`
}

func (s *HospitalService) GetHospitalDoctors(ctx context.Context, hospitalID uint) ([]HospitalDoctor, error) {
	var results []HospitalDoctor
	err := s.db.WithContext(ctx).
		Model(&model.UserDetail{}).
		Select("user_details.user_id, user_details.full_name as fullname").
		Joins("JOIN hospital_members ON hospital_members.user_id = user_details.user_id").
		Where("hospital_members.hospital_id = ? AND hospital_members.role = ?", hospitalID, model.RoleDoctor).
		Scan(&results).Error
	return results, err
}

type HospitalMemberDetail struct {
	UserID   uint   `json:"user_id"`
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

func (s *HospitalService) GetHospitalMembers(ctx context.Context, hospitalID uint) ([]HospitalMemberDetail, error) {
	var results []HospitalMemberDetail
	err := s.db.WithContext(ctx).
		Model(&model.HospitalMember{}).
		Select("hospital_members.user_id, user_details.full_name as fullname, users.email, hospital_members.role").
		Joins("JOIN users ON users.id = hospital_members.user_id").
		Joins("JOIN user_details ON user_details.user_id = hospital_members.user_id").
		Where("hospital_members.hospital_id = ?", hospitalID).
		Scan(&results).Error
	return results, err
}

func (s *HospitalService) GetMemberRole(ctx context.Context, hospitalID, userID uint) (string, error) {
	key := cache.MemberRoleKey(hospitalID, userID)
	var role model.UserRole
	if hit, _ := s.cache.Get(ctx, key, &role); hit {
		return string(role), nil
	}

	var member model.HospitalMember
	if err := s.db.WithContext(ctx).
		Where("hospital_id = ? AND user_id = ?", hospitalID, userID).
		First(&member).Error; err != nil {
		return "", err
	}

	s.cache.Set(ctx, key, member.Role, 5*time.Minute)
	return string(member.Role), nil
}

func (s *HospitalService) AddMemberByEmail(ctx context.Context, hospitalID uint, email string, role model.UserRole) error {
	var user model.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return apperror.ErrUserNotFound
	}
	member := model.HospitalMember{HospitalID: hospitalID, UserID: user.ID, Role: role}
	err := s.db.WithContext(ctx).
		Where(model.HospitalMember{HospitalID: hospitalID, UserID: user.ID}).
		FirstOrCreate(&member).Error

	if err != nil {
		s.cache.Delete(ctx, cache.MemberRoleKey(hospitalID, user.ID))
	}
	return err

}

func (s *HospitalService) RemoveMember(ctx context.Context, hospitalID uint, userID uint) error {
	err := s.db.WithContext(ctx).
		Where("hospital_id = ? AND user_id = ?", hospitalID, userID).
		Delete(&model.HospitalMember{}).Error
	if err != nil {
		s.cache.Delete(ctx, cache.MemberRoleKey(hospitalID, userID))
	}
	return err
}

func (s *HospitalService) GetHospitalByUser(ctx context.Context, userID uint) ([]HospitalWithRole, error) {
	var results []HospitalWithRole
	err := s.db.WithContext(ctx).
		Model(&model.Hospital{}).
		Select("hospitals.*, hospital_members.role").
		Joins("JOIN hospital_members ON hospital_members.hospital_id = hospitals.id").
		Where("hospital_members.user_id = ?", userID).
		Scan(&results).Error
	return results, err
}
