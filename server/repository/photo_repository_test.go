package repository

import (
	"context"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"mqtt-streaming-server/domain"
)

func makeTestPhoto(ts time.Time, deviceID, text string) *domain.Photo {
	return &domain.Photo{
		ID:        primitive.NewObjectID(),
		Timestamp: ts,
		ImageType: "jpg",
		DeviceID:  deviceID,
		Text:      text,
	}
}

func photoDoc(photo *domain.Photo) bson.D {
	return bson.D{
		{Key: "_id", Value: photo.ID},
		{Key: "timestamp", Value: photo.Timestamp},
		{Key: "image_type", Value: photo.ImageType},
		{Key: "device_id", Value: photo.DeviceID},
		{Key: "text", Value: photo.Text},
	}
}

func TestPhotoRepository_BasicOperations(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("save", func(mt *mtest.T) {
		repo := NewPhotoRepository(mt.DB)
		photo := makeTestPhoto(time.Now(), "dev1", "desc1")

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := repo.Save(context.TODO(), photo)
		assert.NoError(t, err)
	})

	mt.Run("get photos", func(mt *mtest.T) {
		repo := NewPhotoRepository(mt.DB)
		photo := makeTestPhoto(time.Now(), "dev2", "desc2")

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "test.photos", mtest.FirstBatch, photoDoc(photo)),
			mtest.CreateCursorResponse(0, "test.photos", mtest.NextBatch),
		)

		photos, err := repo.GetPhotos(context.TODO(), map[string]any{"device_id": "dev2"})
		assert.NoError(t, err)
		assert.Len(t, photos, 1)
		assert.Equal(t, "dev2", photos[0].DeviceID)
	})

	mt.Run("get by id", func(mt *mtest.T) {
		repo := NewPhotoRepository(mt.DB)
		photo := makeTestPhoto(time.Now(), "dev3", "desc3")

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "test.photos", mtest.FirstBatch, photoDoc(photo)),
		)

		found, err := repo.GetByID(context.TODO(), photo.ID.Hex())
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, photo.ID, found.ID)
	})

	mt.Run("delete", func(mt *mtest.T) {
		repo := NewPhotoRepository(mt.DB)
		photoID := primitive.NewObjectID()

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := repo.Delete(context.TODO(), photoID.Hex())
		assert.NoError(t, err)
	})

	mt.Run("delete all", func(mt *mtest.T) {
		repo := NewPhotoRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		count, err := repo.DeleteAll(context.TODO())
		assert.NoError(t, err)
		assert.True(t, count >= 0)
	})
}

func TestPhotoRepository_GetPhotos_Error(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("find error", func(mt *mtest.T) {
		repo := NewPhotoRepository(mt.DB)

		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{Message: "find error"}),
		)

		_, err := repo.GetPhotos(context.TODO(), map[string]any{})
		assert.Error(t, err)
	})
}

func TestPhotoRepository_Save_Error(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("insert error", func(mt *mtest.T) {
		repo := NewPhotoRepository(mt.DB)

		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{Message: "insert error"}),
		)

		err := repo.Save(context.TODO(), &domain.Photo{})
		assert.Error(t, err)
	})
}

func TestPhotoRepository_GetByID_InvalidID(t *testing.T) {
	repo := NewPhotoRepository(nil)

	_, err := repo.GetByID(context.TODO(), "nothex")
	assert.Error(t, err)
}

func TestPhotoRepository_Delete_InvalidID(t *testing.T) {
	repo := NewPhotoRepository(nil)

	err := repo.Delete(context.TODO(), "nothex")
	assert.Error(t, err)
}
