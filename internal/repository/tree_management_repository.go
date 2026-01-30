package repository

import (
	"context"
	"errors"
	"reforest/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	speciesCollection = "species"
	plotCollection    = "plots"
	treeCollection    = "trees"
	logCollection     = "logs"
)

type TreeManagementRepository interface {
	CreateSpecies(ctx context.Context, species *models.Species) (*models.Species, error)
	GetSpecies(ctx context.Context, id primitive.ObjectID) (*models.Species, error)
	ListSpecies(ctx context.Context) ([]*models.Species, error)
	UpdateSpecies(ctx context.Context, species *models.Species) (*models.Species, error)
	DeleteSpecies(ctx context.Context, id primitive.ObjectID) error

	CreatePlot(ctx context.Context, plot *models.Plot) (*models.Plot, error)
	GetPlot(ctx context.Context, id primitive.ObjectID) (*models.Plot, error)
	ListPlots(ctx context.Context) ([]*models.Plot, error)
	UpdatePlot(ctx context.Context, plot *models.Plot) (*models.Plot, error)
	DeletePlot(ctx context.Context, id primitive.ObjectID) error

	CreateTree(ctx context.Context, tree *models.Tree) (*models.Tree, error)
	GetTree(ctx context.Context, id primitive.ObjectID) (*models.Tree, error)
	ListTrees(ctx context.Context) ([]*models.Tree, error)
	UpdateTree(ctx context.Context, tree *models.Tree) (*models.Tree, error)
	DeleteTree(ctx context.Context, id primitive.ObjectID) error

	CreateLog(ctx context.Context, log *models.LogEntry) (*models.LogEntry, error)
	GetLogsByTreeID(ctx context.Context, treeID primitive.ObjectID) ([]*models.LogEntry, error)
	UpdateLog(ctx context.Context, log *models.LogEntry) (*models.LogEntry, error)
	DeleteLog(ctx context.Context, id primitive.ObjectID) error
}

type treeManagementRepository struct {
	db *mongo.Database
}

func NewTreeManagementRepository(db *mongo.Database) TreeManagementRepository {
	return &treeManagementRepository{db: db}
}

func (r *treeManagementRepository) CreateSpecies(ctx context.Context, species *models.Species) (*models.Species, error) {
	res, err := r.db.Collection(speciesCollection).InsertOne(ctx, species)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, models.ErrAlreadyExists
		}
		return nil, err
	}
	species.ID = res.InsertedID.(primitive.ObjectID)
	return species, nil
}

func (r *treeManagementRepository) GetSpecies(ctx context.Context, id primitive.ObjectID) (*models.Species, error) {
	var species models.Species
	err := r.db.Collection(speciesCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&species)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, models.ErrNotFound
		}
		return nil, err
	}
	return &species, nil
}

func (r *treeManagementRepository) ListSpecies(ctx context.Context) ([]*models.Species, error) {
	cursor, err := r.db.Collection(speciesCollection).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var speciesList []*models.Species
	if err = cursor.All(ctx, &speciesList); err != nil {
		return nil, err
	}
	return speciesList, nil
}

func (r *treeManagementRepository) UpdateSpecies(ctx context.Context, species *models.Species) (*models.Species, error) {
	update := bson.M{
		"$set": bson.M{
			"common_name":       species.CommonName,
			"space_required_m2": species.SpaceRequiredM2,
			"price":             species.Price,
		},
	}
	res, err := r.db.Collection(speciesCollection).UpdateOne(ctx, bson.M{"_id": species.ID}, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, models.ErrAlreadyExists
		}
		return nil, err
	}
	if res.MatchedCount == 0 {
		return nil, models.ErrNotFound
	}
	return r.GetSpecies(ctx, species.ID)
}

func (r *treeManagementRepository) DeleteSpecies(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.db.Collection(speciesCollection).DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *treeManagementRepository) CreatePlot(ctx context.Context, plot *models.Plot) (*models.Plot, error) {
	res, err := r.db.Collection(plotCollection).InsertOne(ctx, plot)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, models.ErrAlreadyExists
		}
		return nil, err
	}
	plot.ID = res.InsertedID.(primitive.ObjectID)
	return plot, nil
}

func (r *treeManagementRepository) GetPlot(ctx context.Context, id primitive.ObjectID) (*models.Plot, error) {
	var plot models.Plot
	err := r.db.Collection(plotCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&plot)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, models.ErrNotFound
		}
		return nil, err
	}
	return &plot, nil
}

func (r *treeManagementRepository) ListPlots(ctx context.Context) ([]*models.Plot, error) {
	cursor, err := r.db.Collection(plotCollection).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var plots []*models.Plot
	if err = cursor.All(ctx, &plots); err != nil {
		return nil, err
	}
	return plots, nil
}

func (r *treeManagementRepository) UpdatePlot(ctx context.Context, plot *models.Plot) (*models.Plot, error) {
	update := bson.M{
		"$set": bson.M{
			"location_name":      plot.LocationName,
			"address":            plot.Address,
			"available_space_m2": plot.AvailableSpaceM2,
		},
	}
	res, err := r.db.Collection(plotCollection).UpdateOne(ctx, bson.M{"_id": plot.ID}, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, models.ErrAlreadyExists
		}
		return nil, err
	}
	if res.MatchedCount == 0 {
		return nil, models.ErrNotFound
	}
	return r.GetPlot(ctx, plot.ID)
}

func (r *treeManagementRepository) DeletePlot(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.db.Collection(plotCollection).DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *treeManagementRepository) CreateTree(ctx context.Context, tree *models.Tree) (*models.Tree, error) {
	res, err := r.db.Collection(treeCollection).InsertOne(ctx, tree)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, models.ErrAlreadyExists
		}
		return nil, err
	}
	tree.ID = res.InsertedID.(primitive.ObjectID)
	return tree, nil
}

func (r *treeManagementRepository) GetTree(ctx context.Context, id primitive.ObjectID) (*models.Tree, error) {
	var tree models.Tree
	err := r.db.Collection(treeCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&tree)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, models.ErrNotFound
		}
		return nil, err
	}
	return &tree, nil
}

func (r *treeManagementRepository) ListTrees(ctx context.Context) ([]*models.Tree, error) {
	cursor, err := r.db.Collection(treeCollection).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var trees []*models.Tree
	if err = cursor.All(ctx, &trees); err != nil {
		return nil, err
	}
	return trees, nil
}

func (r *treeManagementRepository) UpdateTree(ctx context.Context, tree *models.Tree) (*models.Tree, error) {
	update := bson.M{
		"$set": bson.M{
			"sponsor_id":            tree.SponsorID,
			"species_id":            tree.SpeciesID,
			"plot_id":               tree.PlotID,
			"custom_name":           tree.CustomName,
			"initial_height_meters": tree.InitialHeightMeters,
			"total_funded_lifetime": tree.TotalFundedLifetime,
			"last_care_date":        tree.LastCareDate,
			"adopted_at":            tree.AdoptedAt,
		},
	}
	res, err := r.db.Collection(treeCollection).UpdateOne(ctx, bson.M{"_id": tree.ID}, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, models.ErrAlreadyExists
		}
		return nil, err
	}
	if res.MatchedCount == 0 {
		return nil, models.ErrNotFound
	}
	return r.GetTree(ctx, tree.ID)
}

func (r *treeManagementRepository) DeleteTree(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.db.Collection(treeCollection).DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *treeManagementRepository) CreateLog(ctx context.Context, log *models.LogEntry) (*models.LogEntry, error) {
	res, err := r.db.Collection(logCollection).InsertOne(ctx, log)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, models.ErrAlreadyExists
		}
		return nil, err
	}
	log.ID = res.InsertedID.(primitive.ObjectID)
	return log, nil
}

func (r *treeManagementRepository) GetLogsByTreeID(ctx context.Context, treeID primitive.ObjectID) ([]*models.LogEntry, error) {
	cursor, err := r.db.Collection(logCollection).Find(ctx, bson.M{"adopted_tree_id": treeID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var logs []*models.LogEntry
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *treeManagementRepository) UpdateLog(ctx context.Context, log *models.LogEntry) (*models.LogEntry, error) {
	update := bson.M{
		"$set": bson.M{
			"adopted_tree_id":       log.AdoptedTreeID,
			"admin_id":              log.AdminID,
			"current_height_meters": log.CurrentHeightMeters,
			"activity":              log.Activity,
			"note":                  log.Note,
			"recorded_at":           log.RecordedAt,
		},
	}
	res, err := r.db.Collection(logCollection).UpdateOne(ctx, bson.M{"_id": log.ID}, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, models.ErrAlreadyExists
		}
		return nil, err
	}
	if res.MatchedCount == 0 {
		return nil, models.ErrNotFound
	}
	return log, nil
}

func (r *treeManagementRepository) DeleteLog(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.db.Collection(logCollection).DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return models.ErrNotFound
	}
	return nil
}