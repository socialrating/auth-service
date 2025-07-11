package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/socialrating/auth-service/internal/models"
)

type TokenRepository interface {
	Store(ctx context.Context, token models.RefreshTokenRecord) error
	FindByJTI(ctx context.Context, jti string) (*models.RefreshTokenRecord, error)
	DeleteByJTI(ctx context.Context, jti string) error
}

const (
	// accessTokenTTL is the default time-to-live for access tokens.
	accessTokenTTL = 15 * time.Minute
	// refreshTokenTTL is the default time-to-live for refresh tokens.
	refreshTokenTTL = 7 * 24 * time.Hour
)

var signedMethod = jwt.SigningMethodHS512

type TokenService struct {
	secretKey string
	TokenRepo TokenRepository
}

func NewTokenService(JWTSecret string, TokenRepo TokenRepository) *TokenService {
	return &TokenService{secretKey: JWTSecret, TokenRepo: TokenRepo}
}

func (s *TokenService) GenerateTokenPair(ctx context.Context, userID string) (*models.TokenPair, error) {
	jti := uuid.New()

	tokenPair, err := s.generateTokens(ctx, userID, jti.String())
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	return tokenPair, nil
}

func (s *TokenService) RefreshTokens(
	ctx context.Context,
	refreshToken string,
) (*models.TokenPair, error) {
	rtParsed, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.secretKey), nil
	})
	if err != nil || !rtParsed.Valid {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	atClaims, ok := rtParsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type %w", err)
	}

	jti, ok := atClaims["jti"].(string)
	if !ok {
		return nil, fmt.Errorf("missing jti claim")
	}

	userID, ok := atClaims["sub"].(string)
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

	if err := s.TokenRepo.DeleteByJTI(ctx, jti); err != nil {
		return nil, fmt.Errorf("delete old token record: %w", err)
	}

	return s.GenerateTokenPair(ctx, userID)
}

func (s *TokenService) generateTokens(ctx context.Context, userID, jti string) (*models.TokenPair, error) {
	iat := time.Now()
	atExp := time.Now().Add(accessTokenTTL)
	rtExp := time.Now().Add(refreshTokenTTL)

	accessTokenClaims := jwt.MapClaims{
		"sub": userID,
		"jti": jti,
		"iat": iat.Unix(),
		"exp": atExp.Unix(),
	}
	jwtAccessToken := jwt.NewWithClaims(signedMethod, accessTokenClaims)

	token, err := jwtAccessToken.SignedString([]byte(s.secretKey))
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refreshTokenClaims := jwt.MapClaims{
		"sub": userID,
		"jti": jti,
		"iat": iat.Unix(),
		"exp": rtExp.Unix(),
	}
	jwtRefreshToken := jwt.NewWithClaims(signedMethod, refreshTokenClaims)

	tokenRefresh, err := jwtRefreshToken.SignedString([]byte(s.secretKey))
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	if err := s.TokenRepo.Store(ctx, models.RefreshTokenRecord{
		JTI:       jti,
		UserID:    userID,
		IssuedAt:  iat,
		ExpiresAt: rtExp,
	}); err != nil {
		return nil, fmt.Errorf("store token record: %w", err)
	}

	return &models.TokenPair{AccessToken: token, RefreshToken: tokenRefresh}, nil
}
