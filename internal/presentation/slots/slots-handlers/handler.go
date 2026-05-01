package slots_handlers

import (
	"net/http"

	"dnd_schedule/internal/config"
	"dnd_schedule/internal/domain/models"
	slotsservice "dnd_schedule/internal/domain/service/slots-service"
	"github.com/gin-gonic/gin"
)

type SlotsResponse struct {
	Slots []models.Slot `json:"slots"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ISlotsHandler interface {
	GetSlots(ctx *gin.Context)
	AddSlots(ctx *gin.Context)
	DeleteSlots(c *gin.Context)
}

type SlotsHandler struct {
	service slotsservice.ISlotsService
	cfg     config.ConfigProvider
}

func NewSlotsHandler(cfg config.ConfigProvider, service slotsservice.ISlotsService) *SlotsHandler {
	return &SlotsHandler{service: service, cfg: cfg}
}

// GetSlots godoc
// @Summary Get all slots-
// @Description Returns a list of slots for current week
// @Tags slots
// @Produce json
// @Success 200 {object} SlotsResponse
// @Failure 500 {object} ErrorResponse
// @Router /slots [get]
func (m SlotsHandler) GetSlots(ctx *gin.Context) {
	dndSlots, err := m.service.GetSlots(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := SlotsResponse{Slots: dndSlots}
	ctx.JSON(http.StatusOK, resp)
}

func (m SlotsHandler) AddSlots(ctx *gin.Context) {
	// TODO implement me
	panic("implement me")
}

func (m SlotsHandler) DeleteSlots(ctx *gin.Context) {
	// TODO implement me
	panic("implement me")
}
