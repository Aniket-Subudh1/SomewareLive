package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/your-username/slido-clone/user-service/db"
	"github.com/your-username/slido-clone/user-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// OrganizationRepository is a repository for organizations
type OrganizationRepository struct {
	collection *mongo.Collection
}

// NewOrganizationRepository creates a new organization repository
func NewOrganizationRepository(mongoDB *db.MongoDB) *OrganizationRepository {
	return &OrganizationRepository{
		collection: mongoDB.GetCollection(db.OrganizationsCollection),
	}
}

// Create creates a new organization
func (r *OrganizationRepository) Create(ctx context.Context, org *models.Organization) error {
	// Check if organization with the same name already exists
	existingOrg, err := r.GetByName(ctx, org.Name)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		log.Error().Err(err).Str("name", org.Name).Msg("Error checking existing organization")
		return err
	}
	if existingOrg != nil {
		return errors.New("organization with this name already exists")
	}

	// Create organization
	result, err := r.collection.InsertOne(ctx, org)
	if err != nil {
		log.Error().Err(err).Interface("organization", org).Msg("Error creating organization")
		return err
	}

	// Update ID if inserted with a new ObjectId
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		org.ID = oid.Hex()
	}

	log.Debug().Str("id", org.ID).Str("name", org.Name).Msg("Organization created")
	return nil
}

// GetByID gets an organization by ID
func (r *OrganizationRepository) GetByID(ctx context.Context, id string) (*models.Organization, error) {
	var org models.Organization

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objID}
	err = r.collection.FindOne(ctx, filter).Decode(&org)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, mongo.ErrNoDocuments
		}
		log.Error().Err(err).Str("id", id).Msg("Error getting organization by ID")
		return nil, err
	}

	return &org, nil
}

// GetByName gets an organization by name
func (r *OrganizationRepository) GetByName(ctx context.Context, name string) (*models.Organization, error) {
	var org models.Organization

	filter := bson.M{"name": name}
	err := r.collection.FindOne(ctx, filter).Decode(&org)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, mongo.ErrNoDocuments
		}
		log.Error().Err(err).Str("name", name).Msg("Error getting organization by name")
		return nil, err
	}

	return &org, nil
}

// GetOrganizationsByUser gets organizations by user ID
func (r *OrganizationRepository) GetOrganizationsByUser(ctx context.Context, userID string, page, limit int) ([]*models.Organization, int64, error) {
	var orgs []*models.Organization

	// Build filter for organizations where the user is a member
	filter := bson.M{"members.userId": userID}

	// Count total
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		log.Error().Err(err).Str("userId", userID).Msg("Error counting user organizations")
		return nil, 0, err
	}

	// Set options for pagination and sorting
	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"name": 1})

	// Find organizations
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		log.Error().Err(err).Str("userId", userID).Msg("Error finding user organizations")
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// Decode organizations
	if err := cursor.All(ctx, &orgs); err != nil {
		log.Error().Err(err).Msg("Error decoding user organizations")
		return nil, 0, err
	}

	return orgs, total, nil
}

// ListOrganizations lists all organizations with pagination
func (r *OrganizationRepository) ListOrganizations(ctx context.Context, page, limit int) ([]*models.Organization, int64, error) {
	var orgs []*models.Organization

	// Build filter
	filter := bson.M{}

	// Count total
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		log.Error().Err(err).Msg("Error counting organizations")
		return nil, 0, err
	}

	// Set options for pagination and sorting
	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"name": 1})

	// Find organizations
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		log.Error().Err(err).Msg("Error finding organizations")
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// Decode organizations
	if err := cursor.All(ctx, &orgs); err != nil {
		log.Error().Err(err).Msg("Error decoding organizations")
		return nil, 0, err
	}

	return orgs, total, nil
}

// Update updates an organization
func (r *OrganizationRepository) Update(ctx context.Context, org *models.Organization) error {
	objID, err := primitive.ObjectIDFromHex(org.ID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objID}

	// Check if updating name and if new name conflicts with existing organization
	existingOrg, err := r.GetByName(ctx, org.Name)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		log.Error().Err(err).Str("name", org.Name).Msg("Error checking organization name conflict")
		return err
	}
	if existingOrg != nil && existingOrg.ID != org.ID {
		return errors.New("another organization with this name already exists")
	}

	// Update organization
	update := bson.M{
		"$set": bson.M{
			"name":        org.Name,
			"description": org.Description,
			"logoUrl":     org.LogoURL,
			"website":     org.Website,
			"industry":    org.Industry,
			"size":        org.Size,
			"location":    org.Location,
			"members":     org.Members,
			"settings":    org.Settings,
			"updatedAt":   time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Error().Err(err).Str("id", org.ID).Msg("Error updating organization")
		return err
	}

	log.Debug().Str("id", org.ID).Msg("Organization updated")
	return nil
}

// Delete deletes an organization
func (r *OrganizationRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objID}
	_, err = r.collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("Error deleting organization")
		return err
	}

	log.Debug().Str("id", id).Msg("Organization deleted")
	return nil
}

// AddMember adds a member to an organization
func (r *OrganizationRepository) AddMember(ctx context.Context, orgID, userID string, role models.OrganizationMemberRole, invitedBy string) error {
	objID, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return err
	}

	// Check if member already exists
	filter := bson.M{
		"_id":            objID,
		"members.userId": userID,
	}

	existingOrg, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		log.Error().Err(err).Str("orgId", orgID).Str("userId", userID).
			Msg("Error checking existing organization member")
		return err
	}

	if existingOrg > 0 {
		// Member already exists, update role
		filter = bson.M{"_id": objID, "members.userId": userID}
		update := bson.M{
			"$set": bson.M{
				"members.$.role": role,
				"updatedAt":      time.Now(),
			},
		}

		_, err = r.collection.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Error().Err(err).Str("orgId", orgID).Str("userId", userID).
				Msg("Error updating organization member role")
			return err
		}

		log.Debug().Str("orgId", orgID).Str("userId", userID).
			Str("role", string(role)).Msg("Organization member role updated")
	} else {
		// Add new member
		now := time.Now()
		filter = bson.M{"_id": objID}
		update := bson.M{
			"$push": bson.M{
				"members": bson.M{
					"userId":    userID,
					"role":      role,
					"joinedAt":  now,
					"invitedBy": invitedBy,
				},
			},
			"$set": bson.M{
				"updatedAt": now,
			},
		}

		_, err = r.collection.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Error().Err(err).Str("orgId", orgID).Str("userId", userID).
				Msg("Error adding organization member")
			return err
		}

		log.Debug().Str("orgId", orgID).Str("userId", userID).
			Str("role", string(role)).Msg("Organization member added")
	}

	return nil
}

// RemoveMember removes a member from an organization
func (r *OrganizationRepository) RemoveMember(ctx context.Context, orgID, userID string) error {
	objID, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objID}
	update := bson.M{
		"$pull": bson.M{
			"members": bson.M{"userId": userID},
		},
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Error().Err(err).Str("orgId", orgID).Str("userId", userID).
			Msg("Error removing organization member")
		return err
	}

	if result.ModifiedCount == 0 {
		return errors.New("member not found in organization")
	}

	log.Debug().Str("orgId", orgID).Str("userId", userID).Msg("Organization member removed")
	return nil
}

// AddTeam adds a team to an organization
func (r *OrganizationRepository) AddTeam(ctx context.Context, orgID, teamID string) error {
	objID, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objID}
	update := bson.M{
		"$addToSet": bson.M{
			"teamIds": teamID,
		},
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Error().Err(err).Str("orgId", orgID).Str("teamId", teamID).
			Msg("Error adding team to organization")
		return err
	}

	log.Debug().Str("orgId", orgID).Str("teamId", teamID).Msg("Team added to organization")
	return nil
}

// RemoveTeam removes a team from an organization
func (r *OrganizationRepository) RemoveTeam(ctx context.Context, orgID, teamID string) error {
	objID, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objID}
	update := bson.M{
		"$pull": bson.M{
			"teamIds": teamID,
		},
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Error().Err(err).Str("orgId", orgID).Str("teamId", teamID).
			Msg("Error removing team from organization")
		return err
	}

	if result.ModifiedCount == 0 {
		return errors.New("team not found in organization")
	}

	log.Debug().Str("orgId", orgID).Str("teamId", teamID).Msg("Team removed from organization")
	return nil
}
