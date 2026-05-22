package authenticate_handlers

import (
	"net/http"

	"dnd_schedule/internal/config"
	authenticate_service "dnd_schedule/internal/domain/service/authenticate-service"
	"dnd_schedule/internal/presentation/features/authenticate/authenticate_requests"
	"github.com/gin-gonic/gin"
)

type IAuthHandler interface {
	AddUser(ctx *gin.Context)
	GetToken(ctx *gin.Context)
}

type AuthHandler struct {
	service authenticate_service.AuthService
	cfg     config.ConfigProvider
}

func NewAuthHandler(cfg config.ConfigProvider, service authenticate_service.AuthService) *AuthHandler {
	return &AuthHandler{service: service, cfg: cfg}
}

// AddUser godoc
// @Summary Add new user
// @Description Add new user and returns token for basic auth
// @Tags auth
// @Produce json
// @Param	request	body	authenticate_requests.AuthRequest	true	"User payload"
// @Success 200 {object} authenticate_requests.AuthResponse
// @Failure 400 {object} authenticate_requests.ErrorResponse
// @Failure 500 {object} authenticate_requests.ErrorResponse
// @Router /register [post]
func (m AuthHandler) AddUser(c *gin.Context) {
	var request authenticate_requests.AuthRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, authenticate_requests.ErrorResponse{Error: err.Error()})
		return
	}
	token, err := m.service.RegisterUser(c, request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, authenticate_requests.ErrorResponse{Error: err.Error()})
		return
	}
	resp := authenticate_requests.AuthResponse{Token: token}
	c.JSON(http.StatusOK, resp)
}

// GetToken godoc
// @Summary get token
// @Description Get token for current user
// @Tags auth
// @Produce json
// @Param	request	body	authenticate_requests.AuthRequest	true	"User payload"
// @Success 200 {object} authenticate_requests.AuthResponse
// @Failure 400 {object} authenticate_requests.ErrorResponse
// @Failure 500 {object} authenticate_requests.ErrorResponse
// @Router /token [get]
func (m AuthHandler) GetToken(c *gin.Context) {
	var request authenticate_requests.AuthRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, authenticate_requests.ErrorResponse{Error: err.Error()})
		return
	}

	token, err := m.service.GetToken(c, request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, authenticate_requests.ErrorResponse{Error: err.Error()})
		return
	}
	resp := authenticate_requests.AuthResponse{Token: token}
	c.JSON(http.StatusOK, resp)
}
