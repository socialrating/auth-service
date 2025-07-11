package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/socialrating/auth-service/internal/models"
)

type TokenService interface {
	GenerateTokenPair(ctx context.Context, userID string) (*models.TokenPair, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*models.TokenPair, error)
}

type Handler struct {
	service TokenService
}

func NewHandler(s TokenService) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/auth")
	api.POST("/login", h.Login)
	api.POST("/refresh", h.Refresh)
}

func (h *Handler) Login(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user_id"})
		return
	}

	pair, err := h.service.GenerateTokenPair(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.JSON(http.StatusOK, pair)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON req"})
		return
	}

	pair, err := h.service.RefreshTokens(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})

		return
	}

	c.JSON(http.StatusOK, pair)
}
