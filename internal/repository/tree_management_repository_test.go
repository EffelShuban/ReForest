package repository

import (
	"context"
	"testing"
	"time"

	"reforest/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestCreateSpecies(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "insertedId", Value: primitive.NewObjectID()}))

		sp, err := repo.CreateSpecies(context.Background(), &models.Species{CommonName: "Jati"})
		if err != nil {
			t.Fatalf("CreateSpecies returned error: %v", err)
		}
		if sp.ID == primitive.NilObjectID {
			t.Fatalf("expected non-nil object id")
		}
	})
}

func TestGetSpecies_NotFound(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("not found", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "db.species", mtest.FirstBatch))

		_, err := repo.GetSpecies(context.Background(), primitive.NewObjectID())
		if err != models.ErrNotFound {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestUpdateSpecies_NotFound(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("matched zero", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: int32(0)},
			bson.E{Key: "nModified", Value: int32(0)},
			bson.E{Key: "ok", Value: 1},
		))

		_, err := repo.UpdateSpecies(context.Background(), &models.Species{ID: primitive.NewObjectID()})
		if err != models.ErrNotFound {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestDeleteSpecies(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)}, bson.E{Key: "ok", Value: 1}))

		if err := repo.DeleteSpecies(context.Background(), primitive.NewObjectID()); err != nil {
			t.Fatalf("DeleteSpecies returned error: %v", err)
		}
	})

	mt.Run("not found", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(0)}, bson.E{Key: "ok", Value: 1}))

		if err := repo.DeleteSpecies(context.Background(), primitive.NewObjectID()); err != models.ErrNotFound {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestListPlots(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("returns plots", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		doc1 := bson.D{{Key: "_id", Value: primitive.NewObjectID()}, {Key: "location_name", Value: "A"}, {Key: "address", Value: "Addr"}, {Key: "available_space_m2", Value: 10}}
		doc2 := bson.D{{Key: "_id", Value: primitive.NewObjectID()}, {Key: "location_name", Value: "B"}, {Key: "address", Value: "Addr2"}, {Key: "available_space_m2", Value: 20}}
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "db.plots", mtest.FirstBatch, doc1, doc2))

		list, err := repo.ListPlots(context.Background())
		if err != nil {
			t.Fatalf("ListPlots returned error: %v", err)
		}
		if len(list) != 2 {
			t.Fatalf("expected 2 plots, got %d", len(list))
		}
	})
}

func TestCreateAdoptionIntent(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "insertedId", Value: primitive.NewObjectID()}))

		intent, err := repo.CreateAdoptionIntent(context.Background(), &models.AdoptionIntent{SponsorID: "user"})
		if err != nil {
			t.Fatalf("CreateAdoptionIntent error: %v", err)
		}
		if intent.ID == primitive.NilObjectID {
			t.Fatalf("expected non-nil id")
		}
	})
}

func TestUpdateAdoptionIntentStatus(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)}, bson.E{Key: "ok", Value: 1}))

		if err := repo.UpdateAdoptionIntentStatus(context.Background(), primitive.NewObjectID(), "COMPLETED"); err != nil {
			t.Fatalf("UpdateAdoptionIntentStatus error: %v", err)
		}
	})
}

func TestGetTree_NotFound(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("not found", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "db.trees", mtest.FirstBatch))

		_, err := repo.GetTree(context.Background(), primitive.NewObjectID())
		if err != models.ErrNotFound {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestGetLogsByTreeID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("returns logs", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		treeID := primitive.NewObjectID()
		doc := bson.D{
			{Key: "_id", Value: primitive.NewObjectID()},
			{Key: "adopted_tree_id", Value: treeID},
			{Key: "admin_id", Value: "admin"},
			{Key: "current_height_meters", Value: 1.2},
			{Key: "activity", Value: "check"},
			{Key: "note", Value: "ok"},
			{Key: "recorded_at", Value: time.Now()},
		}
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "db.logs", mtest.FirstBatch, doc))

		logs, err := repo.GetLogsByTreeID(context.Background(), treeID)
		if err != nil {
			t.Fatalf("GetLogsByTreeID error: %v", err)
		}
		if len(logs) != 1 {
			t.Fatalf("expected 1 log, got %d", len(logs))
		}
		if logs[0].AdoptedTreeID != treeID {
			t.Fatalf("log tree id mismatch")
		}
	})
}
