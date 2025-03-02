package db

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/your-username/slido-clone/user-service/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDB is a MongoDB client instance
type MongoDB struct {
	Client *mongo.Client
	DB     *mongo.Database
}

// Collections represents the collection names
const (
	UsersCollection         = "users"
	TeamsCollection         = "teams"
	OrganizationsCollection = "organizations"
)

// New creates a new MongoDB client
func New(cfg *config.MongoDBConfig) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Create client options
	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to MongoDB")
		return nil, err
	}

	// Check the connection
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Error().Err(err).Msg("Failed to ping MongoDB")
		return nil, err
	}

	// Get the database
	db := client.Database(cfg.DBName)

	// Create indexes
	if err := createIndexes(ctx, db); err != nil {
		log.Error().Err(err).Msg("Failed to create indexes")
		return nil, err
	}

	log.Info().Str("database", cfg.DBName).Msg("Connected to MongoDB")

	return &MongoDB{
		Client: client,
		DB:     db,
	}, nil
}

// Close closes the MongoDB connection
func (m *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := m.Client.Disconnect(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to disconnect from MongoDB")
		return err
	}

	log.Info().Msg("Disconnected from MongoDB")
	return nil
}

// GetCollection returns a collection from the database
func (m *MongoDB) GetCollection(name string) *mongo.Collection {
	return m.DB.Collection(name)
}

// createIndexes creates indexes for the collections
func createIndexes(ctx context.Context, db *mongo.Database) error {
	// Users collection
	usersCollection := db.Collection(UsersCollection)
	userIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{
				"userId": 1,
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{
				"email": 1,
			},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err := usersCollection.Indexes().CreateMany(ctx, userIndexes)
	if err != nil {
		return err
	}

	// Teams collection
	teamsCollection := db.Collection(TeamsCollection)
	teamIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{
				"organizationId": 1,
			},
		},
		{
			Keys: map[string]interface{}{
				"name":           1,
				"organizationId": 1,
			},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err = teamsCollection.Indexes().CreateMany(ctx, teamIndexes)
	if err != nil {
		return err
	}

	// Organizations collection
	orgsCollection := db.Collection(OrganizationsCollection)
	orgIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{
				"name": 1,
			},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err = orgsCollection.Indexes().CreateMany(ctx, orgIndexes)

	return err
}
