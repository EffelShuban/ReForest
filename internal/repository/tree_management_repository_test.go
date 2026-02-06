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

	mt.Run("duplicate", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(
			mtest.WriteError{Code: 11000, Message: "dup"},
		))

		_, err := repo.CreateSpecies(context.Background(), &models.Species{CommonName: "Jati"})
		if err != models.ErrAlreadyExists {
			t.Fatalf("expected ErrAlreadyExists, got %v", err)
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

func TestListSpecies(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("returns list", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		doc1 := bson.D{{Key: "_id", Value: primitive.NewObjectID()}, {Key: "common_name", Value: "Jati"}, {Key: "space_required_m2", Value: 5.0}, {Key: "price", Value: 100}}
		doc2 := bson.D{{Key: "_id", Value: primitive.NewObjectID()}, {Key: "common_name", Value: "Mahoni"}, {Key: "space_required_m2", Value: 6.0}, {Key: "price", Value: 120}}
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "db.species", mtest.FirstBatch, doc1, doc2))

		list, err := repo.ListSpecies(context.Background())
		if err != nil {
			t.Fatalf("ListSpecies error: %v", err)
		}
		if len(list) != 2 {
			t.Fatalf("expected 2 species, got %d", len(list))
		}
	})
}

func TestUpdateSpecies(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		id := primitive.NewObjectID()
		mt.AddMockResponses(
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)}, bson.E{Key: "nModified", Value: int32(1)}, bson.E{Key: "ok", Value: 1}),
			mtest.CreateCursorResponse(1, "db.species", mtest.FirstBatch, bson.D{{Key: "_id", Value: id}, {Key: "common_name", Value: "Jati"}, {Key: "space_required_m2", Value: 5.0}, {Key: "price", Value: 100}}),
		)

		got, err := repo.UpdateSpecies(context.Background(), &models.Species{ID: id, CommonName: "Jati"})
		if err != nil {
			t.Fatalf("UpdateSpecies error: %v", err)
		}
		if got.ID != id {
			t.Fatalf("expected id %v, got %v", id, got.ID)
		}
	})

	mt.Run("duplicate", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{Code: 11000, Message: "dup"}))

		_, err := repo.UpdateSpecies(context.Background(), &models.Species{ID: primitive.NewObjectID(), CommonName: "Jati"})
		if err != models.ErrAlreadyExists {
			t.Fatalf("expected ErrAlreadyExists, got %v", err)
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

func TestPlotOperations(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("create duplicate", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{Code: 11000, Message: "dup"}))

		_, err := repo.CreatePlot(context.Background(), &models.Plot{LocationName: "A"})
		if err != models.ErrAlreadyExists {
			t.Fatalf("expected ErrAlreadyExists, got %v", err)
		}
	})

	mt.Run("get not found", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "db.plots", mtest.FirstBatch))

		_, err := repo.GetPlot(context.Background(), primitive.NewObjectID())
		if err != models.ErrNotFound {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	mt.Run("update success", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		id := primitive.NewObjectID()
		mt.AddMockResponses(
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)}, bson.E{Key: "nModified", Value: int32(1)}, bson.E{Key: "ok", Value: 1}),
			mtest.CreateCursorResponse(1, "db.plots", mtest.FirstBatch, bson.D{{Key: "_id", Value: id}, {Key: "location_name", Value: "A"}, {Key: "address", Value: "Addr"}, {Key: "available_space_m2", Value: 10}}),
		)

		plot, err := repo.UpdatePlot(context.Background(), &models.Plot{ID: id, LocationName: "A"})
		if err != nil || plot.ID != id {
			t.Fatalf("UpdatePlot failed: %v", err)
		}
	})

	mt.Run("update duplicate", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{Code: 11000, Message: "dup"}))

		_, err := repo.UpdatePlot(context.Background(), &models.Plot{ID: primitive.NewObjectID()})
		if err != models.ErrAlreadyExists {
			t.Fatalf("expected ErrAlreadyExists, got %v", err)
		}
	})

	mt.Run("update not found", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(0)}, bson.E{Key: "nModified", Value: int32(0)}, bson.E{Key: "ok", Value: 1}))

		_, err := repo.UpdatePlot(context.Background(), &models.Plot{ID: primitive.NewObjectID()})
		if err != models.ErrNotFound {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	mt.Run("delete not found", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(0)}, bson.E{Key: "ok", Value: 1}))

		if err := repo.DeletePlot(context.Background(), primitive.NewObjectID()); err != models.ErrNotFound {
			t.Fatalf("expected ErrNotFound, got %v", err)
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

func TestGetAdoptionIntent(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		id := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "db.adoption_intents", mtest.FirstBatch, bson.D{{Key: "_id", Value: id}, {Key: "sponsor_id", Value: "user"}}))

		got, err := repo.GetAdoptionIntent(context.Background(), id)
		if err != nil || got.ID != id {
			t.Fatalf("GetAdoptionIntent failed: %v", err)
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

func TestTreeOperations(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("create success", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "insertedId", Value: primitive.NewObjectID()}))

		tree, err := repo.CreateTree(context.Background(), &models.Tree{CustomName: "My Tree"})
		if err != nil || tree.ID == primitive.NilObjectID {
			t.Fatalf("CreateTree failed: %v", err)
		}
	})

	mt.Run("create duplicate", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{Code: 11000, Message: "dup"}))

		_, err := repo.CreateTree(context.Background(), &models.Tree{CustomName: "Dup"})
		if err != models.ErrAlreadyExists {
			t.Fatalf("expected ErrAlreadyExists, got %v", err)
		}
	})

	mt.Run("update not found", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(0)}, bson.E{Key: "nModified", Value: int32(0)}, bson.E{Key: "ok", Value: 1}))

		_, err := repo.UpdateTree(context.Background(), &models.Tree{ID: primitive.NewObjectID()})
		if err != models.ErrNotFound {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	mt.Run("update duplicate", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{Code: 11000, Message: "dup"}))

		_, err := repo.UpdateTree(context.Background(), &models.Tree{ID: primitive.NewObjectID()})
		if err != models.ErrAlreadyExists {
			t.Fatalf("expected ErrAlreadyExists, got %v", err)
		}
	})

	mt.Run("list trees", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		doc := bson.D{{Key: "_id", Value: primitive.NewObjectID()}, {Key: "custom_name", Value: "T"}, {Key: "plot_id", Value: primitive.NewObjectID()}, {Key: "species_id", Value: primitive.NewObjectID()}}
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "db.trees", mtest.FirstBatch, doc))

		list, err := repo.ListTrees(context.Background())
		if err != nil || len(list) != 1 {
			t.Fatalf("ListTrees failed: %v", err)
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

func TestLogOperations(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("create success", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "insertedId", Value: primitive.NewObjectID()}))

		log, err := repo.CreateLog(context.Background(), &models.LogEntry{Activity: "check"})
		if err != nil || log.ID == primitive.NilObjectID {
			t.Fatalf("CreateLog failed: %v", err)
		}
	})

	mt.Run("create duplicate", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{Code: 11000, Message: "dup"}))

		_, err := repo.CreateLog(context.Background(), &models.LogEntry{Activity: "dup"})
		if err != models.ErrAlreadyExists {
			t.Fatalf("expected ErrAlreadyExists, got %v", err)
		}
	})

	mt.Run("update not found", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(0)}, bson.E{Key: "nModified", Value: int32(0)}, bson.E{Key: "ok", Value: 1}))

		_, err := repo.UpdateLog(context.Background(), &models.LogEntry{ID: primitive.NewObjectID()})
		if err != models.ErrNotFound {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	mt.Run("update duplicate", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{Code: 11000, Message: "dup"}))

		_, err := repo.UpdateLog(context.Background(), &models.LogEntry{ID: primitive.NewObjectID()})
		if err != models.ErrAlreadyExists {
			t.Fatalf("expected ErrAlreadyExists, got %v", err)
		}
	})

	mt.Run("delete not found", func(mt *mtest.T) {
		repo := NewTreeManagementRepository(mt.Client.Database("db"))
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(0)}, bson.E{Key: "ok", Value: 1}))

		if err := repo.DeleteLog(context.Background(), primitive.NewObjectID()); err != models.ErrNotFound {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}
