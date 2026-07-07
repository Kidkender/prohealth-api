package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/audit"
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

func (s *HospitalService) CreateHospital(ctx context.Context, requesterID uint, req dto.CreateHospitalRequest) error {
	generatedSlug := slug.From(req.Name)
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.Hospital{}).
		Where("slug = ?", generatedSlug).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return apperror.ErrHospitalSlugAlreadyExists
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		hospital := model.Hospital{
			Name:        req.Name,
			Description: req.Description,
			Address:     req.Address,
			Slug:        generatedSlug,
			Status:      model.HospitalStatusPending,
		}
		if err := tx.Create(&hospital).Error; err != nil {
			return err
		}

		member := model.HospitalMember{
			HospitalID: hospital.ID,
			UserID:     requesterID,
			Role:       model.RoleAdmin,
		}
		if err := tx.Create(&member).Error; err != nil {
			return err
		}

		return audit.Record(tx, hospital.ID, &requesterID, model.ActionCreate, "hospital", hospital.ID)
	})
}

func (s *HospitalService) UpdateHospitalStatus(ctx context.Context, actorID uint, id uint, status model.HospitalStatus) error {
	hospital, err := s.GetHospitalById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrHospitalNotFound
		}
		return err
	}

	hospital.Status = status
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(hospital).Error; err != nil {
			return err
		}
		return audit.Record(tx, id, &actorID, model.ActionUpdate, "hospital_status", id)
	})
}

func (s *HospitalService) GetHospitalsByStatus(ctx context.Context, status *model.HospitalStatus) ([]model.Hospital, error) {
	query := s.db.WithContext(ctx).Model(&model.Hospital{})
	if status != nil {
		query = query.Where("hospital_status = ?", *status)
	}
	var hospitals []model.Hospital
	if err := query.Order("created_at DESC").Find(&hospitals).Error; err != nil {
		return nil, err
	}
	return hospitals, nil
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
	if err := s.db.WithContext(ctx).Where("hospital_status = ?", model.HospitalStatusApproved).Find(&hospitals).Error; err != nil {
		return nil, err
	}
	if err := s.cache.Set(ctx, key, hospitals, time.Minute); err != nil {
		return nil, err
	}
	return hospitals, nil
}

func (s *HospitalService) UpdateHospital(ctx context.Context, actorID uint, id uint, req dto.UpdateHospitalRequest) error {
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
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(hospital).Error; err != nil {
			return err
		}
		return audit.Record(tx, id, &actorID, model.ActionUpdate, "hospital", id)
	})
}

func (s *HospitalService) DeleteHospital(ctx context.Context, actorID uint, id uint) error {
	if _, err := s.GetHospitalById(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrHospitalNotFound
		}
		return err
	}

	var dependents int64
	if err := s.db.WithContext(ctx).Model(&model.Appointment{}).
		Where("hospital_id = ?", id).Count(&dependents).Error; err != nil {
		return err
	}
	if dependents > 0 {
		return apperror.ErrHospitalHasDependents
	}

	if err := s.db.WithContext(ctx).Model(&model.SchedulePeriod{}).
		Where("hospital_id = ?", id).Count(&dependents).Error; err != nil {
		return err
	}
	if dependents > 0 {
		return apperror.ErrHospitalHasDependents
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&model.Hospital{}, id).Error; err != nil {
			return err
		}
		return audit.Record(tx, id, &actorID, model.ActionDelete, "hospital", id)
	})
}

type HospitalStatsResult struct {
	Appointments     AppointmentStatResult `json:"appointments"`
	HospitalOverview HospitalOverview      `json:"hospital_overview"`
}

func (s *HospitalService) GetHospitalStats(ctx context.Context, hospitalID uint) (HospitalStatsResult, error) {
	var stat HospitalStatsResult

	s.db.WithContext(ctx).Model(&model.Appointment{}).Where("hospital_id = ?", hospitalID).Count(&stat.Appointments.Total)
	s.db.WithContext(ctx).Model(&model.Appointment{}).Where("status = ? AND hospital_id = ?", model.AppointmentPending, hospitalID).Count(&stat.Appointments.Pending)
	s.db.WithContext(ctx).Model(&model.Appointment{}).Where("status = ? AND hospital_id = ?", model.AppointmentConfirmed, hospitalID).Count(&stat.Appointments.Confirmed)
	if err := s.db.WithContext(ctx).
		Model(&model.Hospital{}).
		Select(`hospitals.id,
				    hospitals.name,
				     COUNT(DISTINCT hospital_members.user_id) AS member_count,
				     COUNT(DISTINCT appointments.id) AS appointments_this_month
			`).
		Joins("LEFT JOIN hospital_members ON  hospital_members.hospital_id = hospitals.id").
		Joins("LEFT JOIN appointments ON appointments.hospital_id = hospitals.id AND DATE_TRUNC('month', appointments.created_at) = DATE_TRUNC('month', NOW())").
		Where("hospitals.id = ?", hospitalID).
		Group("hospitals.id, hospitals.name").
		Scan(&stat.HospitalOverview).Error; err != nil {
		return stat, err
	}
	return stat, nil
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

func (s *HospitalService) AddMemberByEmail(ctx context.Context, actorID uint, hospitalID uint, email string, role model.UserRole) error {
	var user model.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return apperror.ErrUserNotFound
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		member := model.HospitalMember{HospitalID: hospitalID, UserID: user.ID, Role: role}
		if err := tx.
			Where(model.HospitalMember{HospitalID: hospitalID, UserID: user.ID}).
			FirstOrCreate(&member).Error; err != nil {
			return err
		}
		return audit.Record(tx, hospitalID, &actorID, model.ActionCreate, "hospital_member", user.ID)
	})

	if err != nil {
		s.cache.Delete(ctx, cache.MemberRoleKey(hospitalID, user.ID))
	}
	return err
}

func (s *HospitalService) RemoveMember(ctx context.Context, actorID uint, hospitalID uint, userID uint) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Where("hospital_id = ? AND user_id = ?", hospitalID, userID).
			Delete(&model.HospitalMember{}).Error; err != nil {
			return err
		}
		return audit.Record(tx, hospitalID, &actorID, model.ActionDelete, "hospital_member", userID)
	})
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

const (
	TrendIntervalWeek  = "week"
	TrendIntervalMonth = "month"
)

type AppointmentTrendPoint struct {
	PeriodStart time.Time `json:"period_start"`
	Total       int64     `json:"total"`
	Cancelled   int64     `json:"cancelled"`
	NoShowRate  float64   `json:"no_show_rate"`
}

type TopDoctor struct {
	UserID           uint   `json:"user_id"`
	Fullname         string `json:"fullname"`
	AppointmentCount int64  `json:"appointment_count"`
}

type AppointmentTrendResult struct {
	Interval   string                  `json:"interval"`
	Trends     []AppointmentTrendPoint `json:"trends"`
	TopDoctors []TopDoctor             `json:"top_doctors"`
}

func (s *HospitalService) GetAppointmentTrends(ctx context.Context, hospitalID uint, interval string, topDoctorLimit int) (AppointmentTrendResult, error) {
	result := AppointmentTrendResult{Interval: interval}

	if interval != TrendIntervalWeek && interval != TrendIntervalMonth {
		return result, apperror.ErrInvalidTrendInterval
	}
	if topDoctorLimit <= 0 {
		topDoctorLimit = 5
	}

	if err := s.db.WithContext(ctx).
		Model(&model.Appointment{}).
		Select(fmt.Sprintf(`DATE_TRUNC('%s', schedule_shifts.date) AS period_start,
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE appointments.status = ?) AS cancelled`, interval), model.AppointmentCancelled).
		Joins("JOIN schedule_shifts ON schedule_shifts.id = appointments.shift_id").
		Where("appointments.hospital_id = ?", hospitalID).
		Group("period_start").
		Order("period_start").
		Scan(&result.Trends).Error; err != nil {
		return result, err
	}

	for i := range result.Trends {
		if result.Trends[i].Total > 0 {
			result.Trends[i].NoShowRate = float64(result.Trends[i].Cancelled) / float64(result.Trends[i].Total)
		}
	}

	if err := s.db.WithContext(ctx).
		Model(&model.Appointment{}).
		Select(`user_details.user_id, user_details.full_name AS fullname, COUNT(DISTINCT appointments.id) AS appointment_count`).
		Joins("JOIN schedule_shifts ON schedule_shifts.id = appointments.shift_id").
		Joins("JOIN schedule_shift_doctors ON schedule_shift_doctors.shift_id = schedule_shifts.id").
		Joins("JOIN user_details ON user_details.user_id = schedule_shift_doctors.user_id").
		Where("appointments.hospital_id = ?", hospitalID).
		Group("user_details.user_id, user_details.full_name").
		Order("appointment_count DESC").
		Limit(topDoctorLimit).
		Scan(&result.TopDoctors).Error; err != nil {
		return result, err
	}

	return result, nil
}

type DepartmentPerformance struct {
	DepartmentID   uint   `json:"department_id"`
	DepartmentName string `json:"department_name"`
	Total          int64  `json:"total"`
	Completed      int64  `json:"completed"`
	Cancelled      int64  `json:"cancelled"`
}

func (s *HospitalService) GetDepartmentPerformance(ctx context.Context, hospitalID uint) ([]DepartmentPerformance, error) {
	var results []DepartmentPerformance
	if err := s.db.WithContext(ctx).
		Model(&model.Appointment{}).
		Select(`schedule_shifts.department_id AS department_id,
			departments.name AS department_name,
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE appointments.status = ?) AS completed,
			COUNT(*) FILTER (WHERE appointments.status = ?) AS cancelled`,
			model.AppointmentCompleted, model.AppointmentCancelled).
		Joins("JOIN schedule_shifts ON schedule_shifts.id = appointments.shift_id").
		Joins("JOIN departments ON departments.id = schedule_shifts.department_id").
		Where("appointments.hospital_id = ?", hospitalID).
		Group("schedule_shifts.department_id, departments.name").
		Order("total DESC").
		Scan(&results).Error; err != nil {
		return nil, err
	}

	slog.Info("result: ", "results", results)
	return results, nil
}
