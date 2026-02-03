package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reforest/internal/models"
	"reforest/internal/repository"
	"reforest/pkg/mq"
	"reforest/pkg/pb"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TreeManagementService interface {
	CreateSpecies(ctx context.Context, req *pb.Species) (*models.Species, error)
	GetSpecies(ctx context.Context, id primitive.ObjectID) (*models.Species, error)
	ListSpecies(ctx context.Context) ([]*models.Species, error)
	UpdateSpecies(ctx context.Context, id primitive.ObjectID, req *pb.Species) (*models.Species, error)
	DeleteSpecies(ctx context.Context, id primitive.ObjectID) error

	CreatePlot(ctx context.Context, req *pb.Plot) (*models.Plot, error)
	GetPlot(ctx context.Context, id primitive.ObjectID) (*models.Plot, error)
	ListPlots(ctx context.Context) ([]*models.Plot, error)
	UpdatePlot(ctx context.Context, id primitive.ObjectID, req *pb.Plot) (*models.Plot, error)
	DeletePlot(ctx context.Context, id primitive.ObjectID) error

	AdoptTree(ctx context.Context, req *pb.AdoptTreeRequest, sponsorID string) (*models.AdoptionIntent, string, string, error)
	GetTree(ctx context.Context, id primitive.ObjectID) (*models.Tree, error)
	ListTrees(ctx context.Context) ([]*models.Tree, error)
	UpdateTree(ctx context.Context, id primitive.ObjectID, req *pb.Tree) (*models.Tree, error)
	DeleteTree(ctx context.Context, id primitive.ObjectID) error

	CreateLog(ctx context.Context, req *pb.CreateLogRequest, adminID string) (*models.LogEntry, error)
	GetLogsByTreeID(ctx context.Context, treeID primitive.ObjectID) ([]*models.LogEntry, error)
	UpdateLog(ctx context.Context, id primitive.ObjectID, req *pb.LogEntry) (*models.LogEntry, error)
	DeleteLog(ctx context.Context, id primitive.ObjectID) error
	StartConsumers()
}

type treeManagementService struct {
	repo          repository.TreeManagementRepository
	financeClient pb.FinanceServiceClient
	mqClient      *mq.Client
}

func NewTreeManagementService(repo repository.TreeManagementRepository, financeClient pb.FinanceServiceClient, mqClient *mq.Client) TreeManagementService {
	return &treeManagementService{repo: repo, financeClient: financeClient, mqClient: mqClient}
}

func (s *treeManagementService) CreateSpecies(ctx context.Context, req *pb.Species) (*models.Species, error) {
	species := &models.Species{
		CommonName:      req.CommonName,
		SpaceRequiredM2: float64(req.SpaceRequiredM2),
		Price:           req.Price,
	}
	return s.repo.CreateSpecies(ctx, species)
}

func (s *treeManagementService) GetSpecies(ctx context.Context, id primitive.ObjectID) (*models.Species, error) {
	return s.repo.GetSpecies(ctx, id)
}

func (s *treeManagementService) ListSpecies(ctx context.Context) ([]*models.Species, error) {
	return s.repo.ListSpecies(ctx)
}

func (s *treeManagementService) UpdateSpecies(ctx context.Context, id primitive.ObjectID, req *pb.Species) (*models.Species, error) {
	species := &models.Species{
		ID:              id,
		CommonName:      req.CommonName,
		SpaceRequiredM2: float64(req.SpaceRequiredM2),
		Price:           req.Price,
	}
	return s.repo.UpdateSpecies(ctx, species)
}

func (s *treeManagementService) DeleteSpecies(ctx context.Context, id primitive.ObjectID) error {
	return s.repo.DeleteSpecies(ctx, id)
}

func (s *treeManagementService) CreatePlot(ctx context.Context, req *pb.Plot) (*models.Plot, error) {
	plot := &models.Plot{
		LocationName:     req.LocationName,
		Address:          req.Address,
		AvailableSpaceM2: float64(req.AvailableSpaceM2),
	}
	return s.repo.CreatePlot(ctx, plot)
}

func (s *treeManagementService) GetPlot(ctx context.Context, id primitive.ObjectID) (*models.Plot, error) {
	return s.repo.GetPlot(ctx, id)
}

func (s *treeManagementService) ListPlots(ctx context.Context) ([]*models.Plot, error) {
	return s.repo.ListPlots(ctx)
}

func (s *treeManagementService) UpdatePlot(ctx context.Context, id primitive.ObjectID, req *pb.Plot) (*models.Plot, error) {
	plot := &models.Plot{
		ID:               id,
		LocationName:     req.LocationName,
		Address:          req.Address,
		AvailableSpaceM2: float64(req.AvailableSpaceM2),
	}
	return s.repo.UpdatePlot(ctx, plot)
}

func (s *treeManagementService) DeletePlot(ctx context.Context, id primitive.ObjectID) error {
	return s.repo.DeletePlot(ctx, id)
}

func (s *treeManagementService) AdoptTree(ctx context.Context, req *pb.AdoptTreeRequest, sponsorID string) (*models.AdoptionIntent, string, string, error) {
	speciesID, err := primitive.ObjectIDFromHex(req.SpeciesId)
	if err != nil {
		return nil, "", "", models.ErrInvalidInput
	}
	plotID, err := primitive.ObjectIDFromHex(req.PlotId)
	if err != nil {
		return nil, "", "", models.ErrInvalidInput
	}

	species, err := s.repo.GetSpecies(ctx, speciesID)
	if err != nil {
		return nil, "", "", err
	}

	// 1. Create Intent
	intent := &models.AdoptionIntent{
		SponsorID:  sponsorID,
		SpeciesID:  speciesID,
		PlotID:     plotID,
		CustomName: req.CustomName,
		Status:     "PENDING",
		CreatedAt:  time.Now(),
	}
	intent, err = s.repo.CreateAdoptionIntent(ctx, intent)
	if err != nil {
		return nil, "", "", err
	}

	// 2. Request Payment (Invoice)
	tx, err := s.financeClient.CreateTransaction(ctx, &pb.TransactionRequest{
		UserId:      sponsorID,
		Amount:      int64(species.Price),
		Type:        pb.TransactionType_ADOPT,
		ReferenceId: intent.ID.Hex(),
	})
	if err != nil {
		return nil, "", "", err
	}

	return intent, tx.PaymentUrl, tx.InvoiceId, nil
}

func (s *treeManagementService) GetTree(ctx context.Context, id primitive.ObjectID) (*models.Tree, error) {
	tree, err := s.repo.GetTree(ctx, id)
	if err != nil {
		return nil, err
	}

	// Populate CurrentHeightMeters from Logs
	logs, err := s.repo.GetLogsByTreeID(ctx, tree.ID)
	if err == nil && len(logs) > 0 {
		var maxHeight float64
		for _, log := range logs {
			if log.CurrentHeightMeters > maxHeight {
				maxHeight = log.CurrentHeightMeters
			}
		}
		tree.CurrentHeightMeters = maxHeight
	}

	txList, err := s.financeClient.GetTransactionHistory(ctx, &emptypb.Empty{})
	if err == nil && txList != nil {
		var totalFunded int32
		for _, tx := range txList.Transactions {
			totalFunded += int32(tx.Amount)
		}
		tree.TotalFundedLifetime = totalFunded
	}

	return tree, nil
}

func (s *treeManagementService) ListTrees(ctx context.Context) ([]*models.Tree, error) {
	trees, err := s.repo.ListTrees(ctx)
	if err != nil {
		return nil, err
	}

	// Populate computed fields for each tree
	// Note: In production, this N+1 query pattern should be optimized with batch fetching
	for _, tree := range trees {
		// Re-use GetTree logic or simplified version
		logs, _ := s.repo.GetLogsByTreeID(ctx, tree.ID)
		for _, log := range logs {
			if log.CurrentHeightMeters > tree.CurrentHeightMeters {
				tree.CurrentHeightMeters = log.CurrentHeightMeters
			}
		}
		// Skipping transaction fetch per tree for list to avoid excessive overhead
	}
	return trees, nil
}

func (s *treeManagementService) UpdateTree(ctx context.Context, id primitive.ObjectID, req *pb.Tree) (*models.Tree, error) {
	speciesID, err := primitive.ObjectIDFromHex(req.SpeciesId)
	if err != nil {
		return nil, models.ErrInvalidInput
	}
	plotID, err := primitive.ObjectIDFromHex(req.PlotId)
	if err != nil {
		return nil, models.ErrInvalidInput
	}

	tree := &models.Tree{
		ID:                  id,
		SponsorID:           req.SponsorId,
		SpeciesID:           speciesID,
		PlotID:              plotID,
		CustomName:          req.CustomName,
		CurrentHeightMeters: float64(req.CurrentHeightMeters), // Mapped from proto
		TotalFundedLifetime: req.TotalFundedLifetime,          // Mapped from proto
		LastCareDate:        req.LastCareDate.AsTime(),
		AdoptedAt:           req.AdoptedAt.AsTime(),
	}
	return s.repo.UpdateTree(ctx, tree)
}

func (s *treeManagementService) DeleteTree(ctx context.Context, id primitive.ObjectID) error {
	return s.repo.DeleteTree(ctx, id)
}

func (s *treeManagementService) CreateLog(ctx context.Context, req *pb.CreateLogRequest, adminID string) (*models.LogEntry, error) {
	treeID, err := primitive.ObjectIDFromHex(req.AdoptedTreeId)
	if err != nil {
		return nil, models.ErrInvalidInput
	}

	log := &models.LogEntry{
		AdoptedTreeID:       treeID,
		AdminID:             adminID,
		CurrentHeightMeters: float64(req.CurrentHeightMeters),
		Activity:            req.Activity,
		Note:                req.Note,
		RecordedAt:          time.Now(),
	}
	return s.repo.CreateLog(ctx, log)
}

func (s *treeManagementService) GetLogsByTreeID(ctx context.Context, treeID primitive.ObjectID) ([]*models.LogEntry, error) {
	return s.repo.GetLogsByTreeID(ctx, treeID)
}

func (s *treeManagementService) UpdateLog(ctx context.Context, id primitive.ObjectID, req *pb.LogEntry) (*models.LogEntry, error) {
	treeID, err := primitive.ObjectIDFromHex(req.AdoptedTreeId)
	if err != nil {
		return nil, models.ErrInvalidInput
	}

	log := &models.LogEntry{
		ID:                  id,
		AdoptedTreeID:       treeID,
		AdminID:             req.AdminId,
		CurrentHeightMeters: float64(req.CurrentHeightMeters),
		Activity:            req.Activity,
		Note:                req.Note,
		RecordedAt:          req.RecordedAt.AsTime(),
	}
	return s.repo.UpdateLog(ctx, log)
}

func (s *treeManagementService) DeleteLog(ctx context.Context, id primitive.ObjectID) error {
	return s.repo.DeleteLog(ctx, id)
}

func (s *treeManagementService) StartConsumers() {
	err := s.mqClient.Consume("payment.success", func(data []byte) error {
		var event struct {
			ReferenceID string `json:"reference_id"`
		}
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}
		
		intentID, err := primitive.ObjectIDFromHex(event.ReferenceID)
		if err != nil { return err }

		intent, err := s.repo.GetAdoptionIntent(context.Background(), intentID)
		if err != nil { return err }

		// Create the actual Tree
		tree := &models.Tree{
			SponsorID:           intent.SponsorID,
			SpeciesID:           intent.SpeciesID,
			PlotID:              intent.PlotID,
			CustomName:          intent.CustomName,
			CurrentHeightMeters: 0,
			TotalFundedLifetime: 0,
			LastCareDate:        time.Now(),
			AdoptedAt:           time.Now(),
		}
		createdTree, err := s.repo.CreateTree(context.Background(), tree)
		if err != nil { return err }

		// Create Initial Log
		_, _ = s.repo.CreateLog(context.Background(), &models.LogEntry{
			AdoptedTreeID:       createdTree.ID,
			Activity:            "Tree Planted",
			Note:                fmt.Sprintf("Tree adopted by %s", intent.SponsorID),
			RecordedAt:          time.Now(),
			CurrentHeightMeters: 0.5,
		})
		return nil
	})
	if err != nil {
		log.Printf("Failed to start consumers: %v", err)
	}

	err = s.mqClient.Consume("payment.expired", func(data []byte) error {
		var event struct {
			ReferenceID string `json:"reference_id"`
		}
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}

		intentID, err := primitive.ObjectIDFromHex(event.ReferenceID)
		if err != nil {
			return err
		}

		// Mark the intent as expired instead of deleting, which is better for auditing.
		log.Printf("Payment expired for adoption intent %s. Marking as EXPIRED.", intentID.Hex())
		return s.repo.UpdateAdoptionIntentStatus(context.Background(), intentID, "EXPIRED")
	})

	if err != nil {
		log.Printf("Failed to start payment.expired consumer: %v", err)
	}
}