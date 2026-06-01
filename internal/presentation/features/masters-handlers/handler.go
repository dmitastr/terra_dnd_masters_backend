package masters_handlers

import (
	"errors"
	"net/http"
	"strconv"

	"dnd_schedule/internal/config"
	"dnd_schedule/internal/domain/models"
	masters_service "dnd_schedule/internal/domain/service/masters-service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type MastersRequestWithPayload struct {
	DndMasters []models.Master `json:"dnd_masters"`
}
type MastersResponseNoContent struct {
	Status string `json:"status"`
}

type MastersResponseWithPayload struct {
	DndMasters []models.Master `json:"dnd_masters"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type IMastersHandler interface {
	GetMasters(ctx *gin.Context)
	AddMasters(ctx *gin.Context)
	DeleteMasters(c *gin.Context)
	UpdateMasters(c *gin.Context)
}

type MastersHandler struct {
	service masters_service.IMastersService
	cfg     config.ConfigProvider
	log     *logrus.Logger
}

func NewMastersHandler(cfg config.ConfigProvider, service masters_service.IMastersService, log *logrus.Logger) IMastersHandler {
	return &MastersHandler{service: service, cfg: cfg, log: log}
}

// GetMasters godoc
// @Summary Get all masters
// @Description Returns a list of masters in JSON format
// @Tags masters
// @Accept json
// @Produce json
// @Success 200 {object} MastersResponseWithPayload
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /masters [get]
func (m MastersHandler) GetMasters(ctx *gin.Context) {
	m.log.Debug("GetMasters: start")
	dndMasters, err := m.service.GetMasters(ctx)
	if err != nil {
		m.log.Errorf("GetMasters: failed to get masters: %s", err.Error())
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	resp := MastersResponseWithPayload{DndMasters: dndMasters}
	ctx.JSON(http.StatusOK, resp)
}

// AddMasters godoc
// @Summary Add list of masters
// @Description Add list of masters
// @Tags masters
// @Accept json
// @Produce json
// @Param demands body MastersRequestWithPayload true "A list of masters"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /masters [post]
func (m MastersHandler) AddMasters(ctx *gin.Context) {
	m.log.Debug("AddMasters: start")
	var request MastersRequestWithPayload
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	err := m.service.AddMasters(ctx, request.DndMasters)
	if err != nil {
		m.log.WithError(err).Error("AddMasters: fail")
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	ctx.Status(http.StatusNoContent)
}

// DeleteMasters godoc
// @Summary Delete masters
// @Description DeleteMasters
// @Tags masters
// @Produce json
// @Param masterIDs query []int true "List of master IDs" collectionFormat(multi)
// @Success     204
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /masters [delete]
func (m MastersHandler) DeleteMasters(ctx *gin.Context) {
	m.log.Debug("DeleteMasters: start")
	masterIDsQuery := ctx.QueryArray("masterIDs")
	if len(masterIDsQuery) == 0 {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "no masterIDs provided"})
		return
	}

	masterIDs := make([]int, len(masterIDsQuery))
	for i, masterID := range masterIDsQuery {
		id, err := strconv.Atoi(masterID)
		if err != nil {
			m.log.WithError(err).Errorf("DeleteMasters: failed to convert string to int")
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		masterIDs[i] = id
	}

	if err := m.service.DeleteMasters(ctx, masterIDs); err != nil {
		m.log.WithError(err).Errorf("DeleteMasters: failed to delete masters")
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	ctx.Status(http.StatusNoContent)
}

// UpdateMasters godoc
// @Summary Update masters
// @Description UpdateMasters
// @Tags masters
// @Accept  json
// @Produce json
// @Param demands body MastersRequestWithPayload true "A list of masters"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /masters [put]
func (m MastersHandler) UpdateMasters(ctx *gin.Context) {
	m.log.Debug("UpdateMasters: start")
	var request MastersRequestWithPayload
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := m.service.UpdateMasters(ctx, request.DndMasters); err != nil {
		switch {
		case errors.Is(err, masters_service.ErrEmptyInput),
			errors.Is(err, masters_service.ErrInvalidMasterID),
			errors.Is(err, masters_service.ErrEmptyMasterName):
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			m.log.WithError(err).Error("UpdateMasters: service error")
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		}
		return
	}
	ctx.Status(http.StatusNoContent)
}
