package service

import (
	"context"
	"reforest/internal/models"
	"reforest/internal/repository"
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
	repo repository.TreeManagementRepository
}

func NewTreeManagementService(repo repository.TreeManagementRepository) TreeManagementService {
	return &treeManagementService{repo: repo}
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
		return nil, err
	}
	plotID, err := primitive.ObjectIDFromHex(req.PlotId)
	if err != nil {
		return nil, err
	}

	tree := &models.Tree{
		SponsorID:           sponsorID,
		SpeciesID:           speciesID,
		PlotID:              plotID,
		CustomName:          req.CustomName,
		InitialHeightMeters: 0,
		TotalFundedLifetime: 0,
		LastCareDate:        time.Now(),
		AdoptedAt:           time.Now(),
	}
	return s.repo.CreateTree(ctx, tree)
}

func (s *treeManagementService) GetTree(ctx context.Context, id primitive.ObjectID) (*models.Tree, error) {
	return s.repo.GetTree(ctx, id)
}

func (s *treeManagementService) ListTrees(ctx context.Context) ([]*models.Tree, error) {
	return s.repo.ListTrees(ctx)
}

func (s *treeManagementService) UpdateTree(ctx context.Context, id primitive.ObjectID, req *pb.Tree) (*models.Tree, error) {
	speciesID, err := primitive.ObjectIDFromHex(req.SpeciesId)
	if err != nil {
		return nil, err
	}
	plotID, err := primitive.ObjectIDFromHex(req.PlotId)
	if err != nil {
		return nil, err
	}

	tree := &models.Tree{
		ID:                  id,
		SponsorID:           req.SponsorId,
		SpeciesID:           speciesID,
		PlotID:              plotID,
		CustomName:          req.CustomName,
		InitialHeightMeters: float64(req.InitialHeightMeters),
		TotalFundedLifetime: req.TotalFundedLifetime,
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
		return nil, err
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
		return nil, err
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