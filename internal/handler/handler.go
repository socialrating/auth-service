package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/socialrating/auth-service/internal/service"
)

type Handler struct {
	service *service.TokenService
}

func NewHandler(s *service.TokenService) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Login(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {

		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user_id"})
		return
	}

	pair, err := h.service.GenerateTokenPair(context.Background(), "user-1")
	if err != nil {
		wrappedErr := fmt.Errorf("failed to generate token pair: %w", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": wrappedErr.Error()})
		return
	}
	c.JSON(http.StatusOK, pair)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON req"})
		return
	}

	pair, err := h.service.RefreshTokens(context.Background(), req.AccessToken, req.RefreshToken)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to refresh token pair: %w", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": wrappedErr.Error()})
		return
	}
	c.JSON(http.StatusOK, pair)
}
