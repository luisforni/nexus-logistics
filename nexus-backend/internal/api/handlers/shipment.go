package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/luisforni/nexus-logistics/backend/internal/domain"
)

type shipmentService interface {
	Create(ctx context.Context, req domain.CreateShipmentRequest, senderID uuid.UUID) (*domain.Shipment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Shipment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, req domain.UpdateStatusRequest, actorID uuid.UUID) (*domain.Shipment, error)
	List(ctx context.Context, senderID uuid.UUID, limit, offset int) ([]domain.Shipment, int64, error)
}

type ShipmentHandler struct {
	svc shipmentService
}

func NewShipmentHandler(svc shipmentService) *ShipmentHandler {
	return &ShipmentHandler{svc: svc}
}

func (h *ShipmentHandler) Create(c *gin.Context) {
	var req domain.CreateShipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	senderID, err := uuid.Parse(currentUserID(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid user id"})
		return
	}

	shipment, err := h.svc.Create(c.Request.Context(), req, senderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, shipment)
}

func (h *ShipmentHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	senderID, _ := uuid.Parse(currentUserID(c))

	shipments, total, err := h.svc.List(c.Request.Context(), senderID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, listResponse{
		Data:   shipments,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

func (h *ShipmentHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid shipment id"})
		return
	}

	shipment, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, shipment)
}

func (h *ShipmentHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid shipment id"})
		return
	}

	var req domain.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	actorID, _ := uuid.Parse(currentUserID(c))
	shipment, err := h.svc.UpdateStatus(c.Request.Context(), id, req, actorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, shipment)
}

func (h *ShipmentHandler) GetBlockchainTrace(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid shipment id"})
		return
	}

	shipment, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"shipment_id":       shipment.ID,
		"tracking_number":   shipment.TrackingNumber,
		"blockchain_anchor": shipment.BlockchainTxHash,
		"events":            shipment.Events,
	})
}

type listResponse struct {
	Data   interface{} `json:"data"`
	Total  int64       `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}
