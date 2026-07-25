package slots_handlers

import (
	"net/http"
	"strconv"

	"dnd_schedule/internal/config"
	"dnd_schedule/internal/domain/models"
	slotsservice "dnd_schedule/internal/domain/service/slots-service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type SlotsResponse struct {
	Slots []models.Slot `json:"slots"`
}

type SlotsRequest struct {
	Slots []models.Slot `json:"slots"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ISlotsHandler interface {
	GetSlots(c *gin.Context)
	AddSlots(c *gin.Context)
	DeleteSlots(c *gin.Context)
	UpdateSlots(c *gin.Context)
}

type SlotsHandler struct {
	service slotsservice.ISlotsService
	cfg     config.ConfigProvider
	log     *logrus.Logger
}

func NewSlotsHandler(cfg config.ConfigProvider, service slotsservice.ISlotsService, log *logrus.Logger) *SlotsHandler {
	return &SlotsHandler{service: service, cfg: cfg, log: log}
}

// AddSlots godoc
// @Summary Add slots
// @Description Add list of new slots
// @Tags slots
// @Accept json
// @Produce json
// @Param slots body  SlotsRequest true "list of slots"
// @Success 204
// @Success 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /slots [post]
func (m SlotsHandler) AddSlots(c *gin.Context) {
	m.log.Debug("AddSlots: start")
	var request SlotsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		m.log.WithError(err).Error("Error parsing request body")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if err := m.service.AddSlots(c, request.Slots); err != nil {
		m.log.WithError(err).Error("Error adding slots")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	m.log.Debug("AddSlots: end")
	c.Status(http.StatusNoContent)
}

// DeleteSlots godoc
// @Summary Delete slots
// @Description Delete slots by list of id
// @Tags slots
// @Produce json
// @Param slot_id query []int true "List of slot IDs" collectionFormat(multi)
// @Success 204
// @Success 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /slots [delete]
func (m SlotsHandler) DeleteSlots(c *gin.Context) {
	m.log.Debug("DeleteSlots: start")
	ids := c.QueryArray("slot_id")

	slotIDs := make([]int, len(ids))
	for i, id := range ids {
		slotID, err := strconv.Atoi(id)
		if err != nil {
			m.log.WithError(err).Error("Error converting slot id")
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "error converting slot id"})
			return
		}
		slotIDs[i] = slotID
	}

	if err := m.service.DeleteSlots(c, slotIDs); err != nil {
		m.log.WithError(err).Error("Error deleting slots")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	m.log.Debug("DeleteSlots: end")
	c.Status(http.StatusNoContent)
}

// UpdateSlots godoc
// @Summary Update slots
// @Description Update slots
// @Tags slots
// @Accept json
// @Produce json
// @Param slots body  SlotsRequest true "list of slots"
// @Success 204
// @Success 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /slots [put]
func (m SlotsHandler) UpdateSlots(c *gin.Context) {
	m.log.Debug("UpdateSlots: start")
	var request SlotsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		m.log.WithError(err).Error("Error parsing request body")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if err := m.service.UpdateSlots(c, request.Slots); err != nil {
		m.log.WithError(err).Error("Error updating slots")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	m.log.Debug("UpdateSlots: end")
	c.Status(http.StatusNoContent)
}

// GetSlots godoc
// @Summary Get all slots
// @Description Returns a list of slots for current week
// @Tags slots
// @Produce json
// @Param from query string true "min date for getting slots" example(2020-01-01)
// @Param to query string true "max date for getting slots" example(2020-01-01)
// @Success 200 {object} SlotsResponse
// @Failure 500 {object} ErrorResponse
// @Router /slots [get]
func (m SlotsHandler) GetSlots(c *gin.Context) {
	m.log.Debug("GetSlots: start")
	from := c.Query("from")
	to := c.Query("to")
	if from == "" || to == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from and to must be provided"})
		return
	}

	dndSlots, err := m.service.GetSlots(c, from, to)
	if err != nil {
		m.log.WithError(err).Error("GetSlots")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	resp := SlotsResponse{Slots: dndSlots}
	c.JSON(http.StatusOK, resp)
}
