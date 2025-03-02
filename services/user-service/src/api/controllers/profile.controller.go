package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/your-username/slido-clone/user-service/api/middleware"
	"github.com/your-username/slido-clone/user-service/models"
	"github.com/your-username/slido-clone/user-service/services"
)

// ProfileController handles user profile-related requests
type ProfileController struct {
	userService *services.UserService
	teamService *services.TeamService
	orgService  *services.OrganizationService
	validator   *validator.Validate
}

// NewProfileController creates a new profile controller
func NewProfileController(
	userService *services.UserService,
	teamService *services.TeamService,
	orgService *services.OrganizationService,
) *ProfileController {
	return &ProfileController{
		userService: userService,
		teamService: teamService,
		orgService:  orgService,
		validator:   validator.New(),
	}
}

// GetProfile gets the current user's profile
func (c *ProfileController) GetProfile(ctx *gin.Context) {
	// Get user ID from context
	userID := middleware.GetUserId(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get user
	user, err := c.userService.GetUserByUserID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("userId", userID).Msg("Failed to get user profile")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, user.ToResponse())
}

// UpdateProfile updates the current user's profile
func (c *ProfileController) UpdateProfile(ctx *gin.Context) {
	// Get user ID from context
	userID := middleware.GetUserId(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get user
	user, err := c.userService.GetUserByUserID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("userId", userID).Msg("Failed to get user for profile update")
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

	// Update user profile
	updatedUser, err := c.userService.UpdateUser(ctx, user.ID, req)
	if err != nil {
		log.Error().Err(err).Str("userId", userID).Interface("req", req).Msg("Failed to update user profile")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile", "message": err.Error()})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, updatedUser.ToResponse())
}

// GetUserTeams gets the current user's teams
func (c *ProfileController) GetUserTeams(ctx *gin.Context) {
	// Get user ID from context
	userID := middleware.GetUserId(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse pagination parameters
	page := 1
	limit := 100 // Get all teams for profile view

	// Get teams
	teams, total, err := c.teamService.GetTeamsByUser(ctx, userID, page, limit)
	if err != nil {
		log.Error().Err(err).Str("userId", userID).Msg("Failed to get user teams")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get teams", "message": err.Error()})
		return
	}

	// Convert to response
	teamResponses := make([]models.TeamResponse, len(teams))
	for i, team := range teams {
		teamResponses[i] = team.ToResponse(false)
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{
		"teams": teamResponses,
		"total": total,
	})
}

// GetUserOrganizations gets the current user's organizations
func (c *ProfileController) GetUserOrganizations(ctx *gin.Context) {
	// Get user ID from context
	userID := middleware.GetUserId(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse pagination parameters
	page := 1
	limit := 100 // Get all organizations for profile view

	// Get organizations
	orgs, total, err := c.orgService.GetOrganizationsByUser(ctx, userID, page, limit)
	if err != nil {
		log.Error().Err(err).Str("userId", userID).Msg("Failed to get user organizations")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organizations", "message": err.Error()})
		return
	}

	// Convert to response
	orgResponses := make([]models.OrganizationResponse, len(orgs))
	for i, org := range orgs {
		orgResponses[i] = org.ToResponse(false, false)
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{
		"organizations": orgResponses,
		"total":         total,
	})
}

// GetFullProfile gets the current user's complete profile including teams and organizations
func (c *ProfileController) GetFullProfile(ctx *gin.Context) {
	// Get user ID from context
	userID := middleware.GetUserId(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get user
	user, err := c.userService.GetUserByUserID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("userId", userID).Msg("Failed to get user profile")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get teams
	teams, _, err := c.teamService.GetTeamsByUser(ctx, userID, 1, 100)
	if err != nil {
		log.Error().Err(err).Str("userId", userID).Msg("Failed to get user teams")
		teams = []*models.Team{} // Continue with empty teams
	}

	// Get organizations
	orgs, _, err := c.orgService.GetOrganizationsByUser(ctx, userID, 1, 100)
	if err != nil {
		log.Error().Err(err).Str("userId", userID).Msg("Failed to get user organizations")
		orgs = []*models.Organization{} // Continue with empty organizations
	}

	// Convert to response
	teamResponses := make([]models.TeamResponse, len(teams))
	for i, team := range teams {
		teamResponses[i] = team.ToResponse(false)
	}

	orgResponses := make([]models.OrganizationResponse, len(orgs))
	for i, org := range orgs {
		orgResponses[i] = org.ToResponse(false, false)
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{
		"user":          user.ToResponse(),
		"teams":         teamResponses,
		"organizations": orgResponses,
	})
}

// UpdateUserPreferences updates the current user's preferences
func (c *ProfileController) UpdateUserPreferences(ctx *gin.Context) {
	// Get user ID from context
	userID := middleware.GetUserId(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get user
	user, err := c.userService.GetUserByUserID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("userId", userID).Msg("Failed to get user for preferences update")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Parse request
	var req models.UpdatePreferences
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

	// Create update request with only preferences
	updateReq := models.UpdateUserRequest{
		Preferences: &req,
	}

	// Update user preferences
	updatedUser, err := c.userService.UpdateUser(ctx, user.ID, updateReq)
	if err != nil {
		log.Error().Err(err).Str("userId", userID).Interface("req", req).Msg("Failed to update user preferences")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update preferences", "message": err.Error()})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{
		"message":     "Preferences updated successfully",
		"preferences": updatedUser.Preferences,
	})
}
