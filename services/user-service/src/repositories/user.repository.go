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

// UserRepository is a repository for users
type UserRepository struct {
	collection *mongo.Collection
}

// NewUserRepository creates a new user repository
func NewUserRepository(mongoDB *db.MongoDB) *UserRepository {
	return &UserRepository{
		collection: mongoDB.GetCollection(db.UsersCollection),
	}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	// Check if user with the same userId or email already exists
	existingUser, err := r.GetByUserId(ctx, user.UserID)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		log.Error().Err(err).Str("userId", user.UserID).Msg("Error checking existing user by userId")
		return err
	}
	if existingUser != nil {
		return errors.New("user with this userId already exists")
	}

	existingUser, err = r.GetByEmail(ctx, user.Email)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		log.Error().Err(err).Str("email", user.Email).Msg("Error checking existing user by email")
		return err
	}
	if existingUser != nil {
		return errors.New("user with this email already exists")
	}

	// Create user
	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		log.Error().Err(err).Interface("user", user).Msg("Error creating user")
		return err
	}

	// Update ID if inserted with a new ObjectId
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		user.ID = oid.Hex()
	}

	log.Debug().Str("id", user.ID).Str("userId", user.UserID).Msg("User created")
	return nil
}

// GetByID gets a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objID}
	err = r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, mongo.ErrNoDocuments
		}
		log.Error().Err(err).Str("id", id).Msg("Error getting user by ID")
		return nil, err
	}

	return &user, nil
}

// GetByUserId gets a user by user ID
func (r *UserRepository) GetByUserId(ctx context.Context, userId string) (*models.User, error) {
	var user models.User

	filter := bson.M{"userId": userId}
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, mongo.ErrNoDocuments
		}
		log.Error().Err(err).Str("userId", userId).Msg("Error getting user by userId")
		return nil, err
	}

	return &user, nil
}

// GetByEmail gets a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	filter := bson.M{"email": email}
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, mongo.ErrNoDocuments
		}
		log.Error().Err(err).Str("email", email).Msg("Error getting user by email")
		return nil, err
	}

	return &user, nil
}

// GetUsers gets users with pagination and filtering
func (r *UserRepository) GetUsers(ctx context.Context, page, limit int, search string) ([]*models.User, int64, error) {
	var users []*models.User

	// Build filter
	filter := bson.M{}
	if search != "" {
		// Search by name or email
		filter = bson.M{
			"$or": []bson.M{
				{"firstName": bson.M{"$regex": search, "$options": "i"}},
				{"lastName": bson.M{"$regex": search, "$options": "i"}},
				{"email": bson.M{"$regex": search, "$options": "i"}},
			},
		}
	}

	// Count total
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		log.Error().Err(err).Msg("Error counting users")
		return nil, 0, err
	}

	// Set options for pagination and sorting
	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"lastName": 1, "firstName": 1})

	// Find users
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		log.Error().Err(err).Msg("Error finding users")
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// Decode users
	if err := cursor.All(ctx, &users); err != nil {
		log.Error().Err(err).Msg("Error decoding users")
		return nil, 0, err
	}

	return users, total, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	objID, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objID}
	update := bson.M{
		"$set": bson.M{
			"firstName":      user.FirstName,
			"lastName":       user.LastName,
			"status":         user.Status,
			"profilePicture": user.ProfilePicture,
			"bio":            user.Bio,
			"jobTitle":       user.JobTitle,
			"company":        user.Company,
			"location":       user.Location,
			"phone":          user.Phone,
			"website":        user.Website,
			"socialLinks":    user.SocialLinks,
			"preferences":    user.Preferences,
			"updatedAt":      time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Error().Err(err).Str("id", user.ID).Msg("Error updating user")
		return err
	}

	log.Debug().Str("id", user.ID).Msg("User updated")
	return nil
}

// UpdateLastLogin updates a user's last login time
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userId string, lastLogin time.Time) error {
	filter := bson.M{"userId": userId}
	update := bson.M{
		"$set": bson.M{
			"lastLogin": lastLogin,
			"updatedAt": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Error().Err(err).Str("userId", userId).Msg("Error updating user last login")
		return err
	}

	log.Debug().Str("userId", userId).Msg("User last login updated")
	return nil
}

// AddOrganizationToUser adds an organization to a user
func (r *UserRepository) AddOrganizationToUser(ctx context.Context, userId, organizationId string) error {
	filter := bson.M{"userId": userId}
	update := bson.M{
		"$addToSet": bson.M{"organizationIds": organizationId},
		"$set":      bson.M{"updatedAt": time.Now()},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Error().Err(err).Str("userId", userId).Str("organizationId", organizationId).
			Msg("Error adding organization to user")
		return err
	}

	log.Debug().Str("userId", userId).Str("organizationId", organizationId).
		Msg("Organization added to user")
	return nil
}

// RemoveOrganizationFromUser removes an organization from a user
func (r *UserRepository) RemoveOrganizationFromUser(ctx context.Context, userId, organizationId string) error {
	filter := bson.M{"userId": userId}
	update := bson.M{
		"$pull": bson.M{"organizationIds": organizationId},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Error().Err(err).Str("userId", userId).Str("organizationId", organizationId).
			Msg("Error removing organization from user")
		return err
	}

	log.Debug().Str("userId", userId).Str("organizationId", organizationId).
		Msg("Organization removed from user")
	return nil
}

// AddTeamToUser adds a team to a user
func (r *UserRepository) AddTeamToUser(ctx context.Context, userId, teamId string) error {
	filter := bson.M{"userId": userId}
	update := bson.M{
		"$addToSet": bson.M{"teamIds": teamId},
		"$set":      bson.M{"updatedAt": time.Now()},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Error().Err(err).Str("userId", userId).Str("teamId", teamId).
			Msg("Error adding team to user")
		return err
	}

	log.Debug().Str("userId", userId).Str("teamId", teamId).
		Msg("Team added to user")
	return nil
}

// RemoveTeamFromUser removes a team from a user
func (r *UserRepository) RemoveTeamFromUser(ctx context.Context, userId, teamId string) error {
	filter := bson.M{"userId": userId}
	update := bson.M{
		"$pull": bson.M{"teamIds": teamId},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Error().Err(err).Str("userId", userId).Str("teamId", teamId).
			Msg("Error removing team from user")
		return err
	}

	log.Debug().Str("userId", userId).Str("teamId", teamId).
		Msg("Team removed from user")
	return nil
}

// Delete deletes a user (soft delete by updating status)
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objID}
	update := bson.M{
		"$set": bson.M{
			"status":    models.StatusInactive,
			"updatedAt": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("Error deleting user")
		return err
	}

	log.Debug().Str("id", id).Msg("User deleted (soft delete)")
	return nil
}
