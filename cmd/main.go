package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/socialrating/auth-service/config"
	"github.com/socialrating/auth-service/internal/handler"
	"github.com/socialrating/auth-service/internal/repository"
	"github.com/socialrating/auth-service/internal/service"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig("config.yaml")

	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatal(err)
	}
	db := client.Database("auth_db")
	tokenRepo := repository.NewTokenRepository(db)

	tokenService := &service.TokenService{
		SecretKey:       cfg.JWTSecret,
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		TokenRepo:       tokenRepo,
	}

	r := gin.Default()
	h := handler.NewHandler(tokenService)

	r.POST("/login", h.Login)
	r.POST("/refresh", h.Refresh)

	r.Run(":" + cfg.Port)
}
