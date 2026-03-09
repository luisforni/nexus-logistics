package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luisforni/nexus-logistics/backend/internal/api/middleware"
	"github.com/luisforni/nexus-logistics/backend/internal/domain"
)

type authService interface {
	Login(ctx context.Context, req domain.LoginRequest) (*domain.TokenPair, error)
	ValidateToken(token string) (*domain.Claims, error)
	RefreshToken(ctx context.Context, token string) (*domain.TokenPair, error)
}

type AuthHandler struct {
	svc authService
}

func NewAuthHandler(svc authService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	pair, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, pair)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	pair, err := h.svc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, pair)
}

type errorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func currentUserID(c *gin.Context) string {
	id, _ := c.Get(middleware.ContextUserID)
	return id.(string)
}

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func Ready(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

type AnalyticsHandler struct{}

func NewAnalyticsHandler() *AnalyticsHandler { return &AnalyticsHandler{} }

func (h *AnalyticsHandler) DemandForecast(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "proxied to optimizer", "status": "ok"})
}

func (h *AnalyticsHandler) KPIs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"total_shipments":    4821,
		"on_time_rate":       0.947,
		"avg_delivery_hours": 38.2,
	})
}
