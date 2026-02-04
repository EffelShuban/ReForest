package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"reforest/internal/models"
	"reforest/pkg/pb"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type mockTreeRepo struct {
	createdTree   *models.Tree
	tree          *models.Tree
	logsByTree    map[primitive.ObjectID][]*models.LogEntry
	listTreesResp []*models.Tree
	updateTree    *models.Tree
}

func (m *mockTreeRepo) CreateSpecies(ctx context.Context, species *models.Species) (*models.Species, error) {
	return nil, nil
}
func (m *mockTreeRepo) GetSpecies(ctx context.Context, id primitive.ObjectID) (*models.Species, error) { return nil, nil }
func (m *mockTreeRepo) ListSpecies(ctx context.Context) ([]*models.Species, error) { return nil, nil }
func (m *mockTreeRepo) UpdateSpecies(ctx context.Context, species *models.Species) (*models.Species, error) {
	return nil, nil
}
func (m *mockTreeRepo) DeleteSpecies(ctx context.Context, id primitive.ObjectID) error { return nil }
func (m *mockTreeRepo) CreatePlot(ctx context.Context, plot *models.Plot) (*models.Plot, error) { return nil, nil }
func (m *mockTreeRepo) GetPlot(ctx context.Context, id primitive.ObjectID) (*models.Plot, error) { return nil, nil }
func (m *mockTreeRepo) ListPlots(ctx context.Context) ([]*models.Plot, error) { return nil, nil }
func (m *mockTreeRepo) UpdatePlot(ctx context.Context, plot *models.Plot) (*models.Plot, error) { return nil, nil }
func (m *mockTreeRepo) DeletePlot(ctx context.Context, id primitive.ObjectID) error { return nil }

func (m *mockTreeRepo) CreateTree(ctx context.Context, tree *models.Tree) (*models.Tree, error) {
	m.createdTree = tree
	if tree.ID.IsZero() {
		tree.ID = primitive.NewObjectID()
	}
	return tree, nil
}
func (m *mockTreeRepo) GetTree(ctx context.Context, id primitive.ObjectID) (*models.Tree, error) {
	return m.tree, nil
}
func (m *mockTreeRepo) ListTrees(ctx context.Context) ([]*models.Tree, error) {
	return m.listTreesResp, nil
}
func (m *mockTreeRepo) UpdateTree(ctx context.Context, tree *models.Tree) (*models.Tree, error) {
	m.updateTree = tree
	return tree, nil
}
func (m *mockTreeRepo) DeleteTree(ctx context.Context, id primitive.ObjectID) error { return nil }

func (m *mockTreeRepo) CreateLog(ctx context.Context, log *models.LogEntry) (*models.LogEntry, error) {
	return nil, nil
}
func (m *mockTreeRepo) GetLog(ctx context.Context, id primitive.ObjectID) (*models.LogEntry, error) { return nil, nil }
func (m *mockTreeRepo) GetLogsByTreeID(ctx context.Context, treeID primitive.ObjectID) ([]*models.LogEntry, error) {
	if m.logsByTree == nil {
		return nil, nil
	}
	return m.logsByTree[treeID], nil
}
func (m *mockTreeRepo) UpdateLog(ctx context.Context, log *models.LogEntry) (*models.LogEntry, error) {
	return nil, nil
}
func (m *mockTreeRepo) DeleteLog(ctx context.Context, id primitive.ObjectID) error { return nil }

type mockFinanceClient struct {
	txList *pb.TransactionList
	txErr  error
}

func (m *mockFinanceClient) CreateTransaction(ctx context.Context, in *pb.TransactionRequest, opts ...grpc.CallOption) (*pb.Transaction, error) {
	return nil, errors.New("not implemented")
}
func (m *mockFinanceClient) TopUpWallet(ctx context.Context, in *pb.TopUpRequest, opts ...grpc.CallOption) (*pb.Transaction, error) {
	return nil, errors.New("not implemented")
}
func (m *mockFinanceClient) HandleWalletWebhook(ctx context.Context, in *pb.WebhookRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, errors.New("not implemented")
}
func (m *mockFinanceClient) GetBalance(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.BalanceResponse, error) {
	return nil, errors.New("not implemented")
}
func (m *mockFinanceClient) GetTransactionHistory(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.TransactionList, error) {
	return m.txList, m.txErr
}
func (m *mockFinanceClient) CheckPaymentExpiry(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, errors.New("not implemented")
}

func TestTreeService_AdoptTree_InvalidIDs(t *testing.T) {
	repo := &mockTreeRepo{}
	svc := NewTreeManagementService(repo, &mockFinanceClient{})

	_, err := svc.AdoptTree(context.Background(), &pb.AdoptTreeRequest{
		SpeciesId: "bad",
		PlotId:    primitive.NewObjectID().Hex(),
	}, "sponsor")
	if err == nil || err != models.ErrInvalidInput {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestTreeService_AdoptTree_Success(t *testing.T) {
	repo := &mockTreeRepo{}
	svc := NewTreeManagementService(repo, &mockFinanceClient{})

	speciesID := primitive.NewObjectID()
	plotID := primitive.NewObjectID()

	tree, err := svc.AdoptTree(context.Background(), &pb.AdoptTreeRequest{
		SpeciesId: speciesID.Hex(),
		PlotId:    plotID.Hex(),
		CustomName: "My Tree",
	}, "sponsor-1")
	if err != nil {
		t.Fatalf("AdoptTree() error = %v", err)
	}

	if repo.createdTree == nil || repo.createdTree.SponsorID != "sponsor-1" {
		t.Fatalf("CreateTree was not called with sponsor")
	}
	if tree.SpeciesID != speciesID {
		t.Fatalf("species id not set correctly")
	}
}

func TestTreeService_GetTree_ComputesAggregates(t *testing.T) {
	treeID := primitive.NewObjectID()
	repo := &mockTreeRepo{
		tree: &models.Tree{
			ID: treeID,
		},
		logsByTree: map[primitive.ObjectID][]*models.LogEntry{
			treeID: {
				{CurrentHeightMeters: 1.1, RecordedAt: time.Now()},
				{CurrentHeightMeters: 2.2, RecordedAt: time.Now()},
			},
		},
	}
	financeClient := &mockFinanceClient{
		txList: &pb.TransactionList{
			Transactions: []*pb.Transaction{
				{Amount: 100},
				{Amount: 200},
			},
		},
	}
	svc := NewTreeManagementService(repo, financeClient)

	tree, err := svc.GetTree(context.Background(), treeID)
	if err != nil {
		t.Fatalf("GetTree() error = %v", err)
	}
	if tree.CurrentHeightMeters != 2.2 {
		t.Fatalf("expected max height 2.2, got %v", tree.CurrentHeightMeters)
	}
}

func TestTreeService_ListTrees_UsesLogsHeight(t *testing.T) {
	tree1 := &models.Tree{ID: primitive.NewObjectID()}
	tree2 := &models.Tree{ID: primitive.NewObjectID()}
	repo := &mockTreeRepo{
		listTreesResp: []*models.Tree{tree1, tree2},
		logsByTree: map[primitive.ObjectID][]*models.LogEntry{
			tree1.ID: {
				{CurrentHeightMeters: 1.5},
			},
		},
	}
	svc := NewTreeManagementService(repo, &mockFinanceClient{})

	trees, err := svc.ListTrees(context.Background())
	if err != nil {
		t.Fatalf("ListTrees() error = %v", err)
	}
	if trees[0].CurrentHeightMeters != 1.5 {
		t.Fatalf("tree1 height not populated")
	}
	if trees[1].CurrentHeightMeters != 0 {
		t.Fatalf("tree2 height should remain 0")
	}
}

func TestTreeService_UpdateTree_InvalidIDs(t *testing.T) {
	repo := &mockTreeRepo{}
	svc := NewTreeManagementService(repo, &mockFinanceClient{})

	_, err := svc.UpdateTree(context.Background(), primitive.NewObjectID(), &pb.Tree{
		SpeciesId: "bad",
		PlotId:    primitive.NewObjectID().Hex(),
	})
	if err == nil || err != models.ErrInvalidInput {
		t.Fatalf("expected invalid input error")
	}
}
