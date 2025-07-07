package repository

import (
	"context"

	"github.com/socialrating/auth-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type TokenRepository interface {
	Store(ctx context.Context, token models.TokenRecord) error
	FindByJTI(ctx context.Context, jti string) (*models.TokenRecord, error)
	DeleteByJTI(ctx context.Context, jti string) error
}

type mongoTokenRepo struct {
	Collection *mongo.Collection
}

func NewTokenRepository(db *mongo.Database) *mongoTokenRepo {
	return &mongoTokenRepo{
		Collection: db.Collection("refresh_tokens"),
	}
}

func (r *mongoTokenRepo) Store(ctx context.Context, token models.TokenRecord) error {
	_, err := r.Collection.InsertOne(ctx, token)
	return err
}

func (r *mongoTokenRepo) FindByJTI(ctx context.Context, jti string) (*models.TokenRecord, error) {
	var record models.TokenRecord
	err := r.Collection.FindOne(ctx, bson.M{"jti": jti}).Decode(&record)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *mongoTokenRepo) DeleteByJTI(ctx context.Context, jti string) error {
	_, err := r.Collection.DeleteOne(ctx, bson.M{"jti": jti})
	return err
}
