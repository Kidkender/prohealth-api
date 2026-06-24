package service

import (
	"context"

	"github.com/kidkender/prohealth-api/internal/model"
	"gorm.io/gorm"
)

type AdminService struct {
	db *gorm.DB
}

func NewAdminService(db *gorm.DB) *AdminService {
	return &AdminService{db: db}
}

type HospitalStatResult struct {
	Total      int64 `json:"total"`
	NewLast30d int64 `json:"new_last_30d"`
}

type UserStatResult struct {
	Total      int64 `json:"total"`
	NewLast30d int64 `json:"new_last_30d"`
}

type AppointmentStatResult struct {
	Total     int64 `json:"total"`
	Pending   int64 `json:"pending"`
	Confirmed int64 `json:"confirmed"`
}

type HospitalOverview struct {
	ID                    int    `json:"id"`
	Name                  string `json:"name"`
	MemberCount           int    `json:"member_count"`
	AppointmentsThisMonth int    `json:"appointments_this_month"`
}

type SystemStatResult struct {
	Hospitals         HospitalStatResult    `json:"hospitals"`
	Users             UserStatResult        `json:"users"`
	Appointments      AppointmentStatResult `json:"appointments"`
	HospitalsOverview []HospitalOverview    `json:"hospitals_overview"`
}

func (s *AdminService) GetSystemStat(ctx context.Context) (SystemStatResult, error) {
	var stat SystemStatResult
	s.db.WithContext(ctx).Model(&model.Hospital{}).Count(&stat.Hospitals.Total)
	s.db.WithContext(ctx).Model(&model.Hospital{}).Where("created_at >= NOW() - INTERVAL '30 days'").Count(&stat.Hospitals.NewLast30d)
	s.db.WithContext(ctx).Model(&model.User{}).Count(&stat.Users.Total)
	s.db.WithContext(ctx).Model(&model.User{}).Where("created_at >= NOW() - INTERVAL '30 days'").Count(&stat.Users.NewLast30d)
	s.db.WithContext(ctx).Model(&model.Appointment{}).Count(&stat.Appointments.Total)
	s.db.WithContext(ctx).Model(&model.Appointment{}).Where("status = ?", model.AppointmentPending).Count(&stat.Appointments.Pending)
	s.db.WithContext(ctx).Model(&model.Appointment{}).Where("status = ?", model.AppointmentConfirmed).Count(&stat.Appointments.Confirmed)
	if err := s.db.WithContext(ctx).
		Model(&model.Hospital{}).
		Select(`hospitals.id,
			    hospitals.name,
			     COUNT(DISTINCT hospital_members.user_id) AS member_count,
			     COUNT(DISTINCT appointments.id) AS appointments_this_month
		`).
		Joins("LEFT JOIN hospital_members ON  hospital_members.hospital_id = hospitals.id").
		Joins("LEFT JOIN appointments ON appointments.hospital_id = hospitals.id AND DATE_TRUNC('month', appointments.created_at) = DATE_TRUNC('month', NOW())").
		Group("hospitals.id, hospitals.name").
		Scan(&stat.HospitalsOverview).Error; err != nil {
		return stat, err
	}
	return stat, nil
}
