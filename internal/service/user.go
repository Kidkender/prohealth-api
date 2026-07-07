package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/audit"
	"github.com/kidkender/prohealth-api/internal/dto"
	"github.com/kidkender/prohealth-api/internal/model"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

type UserResult struct {
	ID             uint   `json:"id"`
	Email          string `json:"email"`
	Status         string `json:"status"`
	Name           string `json:"name"`
	Role           string `json:"role"`
	HospitalID     uint   `json:"hospital_id"`
	DepartmentName string `json:"department_name"`
	Address        string `json:"address"`
	PhoneNumber    string `json:"phone_number"`
}

func (s *UserService) GetUser(ctx context.Context, userId uint, hospitalID uint) (*UserResult, error) {
	var userDetail model.UserDetail

	if err := s.db.WithContext(ctx).Where("user_id = ?", userId).First(&userDetail).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrUserDetailNotFound
		}
		return nil, err
	}

	if hospitalID == 0 {
		return &UserResult{
			ID:   userId,
			Name: userDetail.Fullname,
		}, nil
	}

	var member model.HospitalMember
	if err := s.db.WithContext(ctx).Where("user_id = ? AND hospital_id = ?", userId, hospitalID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrHospitalMemberNotFound
		}
		return nil, err
	}

	return &UserResult{
		ID:         userId,
		Name:       userDetail.Fullname,
		Role:       string(member.Role),
		HospitalID: member.HospitalID,
	}, nil
}

func (s *UserService) AssignRole(ctx context.Context, actorID uint, hospitalID uint, payload dto.AssignRoleRequest) error {
	var member model.HospitalMember
	if err := s.db.WithContext(ctx).Where("user_id = ? AND hospital_id = ?", payload.UserID, hospitalID).
		First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrHospitalMemberNotFound
		}
		return err
	}

	member.Role = payload.Role
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&member).Error; err != nil {
			return err
		}
		return audit.Record(tx, hospitalID, &actorID, model.ActionUpdate, "hospital_member_role", payload.UserID)
	})
}

func (s *UserService) UpdateStatus(ctx context.Context, actorID uint, targetUserID uint, status model.UserStatus) error {
	if actorID == targetUserID {
		return apperror.ErrCannotBanSelf
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.First(&user, targetUserID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperror.ErrUserNotFound
			}
			return err
		}

		user.UserStatus = status
		if err := tx.Save(&user).Error; err != nil {
			return err
		}

		if status != model.UserStatusActive {
			if err := tx.Model(&model.Session{}).
				Where("user_id = ? AND revoked = false", targetUserID).
				Update("revoked", true).Error; err != nil {
				return err
			}
		}

		return audit.Record(tx, 0, &actorID, model.ActionUpdate, "user_status", targetUserID)
	})
}

type UserHospitalMembership struct {
	HospitalID   uint      `json:"hospital_id"`
	HospitalName string    `json:"hospital_name"`
	Role         string    `json:"role"`
	JoinedAt     time.Time `json:"joined_at"`
}

type UserDetailResult struct {
	ID             uint                     `json:"id"`
	Email          string                   `json:"email"`
	SystemRole     string                   `json:"system_role"`
	Status         string                   `json:"status"`
	Fullname       string                   `json:"fullname"`
	Address        string                   `json:"address"`
	PhoneNumber    string                   `json:"phone_number"`
	DepartmentName string                   `json:"department_name,omitempty"`
	Hospitals      []UserHospitalMembership `json:"hospitals"`
	CreatedAt      time.Time                `json:"created_at"`
}

func (s *UserService) GetUserDetail(ctx context.Context, userID uint) (*UserDetailResult, error) {
	var user model.User
	if err := s.db.WithContext(ctx).
		Preload("Details.Department").
		First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrUserNotFound
		}
		return nil, err
	}

	var memberships []model.HospitalMember
	if err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("Hospital").
		Find(&memberships).Error; err != nil {
		return nil, err
	}

	result := &UserDetailResult{
		ID:         user.ID,
		Email:      user.Email,
		SystemRole: string(user.SystemRole),
		Status:     string(user.UserStatus),
		CreatedAt:  user.CreatedAt,
	}

	if details := user.Details; details != nil {
		result.Fullname = details.Fullname
		result.Address = details.Address
		result.PhoneNumber = details.PhoneNumber
		if details.Department != nil {
			result.DepartmentName = details.Department.Name
		}
	}

	for _, member := range memberships {
		result.Hospitals = append(result.Hospitals, UserHospitalMembership{
			HospitalID:   member.HospitalID,
			HospitalName: member.Hospital.Name,
			Role:         string(member.Role),
			JoinedAt:     member.JoinedAt,
		})
	}

	return result, nil
}

func (s *UserService) GetUsers(ctx context.Context, filter dto.FilterUser) ([]UserResult, int64, error) {
	idQuery := s.db.WithContext(ctx).Model(&model.User{}).
		Distinct("users.id").
		Joins("LEFT JOIN user_details ON user_details.user_id = users.id")

	if filter.HospitalID != nil || filter.Role != nil {
		idQuery = idQuery.Joins("JOIN hospital_members ON hospital_members.user_id = users.id")
		if filter.HospitalID != nil {
			idQuery = idQuery.Where("hospital_members.hospital_id = ?", *filter.HospitalID)
		}
		if filter.Role != nil {
			idQuery = idQuery.Where("hospital_members.role = ?", *filter.Role)
		}
	}

	if filter.DepartmentID != nil {
		idQuery = idQuery.Where("user_details.department_id = ?", *filter.DepartmentID)
	}

	if filter.Status != nil {
		idQuery = idQuery.Where("users.user_status = ?", *filter.Status)
	}

	if filter.Search != nil {
		if search := strings.TrimSpace(*filter.Search); search != "" {
			like := "%" + search + "%"
			idQuery = idQuery.Where("users.email ILIKE ? OR user_details.full_name ILIKE ?", like, like)
		}
	}

	var count int64
	if err := idQuery.Session(&gorm.Session{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	var userIDs []uint
	if err := idQuery.
		Order("users.id").
		Limit(filter.Limit).
		Offset(filter.Offset()).
		Pluck("users.id", &userIDs).
		Error; err != nil {
		return nil, 0, err
	}

	if len(userIDs) == 0 {
		return []UserResult{}, count, nil
	}

	var users []model.User
	if err := s.db.WithContext(ctx).
		Preload("Details.Department").
		Where("id IN ?", userIDs).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	memberQuery := s.db.WithContext(ctx).
		Where("user_id IN ?", userIDs).
		Preload("Hospital")
	if filter.HospitalID != nil {
		memberQuery = memberQuery.Where("hospital_id = ?", *filter.HospitalID)
	}

	var allMembers []model.HospitalMember
	if err := memberQuery.Find(&allMembers).Error; err != nil {
		return nil, 0, err
	}

	membershipsByUser := make(map[uint][]model.HospitalMember, len(allMembers))
	for _, member := range allMembers {
		membershipsByUser[member.UserID] = append(membershipsByUser[member.UserID], member)
	}

	usersByID := make(map[uint]model.User, len(users))
	for _, u := range users {
		usersByID[u.ID] = u
	}

	results := make([]UserResult, 0, len(userIDs))
	for _, id := range userIDs {
		u, ok := usersByID[id]
		if !ok {
			continue
		}
		result := UserResult{
			ID:     u.ID,
			Email:  u.Email,
			Status: string(u.UserStatus),
		}
		if details := u.Details; details != nil {
			result.Name = details.Fullname
			result.Address = details.Address
			result.PhoneNumber = details.PhoneNumber
			if details.Department != nil {
				result.DepartmentName = details.Department.Name
			}
		}
		if members := membershipsByUser[u.ID]; len(members) > 0 {
			result.HospitalID = members[0].HospitalID
			result.Role = string(members[0].Role)
		}
		results = append(results, result)
	}

	return results, count, nil
}
