package repository

import (
	"context"
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"

	"mqtt-streaming-server/domain"
)

func TestDeviceRepository_GetAllDevices(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("maps documents and uses empty filter", func(mt *mtest.T) {
		repo := NewDeviceRepository(mt.DB)

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "test.devices", mtest.FirstBatch,
				bson.D{{Key: "device_id", Value: "dev-1"}, {Key: "device_name", Value: "Camera 1"}, {Key: "device_status", Value: "active"}},
				bson.D{{Key: "device_id", Value: "dev-2"}, {Key: "device_name", Value: "Camera 2"}, {Key: "device_status", Value: "inactive"}},
			),
			mtest.CreateCursorResponse(0, "test.devices", mtest.NextBatch),
		)

		devices, err := repo.GetAllDevices(context.Background())
		if err != nil {
			mt.Fatalf("expected nil error, got %v", err)
		}
		if len(devices) != 2 {
			mt.Fatalf("expected 2 devices, got %d", len(devices))
		}
		if devices[0].DeviceID != "dev-1" || devices[0].DeviceName != "Camera 1" {
			mt.Fatalf("unexpected first device mapping: %+v", devices[0])
		}

		started := mt.GetStartedEvent()
		if started == nil {
			mt.Fatalf("expected started event")
		}
		filterValue := started.Command.Lookup("filter")
		if filterValue.Type != bsontype.EmbeddedDocument {
			mt.Fatalf("expected embedded filter document, got %s", filterValue.Type)
		}
		elements, err := filterValue.Document().Elements()
		if err != nil {
			mt.Fatalf("expected valid filter document, got error %v", err)
		}
		if len(elements) != 0 {
			mt.Fatalf("expected empty filter document, got %s", filterValue.Document().String())
		}
	})

	mt.Run("returns mongo find error", func(mt *mtest.T) {
		repo := NewDeviceRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{Code: 13, Message: "find failed"}))

		_, err := repo.GetAllDevices(context.Background())
		if err == nil {
			mt.Fatalf("expected error, got nil")
		}
	})
}

func TestDeviceRepository_GetByID_UsesCorrectFilter(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("find by device_id", func(mt *mtest.T) {
		repo := NewDeviceRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.devices", mtest.FirstBatch,
			bson.D{{Key: "device_id", Value: "dev-9"}, {Key: "device_name", Value: "Camera 9"}, {Key: "device_status", Value: "active"}},
		))

		device, err := repo.GetByID(context.Background(), "dev-9")
		if err != nil {
			mt.Fatalf("expected nil error, got %v", err)
		}
		if device == nil || device.DeviceID != "dev-9" {
			mt.Fatalf("unexpected device mapping: %+v", device)
		}

		started := mt.GetStartedEvent()
		if started == nil {
			mt.Fatalf("expected started event")
		}
		filter := started.Command.Lookup("filter").Document()
		if got := filter.Lookup("device_id").StringValue(); got != "dev-9" {
			mt.Fatalf("expected filter device_id dev-9, got %q", got)
		}
	})

	mt.Run("returns no documents", func(mt *mtest.T) {
		repo := NewDeviceRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.devices", mtest.FirstBatch))

		_, err := repo.GetByID(context.Background(), "missing")
		if !errors.Is(err, mongo.ErrNoDocuments) {
			mt.Fatalf("expected ErrNoDocuments, got %v", err)
		}
	})
}

func TestDeviceRepository_Update_UsesCorrectQuery(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("updates by device_id with set", func(mt *mtest.T) {
		repo := NewDeviceRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := repo.Update(context.Background(), "dev-3", &domain.Device{DeviceID: "dev-3", DeviceStatus: "inactive"})
		if err != nil {
			mt.Fatalf("expected nil error, got %v", err)
		}

		started := mt.GetStartedEvent()
		if started == nil {
			mt.Fatalf("expected started event")
		}
		updates, err := started.Command.Lookup("updates").Array().Values()
		if err != nil {
			mt.Fatalf("expected updates array values, got error %v", err)
		}
		if len(updates) != 1 {
			mt.Fatalf("expected one update entry, got %d", len(updates))
		}
		entry := updates[0].Document()
		q := entry.Lookup("q").Document()
		if got := q.Lookup("device_id").StringValue(); got != "dev-3" {
			mt.Fatalf("expected update filter device_id dev-3, got %q", got)
		}
		u := entry.Lookup("u").Document()
		if u.Lookup("$set").Type != bsontype.EmbeddedDocument {
			mt.Fatalf("expected $set update document, got %s", u.Lookup("$set").Type)
		}
	})
}
