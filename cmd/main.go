package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/socialrating/auth-service/internal/handler"
	"github.com/socialrating/auth-service/internal/repository"
	"github.com/socialrating/auth-service/internal/service"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	ctx := context.Background()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	db := client.Database("auth_db")
	tokenRepo := repository.NewTokenRepository(db)

	tokenService := &service.TokenService{
		SecretKey:       "supersecret",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		TokenRepo:       tokenRepo,
	}

	r := gin.Default()
	h := handler.NewHandler(tokenService)

	r.POST("/login", h.Login)
	r.POST("/refresh", h.Refresh)

	r.Run(":8080")
}
