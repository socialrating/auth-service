package models

import "time"

type RefreshTokenRecord struct {
	JTI       string    `bson:"jti"`
	UserID    string    `bson:"user_id"`
	IssuedAt  time.Time `bson:"issued_at"`
	ExpiresAt time.Time `bson:"expires_at"`
	TokenHash string    `bson:"token_hash"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
