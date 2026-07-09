package dto

type ChatRequest struct {
	SessionID *string `json:"session_id"`
	Message   string  `json:"message" binding:"required"`
}
