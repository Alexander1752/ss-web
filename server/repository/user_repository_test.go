package repository

import (
	"context"
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestUserRepository_Save_MapsFields(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("inserts email password and default role", func(mt *mtest.T) {
		repo := NewUserRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := repo.Save(context.Background(), "user@example.com", "hashed-password")
		if err != nil {
			mt.Fatalf("expected nil error, got %v", err)
		}

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
		if got := doc.Lookup("email").StringValue(); got != "user@example.com" {
			mt.Fatalf("expected email user@example.com, got %q", got)
		}
		if got := doc.Lookup("password").StringValue(); got != "hashed-password" {
			mt.Fatalf("expected password hashed-password, got %q", got)
		}
		if got := doc.Lookup("role").StringValue(); got != "user" {
			mt.Fatalf("expected role user, got %q", got)
		}
	})

	mt.Run("returns insert error", func(mt *mtest.T) {
		repo := NewUserRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{Code: 11000, Message: "duplicate"}))

		err := repo.Save(context.Background(), "dup@example.com", "pw")
		if err == nil {
			mt.Fatalf("expected error, got nil")
		}
	})
}

func TestUserRepository_FindByEmail_FilterAndErrorHandling(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("maps result and uses email filter", func(mt *mtest.T) {
		repo := NewUserRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.users", mtest.FirstBatch,
			bson.D{{Key: "email", Value: "alice@example.com"}, {Key: "password", Value: "hashed"}, {Key: "role", Value: "admin"}},
		))

		user, err := repo.FindByEmail(context.Background(), "alice@example.com")
		if err != nil {
			mt.Fatalf("expected nil error, got %v", err)
		}
		if user == nil || user.Email != "alice@example.com" || user.Role != "admin" {
			mt.Fatalf("unexpected user mapping: %+v", user)
		}

		started := mt.GetStartedEvent()
		if started == nil {
			mt.Fatalf("expected started event")
		}
		filter := started.Command.Lookup("filter").Document()
		if got := filter.Lookup("email").StringValue(); got != "alice@example.com" {
			mt.Fatalf("expected email filter alice@example.com, got %q", got)
		}
	})

	mt.Run("returns no documents", func(mt *mtest.T) {
		repo := NewUserRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.users", mtest.FirstBatch))

		_, err := repo.FindByEmail(context.Background(), "missing@example.com")
		if !errors.Is(err, mongo.ErrNoDocuments) {
			mt.Fatalf("expected ErrNoDocuments, got %v", err)
		}
	})
}
