package demands_handlers

import (
	"errors"
	"net/http"
	"time"

	"dnd_schedule/internal/config"
	"dnd_schedule/internal/domain/models"
	"dnd_schedule/internal/domain/service/demands-service"
	"github.com/gin-gonic/gin"
)

type DemandsResponse struct {
	Demands []models.Demand `json:"demands"`
}

type DemandsRequest struct {
	Demands []models.Demand `json:"demands"`
}

type ErrorResponse struct {
	Error      string          `json:"error"`
	BadDemands []models.Demand `json:"bad_demands"`
}

type IDemandsHandler interface {
	GetDemands(ctx *gin.Context)
	GetDemandForUser(ctx *gin.Context)
	AddDemands(ctx *gin.Context)
	DeleteDemands(c *gin.Context)
}

type DemandsHandler struct {
	service demands_service.IDemandsService
	cfg     config.ConfigProvider
}

func NewDemandsHandler(cfg config.ConfigProvider, service demands_service.IDemandsService) *DemandsHandler {
	return &DemandsHandler{service: service, cfg: cfg}
}

// GetDemands godoc
// @Summary Get all demands for current week
// @Description Returns a list of demands for session for current week
// @Tags demands
// @Produce json
// @Param id query string true "week"
// @Success 200 {object} DemandsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /demands [get]
func (m DemandsHandler) GetDemands(ctx *gin.Context) {
	week := ctx.Query("week")
	if week == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: errors.New("week is required").Error()})
		return
	}

	dttm, err := time.Parse("2006-01-02", week)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	dndDemands, err := m.service.GetDemands(ctx, dttm)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	resp := DemandsResponse{Demands: dndDemands}
	ctx.JSON(http.StatusOK, resp)
}

// GetDemandForUser godoc
// @Summary Get all demands for specific user for current week
// @Description Returns a list of demands for session for current week
// @Tags demands
// @Produce json
// @Param id query string true "week"
// @Param id query int true "user_id"
// @Success 200 {object} DemandsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /demands [get]
func (m DemandsHandler) GetDemandForUser(ctx *gin.Context) {
	// TODO implement me
	panic("implement me")
}

// AddDemands godoc
// @Summary Add demands
// @Description Add a list of user demands for session
// @Tags demands
// @Produce json
// @Param demands body DemandsRequest true "A list of demands"
// @Success 200 {object} DemandsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /demands [post]
func (m DemandsHandler) AddDemands(ctx *gin.Context) {
	var request DemandsRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	addedDemands, err := m.service.AddDemands(ctx, request.Demands)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	resp := DemandsResponse{Demands: addedDemands}
	ctx.JSON(http.StatusOK, resp)
}

func (m DemandsHandler) DeleteDemands(ctx *gin.Context) {
	// TODO implement me
	panic("implement me")
}
