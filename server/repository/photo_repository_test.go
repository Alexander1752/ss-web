package repository

import (
	"context"
	"testing"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"mqtt-streaming-server/domain"
)

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