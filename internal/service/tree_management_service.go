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

	plot, err := s.repo.GetPlot(ctx, plotID)
	if err != nil {
		return nil, "", "", err
	}

	if plot.AvailableSpaceM2 < species.SpaceRequiredM2 {
		return nil, "", "", fmt.Errorf("insufficient space in plot")
	}

	plot.AvailableSpaceM2 -= species.SpaceRequiredM2
	if _, err := s.repo.UpdatePlot(ctx, plot); err != nil {
		return nil, "", "", err
	}

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
		plot.AvailableSpaceM2 += species.SpaceRequiredM2
		s.repo.UpdatePlot(ctx, plot)
		return nil, "", "", err
	}

	tx, err := s.financeClient.CreateTransaction(ctx, &pb.TransactionRequest{
		UserId:      sponsorID,
		Amount:      int64(species.Price),
		Type:        pb.TransactionType_ADOPT,
		ReferenceId: intent.ID.Hex(),
	})
	if err != nil {
		plot.AvailableSpaceM2 += species.SpaceRequiredM2
		s.repo.UpdatePlot(ctx, plot)
		s.repo.UpdateAdoptionIntentStatus(ctx, intent.ID, "FAILED")
		return nil, "", "", err
	}

	return intent, tx.PaymentUrl, tx.InvoiceId, nil
}

func (s *treeManagementService) GetTree(ctx context.Context, id primitive.ObjectID) (*models.Tree, error) {
	tree, err := s.repo.GetTree(ctx, id)
	if err != nil {
		return nil, err
	}

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

	return tree, nil
}

func (s *treeManagementService) ListTrees(ctx context.Context) ([]*models.Tree, error) {
	trees, err := s.repo.ListTrees(ctx)
	if err != nil {
		return nil, err
	}

	for _, tree := range trees {
		logs, _ := s.repo.GetLogsByTreeID(ctx, tree.ID)
		for _, log := range logs {
			if log.CurrentHeightMeters > tree.CurrentHeightMeters {
				tree.CurrentHeightMeters = log.CurrentHeightMeters
			}
		}
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

		tree := &models.Tree{
			ID:                  intent.ID,
			SponsorID:           intent.SponsorID,
			SpeciesID:           intent.SpeciesID,
			PlotID:              intent.PlotID,
			CustomName:          intent.CustomName,
			LastCareDate:        time.Now(),
			AdoptedAt:           time.Now(),
		}
		createdTree, err := s.repo.CreateTree(context.Background(), tree)
		if err != nil { return err }

		_, _ = s.repo.CreateLog(context.Background(), &models.LogEntry{
			AdoptedTreeID:       createdTree.ID,
			Activity:            "Tree Planted",
			Note:                fmt.Sprintf("Tree adopted by %s", intent.SponsorID),
			RecordedAt:          time.Now(),
			CurrentHeightMeters: 0.5,
		})

		if err := s.repo.UpdateAdoptionIntentStatus(context.Background(), intent.ID, "COMPLETED"); err != nil {
			log.Printf("WARN: failed to update adoption intent %s status to COMPLETED: %v", intent.ID.Hex(), err)
		}

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

		intent, err := s.repo.GetAdoptionIntent(context.Background(), intentID)
		if err != nil {
			return err
		}

		if intent.Status == "PENDING" {
			if _, err := s.repo.GetTree(context.Background(), intent.ID); err == nil {
				log.Printf("WARN: Payment expired event for intent %s but tree exists. Ignoring.", intentID.Hex())
				return nil
			}

			species, err := s.repo.GetSpecies(context.Background(), intent.SpeciesID)
			if err != nil { return err }

			plot, err := s.repo.GetPlot(context.Background(), intent.PlotID)
			if err != nil { return err }

			plot.AvailableSpaceM2 += species.SpaceRequiredM2
			if _, err := s.repo.UpdatePlot(context.Background(), plot); err != nil { return err }
		}

		log.Printf("Payment expired for adoption intent %s. Marking as EXPIRED.", intentID.Hex())
		return s.repo.UpdateAdoptionIntentStatus(context.Background(), intentID, "EXPIRED")
	})

	if err != nil {
		log.Printf("Failed to start payment.expired consumer: %v", err)
	}
}