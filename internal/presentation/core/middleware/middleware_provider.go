package middleware

import (
	"dnd_schedule/internal/config"
	authenticate_service "dnd_schedule/internal/domain/service/authenticate-service"
	"dnd_schedule/internal/presentation/core/middleware/auth"
)

type MiddlewareProvider struct {
	*auth.BearerValidator
}

func NewMiddlewareProvider(cfg config.ConfigProvider, authService authenticate_service.AuthService) *MiddlewareProvider {
	return &MiddlewareProvider{auth.NewBearerValidator(cfg, authService)}
}
