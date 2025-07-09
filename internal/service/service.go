package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
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
type Config struct {
	JWTSecret string `yaml:"jwt_secret"`
}

func NewTokenService(JWTSecret string, accessTTL, refreshTTL time.Duration, TokenRepo repository.TokenRepository) *TokenService {
	return &TokenService{
		SecretKey:       JWTSecret,
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		TokenRepo:       TokenRepo,
	}
}

func (s *TokenService) GenerateTokenPair(ctx context.Context, userID string) (*models.TokenPair, error) {
	jti, err := generateRandomString(32)
	if err != nil {
		return nil, fmt.Errorf("generate random string: %w", err)
	}

	accessToken, err := s.generateAccessToken(userID, jti)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, hash, err := generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	record := models.TokenRecord{
		JTI:       jti,
		UserID:    userID,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(s.RefreshTokenTTL),
		TokenHash: hash,
	}

	if err := s.TokenRepo.Store(ctx, record); err != nil {
		return nil, fmt.Errorf("store token record: %w", err)
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
		return nil, fmt.Errorf("invalid access token: %w", err)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type %w", err)
	}

	jti, ok := claims["jti"].(string)
	if !ok {
		return nil, fmt.Errorf("missing jti claim")
	}
	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("missing sub claim")
	}

	record, err := s.TokenRepo.FindByJTI(ctx, jti)
	if err != nil {
		return nil, fmt.Errorf("refresh token not found: %w", err)
	}

	if time.Now().After(record.ExpiresAt) {
		return nil, fmt.Errorf("refresh token expired")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(record.TokenHash), []byte(refreshToken)); err != nil {
		return nil, fmt.Errorf("refresh token mismatch: %w", err)
	}

	if err := s.TokenRepo.DeleteByJTI(ctx, jti); err != nil {
		return nil, fmt.Errorf("delete old token record: %w", err)
	}
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
		return "", "", fmt.Errorf("read random bytes: %w", err)
	}
	token := base64.URLEncoding.EncodeToString(raw)
	hash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("hash refresh token: %w", err)
	}
	return token, string(hash), nil
}

func generateRandomString(length int) (string, error) {
	raw := make([]byte, length)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(raw), nil
}
