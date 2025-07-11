package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/socialrating/auth-service/internal/models"
)

type mongoTokenRepo struct {
	Collection *mongo.Collection
}

func NewTokenRepository(db *mongo.Database) *mongoTokenRepo {
	return &mongoTokenRepo{
		Collection: db.Collection("refresh_tokens"),
	}
}

func (r *mongoTokenRepo) Store(ctx context.Context, token models.RefreshTokenRecord) error {
	_, err := r.Collection.InsertOne(ctx, token)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("token with JTI %s already exists", token.JTI)
		}

		return fmt.Errorf("insert token: %w", err)
	}

	return nil
}

func (r *mongoTokenRepo) FindByJTI(ctx context.Context, jti string) (*models.RefreshTokenRecord, error) {
	var record models.RefreshTokenRecord

	err := r.Collection.FindOne(ctx, bson.M{"jti": jti}).Decode(&record)
	if err != nil {
		return nil, fmt.Errorf("find token by JTI: %w", err)
	}

	return &record, nil
}

func (r *mongoTokenRepo) DeleteByJTI(ctx context.Context, jti string) error {
	_, err := r.Collection.DeleteOne(ctx, bson.M{"jti": jti})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("token with JTI %s not found", jti)
		}

		return fmt.Errorf("delete token by JTI: %w", err)
	}

	return nil
}
