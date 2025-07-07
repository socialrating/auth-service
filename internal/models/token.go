package models

import "time"

type TokenRecord struct {
	JTI       string    `bson:"jti"`
	UserID    string    `bson:"user_id"`
	IssuedAt  time.Time `bson:"issued_at"`
	ExpiresAt time.Time `bson:"expires_at"`
	TokenHash string    `bson:"token_hash"`
}

type TokenPair struct {
	AccessToken  string `bson:"access_token"`
	RefreshToken string `bson:"refresh_token"`
}
