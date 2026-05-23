package repository

import (
	"context"
	"fmt"
	"log"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mqtt-streaming-server/domain"
	"go.mongodb.org/mongo-driver/bson"
)

type photoRepository struct {
	db *mongo.Database
}

func NewPhotoRepository(db *mongo.Database) *photoRepository {
	return &photoRepository{db: db}
}

func (repo *photoRepository) GetPhotos(ctx context.Context, filters map[string]any) ([]*domain.Photo, error) {
	collection := repo.db.Collection("photos")
	photos := make([]*domain.Photo, 0)
	cursor, err := collection.Find(ctx, filters, &options.FindOptions{
		Sort: map[string]int{"timestamp": -1}, // Sort by timestamp in descending order
	})
	if err != nil {
		log.Printf("GetPhotos: Find error: %v", err)
		return nil, fmt.Errorf("GetPhotos: Find error: %w", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var photo domain.Photo
		if err := cursor.Decode(&photo); err != nil {
			log.Printf("GetPhotos: Decode error: %v", err)
			return nil, fmt.Errorf("GetPhotos: Decode error: %w", err)
		}
		photos = append(photos, &photo)
	}

	if err := cursor.Err(); err != nil {
		log.Printf("GetPhotos: Cursor error: %v", err)
		return nil, fmt.Errorf("GetPhotos: Cursor error: %w", err)
	}

	return photos, nil
}

func (repo *photoRepository) Save(ctx context.Context, photo *domain.Photo) error {
	collection := repo.db.Collection("photos")
<<<<<<< HEAD
	_, err := collection.InsertOne(ctx, photo)
	if err != nil {
		log.Printf("Save: InsertOne error: %v", err)
		return fmt.Errorf("Save: InsertOne error: %w", err)
	}
	return nil
=======
	res, err := collection.InsertOne(ctx, photo)
	if err == nil {
		if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
			photo.ID = oid
		}
	}
	return err
}

func (repo *photoRepository) UpdatePhotoTextAndMedicalData(ctx context.Context, id primitive.ObjectID, updates map[string]any) error {
	collection := repo.db.Collection("photos")
	_, err := collection.UpdateOne(ctx, map[string]any{"_id": id}, map[string]any{"$set": updates})
	return err
>>>>>>> ce1de12 (Created a sepparate service for ocr)
}

func (r *photoRepository) GetByID(ctx context.Context, id string) (*domain.Photo, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	collection := r.db.Collection("photos")

	var photo domain.Photo
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&photo)
	if err != nil {
		return nil, err
	}

	return &photo, nil
}

func (r *photoRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	collection := r.db.Collection("photos")
	_, err = collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (repo *photoRepository) DeleteAll(ctx context.Context) (int64, error) {
	collection := repo.db.Collection("photos")
	result, err := collection.DeleteMany(ctx, map[string]any{})
	if err != nil {
		log.Printf("DeleteAll: DeleteMany error: %v", err)
		return 0, fmt.Errorf("DeleteAll: DeleteMany error: %w", err)
	}
	return result.DeletedCount, nil
}
