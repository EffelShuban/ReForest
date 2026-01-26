package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Species struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	CommonName      string             `bson:"common_name"`
	SpaceRequiredM2 float64            `bson:"space_required_m2"`
	Price           int32              `bson:"price"`
}

type Plot struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	LocationName     string             `bson:"location_name"`
	Address          string             `bson:"address"`
	AvailableSpaceM2 float64            `bson:"available_space_m2"`
}

type Tree struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty"`
	SponsorID           string             `bson:"sponsor_id"`
	SpeciesID           primitive.ObjectID `bson:"species_id"`
	PlotID              primitive.ObjectID `bson:"plot_id"`
	CustomName          string             `bson:"custom_name"`
	InitialHeightMeters float64            `bson:"initial_height_meters"`
	TotalFundedLifetime int32              `bson:"total_funded_lifetime"`
	LastCareDate        time.Time          `bson:"last_care_date"`
	AdoptedAt           time.Time          `bson:"adopted_at"`
}

type LogEntry struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty"`
	AdoptedTreeID       primitive.ObjectID `bson:"adopted_tree_id"`
	AdminID             string             `bson:"admin_id"`
	CurrentHeightMeters float64            `bson:"current_height_meters"`
	Activity            string             `bson:"activity"`
	Note                string             `bson:"note"`
	RecordedAt          time.Time          `bson:"recorded_at"`
}