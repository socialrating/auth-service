package application

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/socialrating/auth-service/config"
	"github.com/socialrating/auth-service/internal/handler"
	"github.com/socialrating/auth-service/internal/repository"
	"github.com/socialrating/auth-service/internal/service"
)

const (
	ttl               = 5 * time.Minute
	readHeaderTimeout = 5 * time.Second
	maxAge            = 24 * time.Hour
)

func Run(ctx context.Context, cfg *config.Config) error {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatal(err)
	}

	slog.Debug("create connection to MongoDB", "uri", cfg.MongoURI)

	db := client.Database("auth_db")
	if err := db.Client().Ping(ctx, nil); err != nil {
		return fmt.Errorf("ping MongoDB: %w", err)
	}

	slog.Debug("connected to MongoDB", "uri", cfg.MongoURI)

	tokenRepo := repository.NewTokenRepository(db)
	tokenService := service.NewTokenService(cfg.JWTSecret, tokenRepo)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		MaxAge:           maxAge,
	}))

	api := handler.NewHandler(tokenService)

	api.RegisterRoutes(r)

	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:           r,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	for _, item := range r.Routes() {
		slog.Debug("registered route", "method", item.Method, "path", item.Path)
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}

	return nil
}
