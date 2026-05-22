package masters_handlers

import (
	"net/http"

	"dnd_schedule/internal/config"
	"dnd_schedule/internal/domain/models"
	"dnd_schedule/internal/domain/service/masters-service"
	"github.com/gin-gonic/gin"
)

type MastersResponse struct {
	DndMasters []models.Master `json:"dnd_masters"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type IMastersHandler interface {
	GetMasters(ctx *gin.Context)
	AddMasters(ctx *gin.Context)
	DeleteMasters(c *gin.Context)
}

type MastersHandler struct {
	service masters_service.IMastersService
	cfg     config.ConfigProvider
}

func NewMastersHandler(cfg config.ConfigProvider, service masters_service.IMastersService) *MastersHandler {
	return &MastersHandler{service: service, cfg: cfg}
}

// GetMasters godoc
// @Summary Get all masters
// @Description Returns a list of masters in JSON format
// @Tags masters
// @Produce json
// @Success 200 {object} MastersResponse
// @Failure 500 {object} ErrorResponse
// @Router /masters [get]
func (m MastersHandler) GetMasters(ctx *gin.Context) {
	dndMasters, err := m.service.GetMasters(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := MastersResponse{DndMasters: dndMasters}
	ctx.JSON(http.StatusOK, resp)
}

func (m MastersHandler) AddMasters(ctx *gin.Context) {
	// TODO implement me
	panic("implement me")
}

func (m MastersHandler) DeleteMasters(ctx *gin.Context) {
	// TODO implement me
	panic("implement me")
}
