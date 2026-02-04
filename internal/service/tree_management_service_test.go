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
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockMQ struct {
	triggers map[string]string
	calls    []string
}

func (m *mockMQ) Consume(routingKey string, handler func([]byte) error) error {
	m.calls = append(m.calls, routingKey)
	if payload, ok := m.triggers[routingKey]; ok {
		return handler([]byte(payload))
	}
	return nil
}

type mockTreeRepo struct {
	createdTree    *models.Tree
	tree           *models.Tree
	species        *models.Species
	speciesList    []*models.Species
	plot           *models.Plot
	plots          []*models.Plot
	logsByTree     map[primitive.ObjectID][]*models.LogEntry
	listTreesResp  []*models.Tree
	updateTree     *models.Tree
	intent         *models.AdoptionIntent
	intentStatus   string
	deletedSpecies primitive.ObjectID
	deletedPlot    primitive.ObjectID
	deletedTree    primitive.ObjectID
	deletedLog     primitive.ObjectID
	createLogInput *models.LogEntry
	updatedLog     *models.LogEntry
}

func (m *mockTreeRepo) CreateSpecies(ctx context.Context, species *models.Species) (*models.Species, error) {
	if species.ID.IsZero() {
		species.ID = primitive.NewObjectID()
	}
	m.species = species
	return species, nil
}
func (m *mockTreeRepo) GetSpecies(ctx context.Context, id primitive.ObjectID) (*models.Species, error) {
	if m.species != nil && m.species.ID == id {
		return m.species, nil
	}
	return nil, models.ErrNotFound
}
func (m *mockTreeRepo) ListSpecies(ctx context.Context) ([]*models.Species, error) {
	return m.speciesList, nil
}
func (m *mockTreeRepo) UpdateSpecies(ctx context.Context, species *models.Species) (*models.Species, error) {
	m.species = species
	return species, nil
}
func (m *mockTreeRepo) DeleteSpecies(ctx context.Context, id primitive.ObjectID) error {
	m.deletedSpecies = id
	return nil
}
func (m *mockTreeRepo) CreatePlot(ctx context.Context, plot *models.Plot) (*models.Plot, error) {
	if plot.ID.IsZero() {
		plot.ID = primitive.NewObjectID()
	}
	m.plot = plot
	return plot, nil
}
func (m *mockTreeRepo) GetPlot(ctx context.Context, id primitive.ObjectID) (*models.Plot, error) {
	if m.plot != nil && m.plot.ID == id {
		return m.plot, nil
	}
	return nil, models.ErrNotFound
}
func (m *mockTreeRepo) ListPlots(ctx context.Context) ([]*models.Plot, error) { return m.plots, nil }
func (m *mockTreeRepo) UpdatePlot(ctx context.Context, plot *models.Plot) (*models.Plot, error) {
	m.plot = plot
	return plot, nil
}
func (m *mockTreeRepo) DeletePlot(ctx context.Context, id primitive.ObjectID) error {
	m.deletedPlot = id
	return nil
}

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
func (m *mockTreeRepo) DeleteTree(ctx context.Context, id primitive.ObjectID) error {
	m.deletedTree = id
	return nil
}

func (m *mockTreeRepo) CreateLog(ctx context.Context, log *models.LogEntry) (*models.LogEntry, error) {
	if log.ID.IsZero() {
		log.ID = primitive.NewObjectID()
	}
	m.createLogInput = log
	return log, nil
}
func (m *mockTreeRepo) GetLog(ctx context.Context, id primitive.ObjectID) (*models.LogEntry, error) {
	return nil, nil
}
func (m *mockTreeRepo) GetLogsByTreeID(ctx context.Context, treeID primitive.ObjectID) ([]*models.LogEntry, error) {
	if m.logsByTree == nil {
		return nil, nil
	}
	return m.logsByTree[treeID], nil
}
func (m *mockTreeRepo) UpdateLog(ctx context.Context, log *models.LogEntry) (*models.LogEntry, error) {
	m.updatedLog = log
	return log, nil
}
func (m *mockTreeRepo) DeleteLog(ctx context.Context, id primitive.ObjectID) error {
	m.deletedLog = id
	return nil
}

func (m *mockTreeRepo) CreateAdoptionIntent(ctx context.Context, intent *models.AdoptionIntent) (*models.AdoptionIntent, error) {
	if intent.ID.IsZero() {
		intent.ID = primitive.NewObjectID()
	}
	m.intent = intent
	return intent, nil
}

func (m *mockTreeRepo) GetAdoptionIntent(ctx context.Context, id primitive.ObjectID) (*models.AdoptionIntent, error) {
	if m.intent != nil && m.intent.ID == id {
		return m.intent, nil
	}
	return nil, models.ErrNotFound
}

func (m *mockTreeRepo) UpdateAdoptionIntentStatus(ctx context.Context, id primitive.ObjectID, status string) error {
	if m.intent != nil && m.intent.ID == id {
		m.intent.Status = status
	}
	m.intentStatus = status
	return nil
}

type mockFinanceClient struct {
	txList      *pb.TransactionList
	txErr       error
	txResp      *pb.Transaction
	txErrCreate error
}

func (m *mockFinanceClient) CreateTransaction(ctx context.Context, in *pb.TransactionRequest, opts ...grpc.CallOption) (*pb.Transaction, error) {
	if m.txResp != nil {
		return m.txResp, m.txErrCreate
	}
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
	svc := NewTreeManagementService(repo, &mockFinanceClient{}, nil)

	_, _, _, err := svc.AdoptTree(context.Background(), &pb.AdoptTreeRequest{
		SpeciesId: "bad",
		PlotId:    primitive.NewObjectID().Hex(),
	}, "sponsor")
	if err == nil || err != models.ErrInvalidInput {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestTreeService_AdoptTree_Success(t *testing.T) {
	speciesID := primitive.NewObjectID()
	plotID := primitive.NewObjectID()

	repo := &mockTreeRepo{
		species: &models.Species{
			ID:    speciesID,
			Price: 2500,
		},
	}
	mockFin := &mockFinanceClient{
		txResp: &pb.Transaction{
			Id:         primitive.NewObjectID().Hex(),
			PaymentUrl: "https://pay.example.com/1",
			InvoiceId:  "inv-1",
		},
	}
	svc := NewTreeManagementService(repo, mockFin, nil)

	intent, paymentURL, invoiceID, err := svc.AdoptTree(context.Background(), &pb.AdoptTreeRequest{
		SpeciesId:  speciesID.Hex(),
		PlotId:     plotID.Hex(),
		CustomName: "My Tree",
	}, "sponsor-1")
	if err != nil {
		t.Fatalf("AdoptTree() error = %v", err)
	}

	if repo.intent == nil || repo.intent.SponsorID != "sponsor-1" {
		t.Fatalf("CreateAdoptionIntent was not called with sponsor")
	}
	if intent.SpeciesID != speciesID {
		t.Fatalf("species id not set correctly")
	}
	if paymentURL != "https://pay.example.com/1" || invoiceID != "inv-1" {
		t.Fatalf("payment details not forwarded correctly")
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
				{Amount: 100, ReferenceId: treeID.Hex()},
				{Amount: 200, ReferenceId: treeID.Hex()},
			},
		},
	}
	svc := NewTreeManagementService(repo, financeClient, nil)

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
	svc := NewTreeManagementService(repo, &mockFinanceClient{}, nil)

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
	svc := NewTreeManagementService(repo, &mockFinanceClient{}, nil)

	_, err := svc.UpdateTree(context.Background(), primitive.NewObjectID(), &pb.Tree{
		SpeciesId: "bad",
		PlotId:    primitive.NewObjectID().Hex(),
	})
	if err == nil || err != models.ErrInvalidInput {
		t.Fatalf("expected invalid input error")
	}
}

func TestTreeService_CreateSpecies_CRUD(t *testing.T) {
	repo := &mockTreeRepo{}
	svc := NewTreeManagementService(repo, &mockFinanceClient{}, nil)

	created, err := svc.CreateSpecies(context.Background(), &pb.Species{
		CommonName:      "Oak",
		SpaceRequiredM2: 2.5,
		Price:           100,
	})
	if err != nil {
		t.Fatalf("CreateSpecies error: %v", err)
	}
	if created.ID.IsZero() || repo.species.CommonName != "Oak" {
		t.Fatalf("species not created correctly")
	}

	repo.speciesList = []*models.Species{created}
	list, err := svc.ListSpecies(context.Background())
	if err != nil || len(list) != 1 {
		t.Fatalf("ListSpecies failed")
	}

	fetched, err := svc.GetSpecies(context.Background(), created.ID)
	if err != nil || fetched.ID != created.ID {
		t.Fatalf("GetSpecies failed")
	}

	updated, err := svc.UpdateSpecies(context.Background(), created.ID, &pb.Species{
		CommonName:      "Updated",
		SpaceRequiredM2: 3.1,
		Price:           200,
	})
	if err != nil {
		t.Fatalf("UpdateSpecies error: %v", err)
	}
	if updated.CommonName != "Updated" || repo.species.Price != 200 {
		t.Fatalf("UpdateSpecies did not persist")
	}

	if err := svc.DeleteSpecies(context.Background(), created.ID); err != nil {
		t.Fatalf("DeleteSpecies error: %v", err)
	}
	if repo.deletedSpecies != created.ID {
		t.Fatalf("DeleteSpecies did not pass id")
	}
}

func TestTreeService_Plot_CRUD(t *testing.T) {
	repo := &mockTreeRepo{}
	svc := NewTreeManagementService(repo, &mockFinanceClient{}, nil)

	p, err := svc.CreatePlot(context.Background(), &pb.Plot{
		LocationName:     "Loc",
		Address:          "Addr",
		AvailableSpaceM2: 9.5,
	})
	if err != nil {
		t.Fatalf("CreatePlot error: %v", err)
	}
	repo.plots = []*models.Plot{p}

	plist, _ := svc.ListPlots(context.Background())
	if len(plist) != 1 || plist[0].LocationName != "Loc" {
		t.Fatalf("ListPlots failed")
	}

	got, err := svc.GetPlot(context.Background(), p.ID)
	if err != nil || got.ID != p.ID {
		t.Fatalf("GetPlot failed")
	}

	updated, err := svc.UpdatePlot(context.Background(), p.ID, &pb.Plot{
		LocationName:     "New",
		Address:          "Addr2",
		AvailableSpaceM2: 5,
	})
	if err != nil || repo.plot.LocationName != "New" || updated.Address != "Addr2" {
		t.Fatalf("UpdatePlot failed")
	}

	if err := svc.DeletePlot(context.Background(), p.ID); err != nil {
		t.Fatalf("DeletePlot error: %v", err)
	}
	if repo.deletedPlot != p.ID {
		t.Fatalf("DeletePlot id mismatch")
	}
}

func TestTreeService_UpdateTree_Success(t *testing.T) {
	repo := &mockTreeRepo{}
	svc := NewTreeManagementService(repo, &mockFinanceClient{}, nil)

	id := primitive.NewObjectID()
	speciesID := primitive.NewObjectID()
	plotID := primitive.NewObjectID()
	lastCare := time.Now().Add(-24 * time.Hour)
	adopted := time.Now().Add(-48 * time.Hour)

	tree, err := svc.UpdateTree(context.Background(), id, &pb.Tree{
		SponsorId:    "s1",
		SpeciesId:    speciesID.Hex(),
		PlotId:       plotID.Hex(),
		CustomName:   "Tree",
		LastCareDate: timestamppb.New(lastCare),
		AdoptedAt:    timestamppb.New(adopted),
	})
	if err != nil {
		t.Fatalf("UpdateTree error: %v", err)
	}
	if repo.updateTree.ID != id || repo.updateTree.SpeciesID != speciesID || tree.CustomName != "Tree" {
		t.Fatalf("UpdateTree did not map fields")
	}
}

func TestTreeService_DeleteTree(t *testing.T) {
	repo := &mockTreeRepo{}
	svc := NewTreeManagementService(repo, &mockFinanceClient{}, nil)
	id := primitive.NewObjectID()
	if err := svc.DeleteTree(context.Background(), id); err != nil {
		t.Fatalf("DeleteTree error: %v", err)
	}
	if repo.deletedTree != id {
		t.Fatalf("delete id mismatch")
	}
}

func TestTreeService_CreateLog_SuccessAndInvalid(t *testing.T) {
	repo := &mockTreeRepo{}
	svc := NewTreeManagementService(repo, &mockFinanceClient{}, nil)

	if _, err := svc.CreateLog(context.Background(), &pb.CreateLogRequest{AdoptedTreeId: "bad"}, "admin"); err != models.ErrInvalidInput {
		t.Fatalf("expected invalid input")
	}

	treeID := primitive.NewObjectID()
	log, err := svc.CreateLog(context.Background(), &pb.CreateLogRequest{
		AdoptedTreeId:       treeID.Hex(),
		CurrentHeightMeters: 1.2,
		Activity:            "Watered",
		Note:                "Ok",
	}, "admin-1")
	if err != nil {
		t.Fatalf("CreateLog error: %v", err)
	}
	if repo.createLogInput == nil || repo.createLogInput.AdminID != "admin-1" || log.AdoptedTreeID != treeID {
		t.Fatalf("CreateLog not saved correctly")
	}
}

func TestTreeService_GetLogs_UpdateLog_DeleteLog(t *testing.T) {
	treeID := primitive.NewObjectID()
	repo := &mockTreeRepo{
		logsByTree: map[primitive.ObjectID][]*models.LogEntry{
			treeID: {
				{ID: primitive.NewObjectID(), CurrentHeightMeters: 1},
			},
		},
	}
	svc := NewTreeManagementService(repo, &mockFinanceClient{}, nil)

	logs, err := svc.GetLogsByTreeID(context.Background(), treeID)
	if err != nil || len(logs) != 1 {
		t.Fatalf("GetLogsByTreeID failed")
	}

	if _, err := svc.UpdateLog(context.Background(), primitive.NewObjectID(), &pb.LogEntry{AdoptedTreeId: "bad"}); err != models.ErrInvalidInput {
		t.Fatalf("expected invalid input for update")
	}

	newID := primitive.NewObjectID()
	updated, err := svc.UpdateLog(context.Background(), newID, &pb.LogEntry{
		AdoptedTreeId:       treeID.Hex(),
		AdminId:             "admin",
		CurrentHeightMeters: 2.3,
		Activity:            "Pruned",
		Note:                "Good",
		RecordedAt:          timestamppb.New(time.Now()),
	})
	if err != nil || repo.updatedLog == nil || repo.updatedLog.ID != newID || updated.Activity != "Pruned" {
		t.Fatalf("UpdateLog failed")
	}

	if err := svc.DeleteLog(context.Background(), newID); err != nil {
		t.Fatalf("DeleteLog error: %v", err)
	}
	if repo.deletedLog != newID {
		t.Fatalf("DeleteLog id mismatch")
	}
}
