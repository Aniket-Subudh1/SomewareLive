package services

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/your-username/slido-clone/user-service/models"
	"github.com/your-username/slido-clone/user-service/pkg/kafka"
	"github.com/your-username/slido-clone/user-service/repositories"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserService is a service for users
type UserService struct {
	userRepo *repositories.UserRepository
	producer *kafka.Producer
}

// NewUserService creates a new user service
func NewUserService(userRepo *repositories.UserRepository, producer *kafka.Producer) *UserService {
	return &UserService{
		userRepo: userRepo,
		producer: producer,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.User, error) {
	// Create user
	user := models.NewUser(req)

	// Save to database
	err := s.userRepo.Create(ctx, user)
	if err != nil {
		log.Error().Err(err).Interface("req", req).Msg("Failed to create user")
		return nil, err
	}

	// Publish event
	go func(u *models.User) {
		err := s.producer.PublishUserEvent(
			kafka.UserCreated,
			models.UserResponse{
				ID:             u.ID,
				Email:          u.Email,
				FirstName:      u.FirstName,
				LastName:       u.LastName,
				FullName:       u.FirstName + " " + u.LastName,
				Role:           u.Role,
				Status:         u.Status,
				ProfilePicture: u.ProfilePicture,
				CreatedAt:      u.CreatedAt,
			},
			u.ID,
			"",
		)
		if err != nil {
			log.Error().Err(err).Str("userId", u.UserID).Msg("Failed to publish user.created event")
		}
	}(user)

	return user, nil
}

// GetUserByID gets a user by ID
func (s *UserService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		log.Error().Err(err).Str("id", id).Msg("Failed to get user by ID")
		return nil, err
	}
	return user, nil
}

// GetUserByUserID gets a user by user ID
func (s *UserService) GetUserByUserID(ctx context.Context, userID string) (*models.User, error) {
	user, err := s.userRepo.GetByUserId(ctx, userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		log.Error().Err(err).Str("userId", userID).Msg("Failed to get user by user ID")
		return nil, err
	}
	return user, nil
}

// GetUserByEmail gets a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		log.Error().Err(err).Str("email", email).Msg("Failed to get user by email")
		return nil, err
	}
	return user, nil
}

// GetUsers gets users with pagination and filtering
func (s *UserService) GetUsers(ctx context.Context, page, limit int, search string) ([]*models.User, int64, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get users
	users, total, err := s.userRepo.GetUsers(ctx, page, limit, search)
	if err != nil {
		log.Error().Err(err).Int("page", page).Int("limit", limit).Str("search", search).
			Msg("Failed to get users")
		return nil, 0, err
	}

	return users, total, nil
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(ctx context.Context, id string, req models.UpdateUserRequest) (*models.User, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		log.Error().Err(err).Str("id", id).Msg("Failed to get user for update")
		return nil, err
	}

	// Apply changes
	user.Apply(req)

	// Save to database
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		log.Error().Err(err).Str("id", id).Interface("req", req).
			Msg("Failed to update user")
		return nil, err
	}

	// Publish event
	go func(u *models.User) {
		err := s.producer.PublishUserEvent(
			kafka.UserUpdated,
			models.UserResponse{
				ID:             u.ID,
				Email:          u.Email,
				FirstName:      u.FirstName,
				LastName:       u.LastName,
				FullName:       u.FirstName + " " + u.LastName,
				Role:           u.Role,
				Status:         u.Status,
				ProfilePicture: u.ProfilePicture,
				Bio:            u.Bio,
				JobTitle:       u.JobTitle,
				Company:        u.Company,
				Location:       u.Location,
				SocialLinks:    u.SocialLinks,
				LastLogin:      u.LastLogin,
				CreatedAt:      u.CreatedAt,
			},
			u.ID,
			"",
		)
		if err != nil {
			log.Error().Err(err).Str("userId", u.UserID).Msg("Failed to publish user.updated event")
		}
	}(user)

	return user, nil
}

// UpdateUserLastLogin updates a user's last login time
func (s *UserService) UpdateUserLastLogin(ctx context.Context, userID string) error {
	// Update last login
	now := time.Now()
	err := s.userRepo.UpdateLastLogin(ctx, userID, now)
	if err != nil {
		log.Error().Err(err).Str("userId", userID).Msg("Failed to update user last login")
		return err
	}
	return nil
}

// DeactivateUser deactivates a user
func (s *UserService) DeactivateUser(ctx context.Context, id string) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("user not found")
		}
		log.Error().Err(err).Str("id", id).Msg("Failed to get user for deactivation")
		return err
	}

	// Apply changes
	user.Status = models.StatusInactive
	user.UpdatedAt = time.Now()

	// Save to database
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("Failed to deactivate user")
		return err
	}

	// Publish event
	go func(u *models.User) {
		err := s.producer.PublishUserEvent(
			kafka.UserDeactivated,
			models.UserResponse{
				ID:        u.ID,
				Email:     u.Email,
				FirstName: u.FirstName,
				LastName:  u.LastName,
				FullName:  u.FirstName + " " + u.LastName,
				Role:      u.Role,
				Status:    u.Status,
				CreatedAt: u.CreatedAt,
			},
			u.ID,
			"",
		)
		if err != nil {
			log.Error().Err(err).Str("userId", u.UserID).Msg("Failed to publish user.deactivated event")
		}
	}(user)

	return nil
}

// ActivateUser activates a user
func (s *UserService) ActivateUser(ctx context.Context, id string) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("user not found")
		}
		log.Error().Err(err).Str("id", id).Msg("Failed to get user for activation")
		return err
	}

	// Apply changes
	user.Status = models.StatusActive
	user.UpdatedAt = time.Now()

	// Save to database
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("Failed to activate user")
		return err
	}

	// Publish event
	go func(u *models.User) {
		err := s.producer.PublishUserEvent(
			kafka.UserActivated,
			models.UserResponse{
				ID:        u.ID,
				Email:     u.Email,
				FirstName: u.FirstName,
				LastName:  u.LastName,
				FullName:  u.FirstName + " " + u.LastName,
				Role:      u.Role,
				Status:    u.Status,
				CreatedAt: u.CreatedAt,
			},
			u.ID,
			"",
		)
		if err != nil {
			log.Error().Err(err).Str("userId", u.UserID).Msg("Failed to publish user.activated event")
		}
	}(user)

	return nil
}

// DeleteUser deletes a user (soft delete)
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("user not found")
		}
		log.Error().Err(err).Str("id", id).Msg("Failed to get user for deletion")
		return err
	}

	// Delete user (soft delete)
	err = s.userRepo.Delete(ctx, id)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("Failed to delete user")
		return err
	}

	// Publish event
	go func(u *models.User) {
		err := s.producer.PublishUserEvent(
			kafka.UserDeleted,
			models.UserResponse{
				ID:        u.ID,
				Email:     u.Email,
				FirstName: u.FirstName,
				LastName:  u.LastName,
				FullName:  u.FirstName + " " + u.LastName,
				Role:      u.Role,
				Status:    u.Status,
				CreatedAt: u.CreatedAt,
			},
			u.ID,
			"",
		)
		if err != nil {
			log.Error().Err(err).Str("userId", u.UserID).Msg("Failed to publish user.deleted event")
		}
	}(user)

	return nil
}

// ProcessAuthUserCreated processes a user.created event from the Auth Service
func (s *UserService) ProcessAuthUserCreated(ctx context.Context, event kafka.Event) error {
	// Parse data
	var userData struct {
		ID        string          `json:"id"`
		Email     string          `json:"email"`
		FirstName string          `json:"firstName"`
		LastName  string          `json:"lastName"`
		Role      models.UserRole `json:"role"`
	}

	// Try to unmarshal the data
	data, ok := event.Data.(map[string]interface{})
	if !ok {
		log.Error().Interface("data", event.Data).Msg("Invalid data format for auth user.created event")
		return errors.New("invalid data format")
	}

	// Extract fields
	userId, _ := data["id"].(string)
	email, _ := data["email"].(string)
	firstName, _ := data["firstName"].(string)
	lastName, _ := data["lastName"].(string)
	roleStr, _ := data["role"].(string)

	// Validate required fields
	if userId == "" || email == "" || firstName == "" || lastName == "" {
		log.Error().Interface("data", data).Msg("Missing required fields for auth user.created event")
		return errors.New("missing required fields")
	}

	// Check if user already exists
	existingUser, err := s.userRepo.GetByUserId(ctx, userId)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		log.Error().Err(err).Str("userId", userId).Msg("Error checking existing user in auth event handler")
		return err
	}

	// If user already exists, do nothing
	if existingUser != nil {
		log.Info().Str("userId", userId).Msg("User already exists, skipping creation")
		return nil
	}

	// Create user request
	var role models.UserRole
	switch roleStr {
	case "admin":
		role = models.RoleAdmin
	case "presenter":
		role = models.RolePresenter
	default:
		role = models.RoleUser
	}

	createReq := models.CreateUserRequest{
		UserID:    userId,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
	}

	// Create user
	_, err = s.CreateUser(ctx, createReq)
	if err != nil {
		log.Error().Err(err).Interface("req", createReq).Msg("Failed to create user from auth event")
		return err
	}

	log.Info().Str("userId", userId).Msg("Created user from auth event")
	return nil
}
