package authenticate_requests

import "dnd_schedule/internal/domain/models"

type AuthResponse struct {
	Token string `json:"token"`
}

type AuthRequest struct {
	User models.User `json:"user"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
