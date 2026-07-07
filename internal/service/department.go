package service

import (
	"context"
	"errors"

	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/dto"
	"github.com/kidkender/prohealth-api/internal/model"
	"gorm.io/gorm"
)

type DepartmentService struct {
	db *gorm.DB
}

func NewDepartmentService(db *gorm.DB) *DepartmentService {
	return &DepartmentService{db: db}
}

func (s *DepartmentService) CreateDepartment(ctx context.Context, req dto.CreateDepartmentRequest) (*model.Department, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.Department{}).
		Where("name = ?", req.Name).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, apperror.ErrDepartmentAlreadyExists
	}

	department := model.Department{Name: req.Name}
	if err := s.db.WithContext(ctx).Create(&department).Error; err != nil {
		return nil, err
	}
	return &department, nil
}

func (s *DepartmentService) GetDepartments(ctx context.Context) ([]model.Department, error) {
	var departments []model.Department
	if err := s.db.WithContext(ctx).Order("name ASC").Find(&departments).Error; err != nil {
		return nil, err
	}
	return departments, nil
}

func (s *DepartmentService) GetDepartmentByID(ctx context.Context, id uint) (*model.Department, error) {
	var department model.Department
	if err := s.db.WithContext(ctx).First(&department, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrDepartmentNotFound
		}
		return nil, err
	}
	return &department, nil
}

func (s *DepartmentService) UpdateDepartment(ctx context.Context, id uint, req dto.UpdateDepartmentRequest) (*model.Department, error) {
	department, err := s.GetDepartmentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != department.Name {
		var count int64
		if err := s.db.WithContext(ctx).Model(&model.Department{}).
			Where("name = ? AND id != ?", req.Name, id).Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, apperror.ErrDepartmentAlreadyExists
		}
	}

	department.Name = req.Name
	if err := s.db.WithContext(ctx).Save(department).Error; err != nil {
		return nil, err
	}
	return department, nil
}

func (s *DepartmentService) DeleteDepartment(ctx context.Context, id uint) error {
	if _, err := s.GetDepartmentByID(ctx, id); err != nil {
		return err
	}

	var inUse int64
	if err := s.db.WithContext(ctx).Model(&model.UserDetail{}).
		Where("department_id = ?", id).Count(&inUse).Error; err != nil {
		return err
	}
	if inUse > 0 {
		return apperror.ErrDepartmentInUse
	}

	if err := s.db.WithContext(ctx).Model(&model.ScheduleShift{}).
		Where("department_id = ?", id).Count(&inUse).Error; err != nil {
		return err
	}
	if inUse > 0 {
		return apperror.ErrDepartmentInUse
	}

	return s.db.WithContext(ctx).Delete(&model.Department{}, id).Error
}
