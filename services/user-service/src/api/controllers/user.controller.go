package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/your-username/slido-clone/user-service/api/middleware"
	"github.com/your-username/slido-clone/user-service/models"
	"github.com/your-username/slido-clone/user-service/services"
)

// UserController handles user-related requests
type UserController struct {
	userService *services.UserService
	validator   *validator.Validate
}

// NewUserController creates a new user controller
func NewUserController(userService *services.UserService) *UserController {
	return &UserController{
		userService: userService,
		validator:   validator.New(),
	}
}

// GetUser gets a user by ID
func (c *UserController) GetUser(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing user ID"})
		return
	}

	// Get user
	user, err := c.userService.GetUserByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("Failed to get user")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, user.ToResponse())
}

// GetCurrentUser gets the current user
func (c *UserController) GetCurrentUser(ctx *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserId(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get user
	user, err := c.userService.GetUserByUserID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("userId", userID).Msg("Failed to get current user")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, user.ToResponse())
}

// CreateUser creates a new user
func (c *UserController) CreateUser(ctx *gin.Context) {
	// Parse request
	var req models.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate request
	if err := c.validator.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Validation error", "details": validationErrors.Error()})
		return
	}

	// Create user
	user, err := c.userService.CreateUser(ctx, req)
	if err != nil {
		log.Error().Err(err).Interface("req", req).Msg("Failed to create user")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "message": err.Error()})
		return
	}

	// Return response
	ctx.JSON(http.StatusCreated, user.ToResponse())
}

// UpdateUser updates a user
func (c *UserController) UpdateUser(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing user ID"})
		return
	}

	// Parse request
	var req models.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate request
	if err := c.validator.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Validation error", "details": validationErrors.Error()})
		return
	}

	// Update user
	user, err := c.userService.UpdateUser(ctx, id, req)
	if err != nil {
		log.Error().Err(err).Str("id", id).Interface("req", req).Msg("Failed to update user")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user", "message": err.Error()})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, user.ToResponse())
}

// UpdateCurrentUser updates the current user
func (c *UserController) UpdateCurrentUser(ctx *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserId(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get user
	user, err := c.userService.GetUserByUserID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("userId", userID).Msg("Failed to get current user for update")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Parse request
	var req models.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate request
	if err := c.validator.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Validation error", "details": validationErrors.Error()})
		return
	}

	// Update user
	updatedUser, err := c.userService.UpdateUser(ctx, user.ID, req)
	if err != nil {
		log.Error().Err(err).Str("id", user.ID).Interface("req", req).Msg("Failed to update current user")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user", "message": err.Error()})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, updatedUser.ToResponse())
}

// DeactivateUser deactivates a user
func (c *UserController) DeactivateUser(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing user ID"})
		return
	}

	// Deactivate user
	err := c.userService.DeactivateUser(ctx, id)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("Failed to deactivate user")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deactivate user", "message": err.Error()})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{"message": "User deactivated successfully"})
}

// ActivateUser activates a user
func (c *UserController) ActivateUser(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing user ID"})
		return
	}

	// Activate user
	err := c.userService.ActivateUser(ctx, id)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("Failed to activate user")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate user", "message": err.Error()})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{"message": "User activated successfully"})
}

// DeleteUser deletes a user
func (c *UserController) DeleteUser(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing user ID"})
		return
	}

	// Delete user
	err := c.userService.DeleteUser(ctx, id)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("Failed to delete user")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user", "message": err.Error()})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ListUsers lists users with pagination and filtering
func (c *UserController) ListUsers(ctx *gin.Context) {
	// Parse pagination parameters
	pageStr := ctx.DefaultQuery("page", "1")
	limitStr := ctx.DefaultQuery("limit", "20")
	search := ctx.DefaultQuery("search", "")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	// Get users
	users, total, err := c.userService.GetUsers(ctx, page, limit, search)
	if err != nil {
		log.Error().Err(err).Int("page", page).Int("limit", limit).Str("search", search).
			Msg("Failed to list users")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users", "message": err.Error()})
		return
	}

	// Convert to response
	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{
		"users":      userResponses,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": (total + int64(limit) - 1) / int64(limit),
	})
}
