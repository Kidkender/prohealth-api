package response

import (
	"time"

	"github.com/kidkender/prohealth-api/internal/model"
)

type PendingInvitationResponse struct {
	ID        uint           `json:"id"`
	Email     string         `json:"email"`
	Role      model.UserRole `json:"role"`
	ExpiresAt time.Time      `json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
}

func ToPendingInvitationResponse(invitation model.HospitalInvitation) PendingInvitationResponse {
	return PendingInvitationResponse{
		ID:        invitation.ID,
		Email:     invitation.Email,
		Role:      invitation.Role,
		ExpiresAt: invitation.ExpiresAt,
		CreatedAt: invitation.CreatedAt,
	}
}

func ToPendingInvitationResponses(invitations []model.HospitalInvitation) []PendingInvitationResponse {
	responses := make([]PendingInvitationResponse, 0, len(invitations))
	for _, invitation := range invitations {
		responses = append(responses, ToPendingInvitationResponse(invitation))
	}
	return responses
}
