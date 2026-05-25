package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

	mt.Run("save inserts photo fields", func(mt *mtest.T) {
		repo := NewPhotoRepository(mt.DB)
		ts := time.Date(2026, 1, 20, 12, 30, 0, 0, time.UTC)
		photo := makeTestPhoto(ts, "dev1", "desc1")

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := repo.Save(context.TODO(), photo)
		assert.NoError(t, err)

		started := mt.GetStartedEvent()
		if started == nil {
			mt.Fatalf("expected started event")
		}
		docs, err := started.Command.Lookup("documents").Array().Values()
		if err != nil {
			mt.Fatalf("expected documents array values, got error %v", err)
		}
		if len(docs) != 1 {
			mt.Fatalf("expected one inserted document, got %d", len(docs))
		}
		doc := docs[0].Document()
		if got := doc.Lookup("device_id").StringValue(); got != "dev1" {
			mt.Fatalf("expected device_id dev1, got %q", got)
		}
		if got := doc.Lookup("image_type").StringValue(); got != "jpg" {
			mt.Fatalf("expected image_type jpg, got %q", got)
		}
		if got := doc.Lookup("text").StringValue(); got != "desc1" {
			mt.Fatalf("expected text desc1, got %q", got)
		}
	})

	mt.Run("get photos uses filter and timestamp desc sort", func(mt *mtest.T) {
		repo := NewPhotoRepository(mt.DB)
		photo := makeTestPhoto(time.Now(), "dev2", "desc2")
		filters := map[string]any{"device_id": "dev2"}

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "test.photos", mtest.FirstBatch, photoDoc(photo)),
			mtest.CreateCursorResponse(0, "test.photos", mtest.NextBatch),
		)

		photos, err := repo.GetPhotos(context.TODO(), filters)
		assert.NoError(t, err)
		assert.Len(t, photos, 1)
		assert.Equal(t, "dev2", photos[0].DeviceID)

		started := mt.GetStartedEvent()
		if started == nil {
			mt.Fatalf("expected started event")
		}
		filter := started.Command.Lookup("filter").Document()
		if got := filter.Lookup("device_id").StringValue(); got != "dev2" {
			mt.Fatalf("expected device_id filter dev2, got %q", got)
		}
		sort := started.Command.Lookup("sort").Document()
		if got := sort.Lookup("timestamp").Int32(); got != -1 {
			mt.Fatalf("expected timestamp sort -1, got %d", got)
		}
	})

	mt.Run("get by id maps document and uses object id filter", func(mt *mtest.T) {
		repo := NewPhotoRepository(mt.DB)
		photo := makeTestPhoto(time.Now(), "dev3", "desc3")

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "test.photos", mtest.FirstBatch, photoDoc(photo)),
		)

		found, err := repo.GetByID(context.TODO(), photo.ID.Hex())
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, photo.ID, found.ID)

		started := mt.GetStartedEvent()
		if started == nil {
			mt.Fatalf("expected started event")
		}
		filter := started.Command.Lookup("filter").Document()
		if got := filter.Lookup("_id").ObjectID(); got != photo.ID {
			mt.Fatalf("expected _id filter %s, got %s", photo.ID.Hex(), got.Hex())
		}
	})

	mt.Run("get by id returns no documents", func(mt *mtest.T) {
		repo := NewPhotoRepository(mt.DB)
		photoID := primitive.NewObjectID()

		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.photos", mtest.FirstBatch))

		_, err := repo.GetByID(context.TODO(), photoID.Hex())
		if !errors.Is(err, mongo.ErrNoDocuments) {
			mt.Fatalf("expected ErrNoDocuments, got %v", err)
		}
	})

	mt.Run("delete uses object id filter", func(mt *mtest.T) {
		repo := NewPhotoRepository(mt.DB)
		photoID := primitive.NewObjectID()

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := repo.Delete(context.TODO(), photoID.Hex())
		assert.NoError(t, err)

		started := mt.GetStartedEvent()
		if started == nil {
			mt.Fatalf("expected started event")
		}
		deletes, err := started.Command.Lookup("deletes").Array().Values()
		if err != nil {
			mt.Fatalf("expected deletes array values, got error %v", err)
		}
		if len(deletes) != 1 {
			mt.Fatalf("expected one delete entry, got %d", len(deletes))
		}
		filter := deletes[0].Document().Lookup("q").Document()
		if got := filter.Lookup("_id").ObjectID(); got != photoID {
			mt.Fatalf("expected _id filter %s, got %s", photoID.Hex(), got.Hex())
		}
	})

	mt.Run("delete all uses empty filter and returns deleted count", func(mt *mtest.T) {
		repo := NewPhotoRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(3)}))

		count, err := repo.DeleteAll(context.TODO())
		assert.NoError(t, err)
		assert.Equal(t, int64(3), count)

		started := mt.GetStartedEvent()
		if started == nil {
			mt.Fatalf("expected started event")
		}
		deletes, err := started.Command.Lookup("deletes").Array().Values()
		if err != nil {
			mt.Fatalf("expected deletes array values, got error %v", err)
		}
		if len(deletes) != 1 {
			mt.Fatalf("expected one delete entry, got %d", len(deletes))
		}
		filterValue := deletes[0].Document().Lookup("q")
		if filterValue.Type != bsontype.EmbeddedDocument {
			mt.Fatalf("expected embedded delete filter, got %s", filterValue.Type)
		}
		elements, err := filterValue.Document().Elements()
		if err != nil {
			mt.Fatalf("expected valid delete filter, got error %v", err)
		}
		if len(elements) != 0 {
			mt.Fatalf("expected empty delete filter, got %s", filterValue.Document().String())
		}
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
