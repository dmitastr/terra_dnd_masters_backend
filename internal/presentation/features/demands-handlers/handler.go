package demands_handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"dnd_schedule/internal/common/constants"
	"dnd_schedule/internal/config"
	"dnd_schedule/internal/domain/models"
	"dnd_schedule/internal/domain/service/demands-service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type DemandsResponse struct {
	Demands []models.Demand `json:"demands"`
}

type DemandsRequest struct {
	Demands []models.Demand `json:"demands"`
}

type ErrorResponse struct {
	Error string `json:"error"`
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
	log     *logrus.Logger
}

func NewDemandsHandler(cfg config.ConfigProvider, service demands_service.IDemandsService) *DemandsHandler {
	return &DemandsHandler{service: service, cfg: cfg}
}

// GetDemands godoc
// @Summary Get all demands for current week
// @Description Returns a list of demands for session for current week
// @Tags demands
// @Produce json
// @Param week query string true "week" example(2020-01-01)
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

	dttm, err := time.Parse(constants.Layout, week)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
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
// @Param week query string true "week for fetching demands"  example(2020-01-01)
// @Param vk_id query int true "vk user id" example(123)
// @Success 200 {object} DemandsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /demands [get]
func (m DemandsHandler) GetDemandForUser(ctx *gin.Context) {
	week := ctx.Query("week")
	vkID := ctx.Query("vkID")
	if week == "" || vkID == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: errors.New("week  and vkID is required").Error()})
		return
	}

	dt, err := time.Parse(constants.Layout, week)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	dndDemands, err := m.service.GetDemands(ctx, dt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	resp := DemandsResponse{Demands: dndDemands}
	ctx.JSON(http.StatusOK, resp)
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

// DeleteDemands godoc
// @Summary Delete demands
// @Description Delete demands by list of id
// @Tags demands
// @Produce json
// @Param demand_id query []int true "List of demand IDs" collectionFormat(multi)
// @Success 204
// @Success 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /slots [delete]
func (m DemandsHandler) DeleteDemands(ctx *gin.Context) {
	m.log.Debug("DeleteDemands: start")
	ids := ctx.QueryArray("demand_id")

	demandIDs := make([]int, len(ids))
	for i, id := range ids {
		demandID, err := strconv.Atoi(id)
		if err != nil {
			m.log.WithError(err).Error("Error converting demand id")
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "error converting demand id"})
			return
		}
		demandIDs[i] = demandID
	}

	if err := m.service.DeleteDemands(ctx, demandIDs); err != nil {
		m.log.WithError(err).Error("Error deleting slots")
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	m.log.Debug("DeleteDemands: end")
	ctx.Status(http.StatusNoContent)
}
