package service

import (
	"context"

	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/dto"
	"github.com/kidkender/prohealth-api/internal/model"
	"github.com/kidkender/prohealth-api/internal/slug"
	"gorm.io/gorm"
)

type HospitalService struct {
	db *gorm.DB
}

type HospitalWithRole struct {
	model.Hospital
	Role model.UserRole `gorm:"column:role"`
}

func NewHospitalService(db *gorm.DB) *HospitalService {
	return &HospitalService{db: db}
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
	if err := s.db.WithContext(ctx).Find(&hospitals).Error; err != nil {
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
