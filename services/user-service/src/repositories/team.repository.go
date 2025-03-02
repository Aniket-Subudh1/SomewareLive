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

// TeamRepository is a repository for teams
type TeamRepository struct {
	collection *mongo.Collection
}

// NewTeamRepository creates a new team repository
func NewTeamRepository(mongoDB *db.MongoDB) *TeamRepository {
	return &TeamRepository{
		collection: mongoDB.GetCollection(db.TeamsCollection),
	}
}

// Create creates a new team
func (r *TeamRepository) Create(ctx context.Context, team *models.Team) error {
	// Check if team with the same name already exists in the organization
	existingTeam, err := r.GetByNameAndOrganization(ctx, team.Name, team.OrganizationID)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		log.Error().Err(err).Str("name", team.Name).Str("orgId", team.OrganizationID).
			Msg("Error checking existing team")
		return err
	}
	if existingTeam != nil {
		return errors.New("team with this name already exists in the organization")
	}

	// Create team
	result, err := r.collection.InsertOne(ctx, team)
	if err != nil {
		log.Error().Err(err).Interface("team", team).Msg("Error creating team")
		return err
	}

	// Update ID if inserted with a new ObjectId
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		team.ID = oid.Hex()
	}

	log.Debug().Str("id", team.ID).Str("name", team.Name).Msg("Team created")
	return nil
}

// GetByID gets a team by ID
func (r *TeamRepository) GetByID(ctx context.Context, id string) (*models.Team, error) {
	var team models.Team

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objID}
	err = r.collection.FindOne(ctx, filter).Decode(&team)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, mongo.ErrNoDocuments
		}
		log.Error().Err(err).Str("id", id).Msg("Error getting team by ID")
		return nil, err
	}

	return &team, nil
}

// GetByNameAndOrganization gets a team by name and organization ID
func (r *TeamRepository) GetByNameAndOrganization(ctx context.Context, name, organizationID string) (*models.Team, error) {
	var team models.Team

	filter := bson.M{"name": name, "organizationId": organizationID}
	err := r.collection.FindOne(ctx, filter).Decode(&team)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, mongo.ErrNoDocuments
		}
		log.Error().Err(err).Str("name", name).Str("orgId", organizationID).
			Msg("Error getting team by name and organization")
		return nil, err
	}

	return &team, nil
}

// GetTeamsByOrganization gets teams by organization ID
func (r *TeamRepository) GetTeamsByOrganization(ctx context.Context, organizationID string, page, limit int) ([]*models.Team, int64, error) {
	var teams []*models.Team

	// Build filter
	filter := bson.M{"organizationId": organizationID}

	// Count total
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		log.Error().Err(err).Str("orgId", organizationID).Msg("Error counting teams")
		return nil, 0, err
	}

	// Set options for pagination and sorting
	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"name": 1})

	// Find teams
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		log.Error().Err(err).Str("orgId", organizationID).Msg("Error finding teams")
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// Decode teams
	if err := cursor.All(ctx, &teams); err != nil {
		log.Error().Err(err).Msg("Error decoding teams")
		return nil, 0, err
	}

	return teams, total, nil
}

// GetTeamsByUser gets teams by user ID
func (r *TeamRepository) GetTeamsByUser(ctx context.Context, userID string, page, limit int) ([]*models.Team, int64, error) {
	var teams []*models.Team

	// Build filter for teams where the user is a member
	filter := bson.M{"members.userId": userID}

	// Count total
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		log.Error().Err(err).Str("userId", userID).Msg("Error counting user teams")
		return nil, 0, err
	}

	// Set options for pagination and sorting
	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"name": 1})

	// Find teams
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		log.Error().Err(err).Str("userId", userID).Msg("Error finding user teams")
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// Decode teams
	if err := cursor.All(ctx, &teams); err != nil {
		log.Error().Err(err).Msg("Error decoding user teams")
		return nil, 0, err
	}

	return teams, total, nil
}

// Update updates a team
func (r *TeamRepository) Update(ctx context.Context, team *models.Team) error {
	objID, err := primitive.ObjectIDFromHex(team.ID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objID}

	// Check if updating name and if new name conflicts with existing team
	existingTeam, err := r.GetByNameAndOrganization(ctx, team.Name, team.OrganizationID)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		log.Error().Err(err).Str("name", team.Name).Str("orgId", team.OrganizationID).
			Msg("Error checking team name conflict")
		return err
	}
	if existingTeam != nil && existingTeam.ID != team.ID {
		return errors.New("another team with this name already exists in the organization")
	}

	// Update team
	update := bson.M{
		"$set": bson.M{
			"name":        team.Name,
			"description": team.Description,
			"logoUrl":     team.LogoURL,
			"members":     team.Members,
			"updatedAt":   time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Error().Err(err).Str("id", team.ID).Msg("Error updating team")
		return err
	}

	log.Debug().Str("id", team.ID).Msg("Team updated")
	return nil
}

// Delete deletes a team
func (r *TeamRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objID}
	_, err = r.collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("Error deleting team")
		return err
	}

	log.Debug().Str("id", id).Msg("Team deleted")
	return nil
}

// AddMember adds a member to a team
func (r *TeamRepository) AddMember(ctx context.Context, teamID, userID string, role models.TeamMemberRole, invitedBy string) error {
	objID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return err
	}

	// Check if member already exists
	filter := bson.M{
		"_id":            objID,
		"members.userId": userID,
	}

	existingTeam, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		log.Error().Err(err).Str("teamId", teamID).Str("userId", userID).
			Msg("Error checking existing team member")
		return err
	}

	if existingTeam > 0 {
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
			log.Error().Err(err).Str("teamId", teamID).Str("userId", userID).
				Msg("Error updating team member role")
			return err
		}

		log.Debug().Str("teamId", teamID).Str("userId", userID).
			Str("role", string(role)).Msg("Team member role updated")
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
			log.Error().Err(err).Str("teamId", teamID).Str("userId", userID).
				Msg("Error adding team member")
			return err
		}

		log.Debug().Str("teamId", teamID).Str("userId", userID).
			Str("role", string(role)).Msg("Team member added")
	}

	return nil
}

// RemoveMember removes a member from a team
func (r *TeamRepository) RemoveMember(ctx context.Context, teamID, userID string) error {
	objID, err := primitive.ObjectIDFromHex(teamID)
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
		log.Error().Err(err).Str("teamId", teamID).Str("userId", userID).
			Msg("Error removing team member")
		return err
	}

	if result.ModifiedCount == 0 {
		return errors.New("member not found in team")
	}

	log.Debug().Str("teamId", teamID).Str("userId", userID).Msg("Team member removed")
	return nil
}
