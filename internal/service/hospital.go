package service

import (
	"github.com/kidkender/prohealth-api/internal/dto"
	"github.com/kidkender/prohealth-api/internal/model"
	"github.com/kidkender/prohealth-api/internal/slug"
	"gorm.io/gorm"
)

type HospitalService struct {
	db *gorm.DB
}

func NewHospitalService(db *gorm.DB) *HospitalService {
	return &HospitalService{db: db}
}

func (s *HospitalService) CreateHospital(req dto.CreateHospitalRequest) error {
	hospital := model.Hospital{
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		Slug:        slug.From(req.Name),
	}
	return s.db.Create(&hospital).Error
}

func (s *HospitalService) GetHospitalBySlug(slug string) (*model.Hospital, error) {
	var hospital model.Hospital
	if err := s.db.Where("slug = ?", slug).First(&hospital).Error; err != nil {
		return nil, err
	}
	return &hospital, nil
}

func (s *HospitalService) GetHospitalById(id uint) (*model.Hospital, error) {
	var hospital model.Hospital
	if err := s.db.Where("id = ?", id).First(&hospital).Error; err != nil {
		return nil, err
	}
	return &hospital, nil
}

func (s *HospitalService) GetHospitals() ([]model.Hospital, error) {
	var hospitals []model.Hospital
	if err := s.db.Find(&hospitals).Error; err != nil {
		return nil, err
	}
	return hospitals, nil
}

func (s *HospitalService) UpdateHospital(id uint, req dto.UpdateHospitalRequest) error {
	hospital, err := s.GetHospitalById(id)
	if err != nil {
		return err
	}
	if req.Description != "" {
		hospital.Description = req.Description
	}
	if req.Address != "" {
		hospital.Address = req.Address
	}
	return s.db.Save(hospital).Error
}
