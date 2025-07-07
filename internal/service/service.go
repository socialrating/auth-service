package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/socialrating/auth-service/internal/models"
	"github.com/socialrating/auth-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type TokenService struct {
	SecretKey       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	TokenRepo       repository.TokenRepository
}

func (s *TokenService) GenerateTokenPair(ctx context.Context, userID string) (*models.TokenPair, error) {
	jti, err := generateRandomString(32)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.generateAccessToken(userID, jti)
	if err != nil {
		return nil, err
	}

	refreshToken, hash, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	record := models.TokenRecord{
		JTI:       jti,
		UserID:    userID,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(s.RefreshTokenTTL),
		TokenHash: hash,
	}

	if err := s.TokenRepo.Store(ctx, record); err != nil {
		return nil, err
	}

	return &models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *TokenService) RefreshTokens(ctx context.Context, accessTokenRaw, refreshToken string) (*models.TokenPair, error) {
	parsed, err := jwt.Parse(accessTokenRaw, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.SecretKey), nil
	})
	if err != nil || !parsed.Valid {
		return nil, errors.New("invalid access token")
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	jti, ok := claims["jti"].(string)
	if !ok {
		return nil, errors.New("missing jti")
	}
	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New("missing sub")
	}

	record, err := s.TokenRepo.FindByJTI(ctx, jti)
	if err != nil {
		return nil, errors.New("refresh token not found")
	}

	if time.Now().After(record.ExpiresAt) {
		return nil, errors.New("refresh token expired")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(record.TokenHash), []byte(refreshToken)); err != nil {
		return nil, errors.New("refresh token mismatch")
	}

	_ = s.TokenRepo.DeleteByJTI(ctx, jti)

	return s.GenerateTokenPair(ctx, userID)
}

func (s *TokenService) generateAccessToken(userID, jti string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"jti": jti,
		"exp": time.Now().Add(s.AccessTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString([]byte(s.SecretKey))
}

func generateRefreshToken() (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	token := base64.URLEncoding.EncodeToString(raw)
	hash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	return token, string(hash), nil
}

func generateRandomString(length int) (string, error) {
	raw := make([]byte, length)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(raw), nil
}
