package service

import (
	"context"
	"reforest/internal/models"
	"reforest/internal/repository"
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

	AdoptTree(ctx context.Context, req *pb.AdoptTreeRequest, sponsorID string) (*models.Tree, error)
	GetTree(ctx context.Context, id primitive.ObjectID) (*models.Tree, error)
	ListTrees(ctx context.Context) ([]*models.Tree, error)
	UpdateTree(ctx context.Context, id primitive.ObjectID, req *pb.Tree) (*models.Tree, error)
	DeleteTree(ctx context.Context, id primitive.ObjectID) error

	CreateLog(ctx context.Context, req *pb.CreateLogRequest, adminID string) (*models.LogEntry, error)
	GetLogsByTreeID(ctx context.Context, treeID primitive.ObjectID) ([]*models.LogEntry, error)
	UpdateLog(ctx context.Context, id primitive.ObjectID, req *pb.LogEntry) (*models.LogEntry, error)
	DeleteLog(ctx context.Context, id primitive.ObjectID) error
}

type treeManagementService struct {
	repo          repository.TreeManagementRepository
	financeClient pb.FinanceServiceClient
}

func NewTreeManagementService(repo repository.TreeManagementRepository, financeClient pb.FinanceServiceClient) TreeManagementService {
	return &treeManagementService{repo: repo, financeClient: financeClient}
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

func (s *treeManagementService) AdoptTree(ctx context.Context, req *pb.AdoptTreeRequest, sponsorID string) (*models.Tree, error) {
	speciesID, err := primitive.ObjectIDFromHex(req.SpeciesId)
	if err != nil {
		return nil, models.ErrInvalidInput
	}
	plotID, err := primitive.ObjectIDFromHex(req.PlotId)
	if err != nil {
		return nil, models.ErrInvalidInput
	}

	tree := &models.Tree{
		SponsorID:           sponsorID,
		SpeciesID:           speciesID,
		PlotID:              plotID,
		CustomName:          req.CustomName,
		CurrentHeightMeters: 0,
		TotalFundedLifetime: 0,
		LastCareDate:        time.Now(),
		AdoptedAt:           time.Now(),
	}
	return s.repo.CreateTree(ctx, tree)
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