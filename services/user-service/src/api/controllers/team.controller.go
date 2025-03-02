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

// TeamController handles team-related requests
type TeamController struct {
	teamService *services.TeamService
	validator   *validator.Validate
}

// NewTeamController creates a new team controller
func NewTeamController(teamService *services.TeamService) *TeamController {
	return &TeamController{
		teamService: teamService,
		validator:   validator.New(),
	}
}

// GetTeam gets a team by ID
func (c *TeamController) GetTeam(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing team ID"})
		return
	}

	// Get team
	team, err := c.teamService.GetTeamByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("Failed to get team")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	// Return response
	includeMembers := ctx.Query("includeMembers") == "true"
	ctx.JSON(http.StatusOK, team.ToResponse(includeMembers))
}

// CreateTeam creates a new team
func (c *TeamController) CreateTeam(ctx *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserId(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse request
	var req models.CreateTeamRequest
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

	// Create team
	team, err := c.teamService.CreateTeam(ctx, req, userID)
	if err != nil {
		log.Error().Err(err).Interface("req", req).Msg("Failed to create team")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team", "message": err.Error()})
		return
	}

	// Return response
	ctx.JSON(http.StatusCreated, team.ToResponse(true))
}

// UpdateTeam updates a team
func (c *TeamController) UpdateTeam(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing team ID"})
		return
	}

	// Get user ID from context
	userID := middleware.GetUserId(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse request
	var req models.UpdateTeamRequest
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

	// Update team
	team, err := c.teamService.UpdateTeam(ctx, id, req, userID)
	if err != nil {
		log.Error().Err(err).Str("id", id).Interface("req", req).Msg("Failed to update team")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update team", "message": err.Error()})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, team.ToResponse(true))
}

// DeleteTeam deletes a team
func (c *TeamController) DeleteTeam(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing team ID"})
		return
	}

	// Get user ID from context
	userID := middleware.GetUserId(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Delete team
	err := c.teamService.DeleteTeam(ctx, id, userID)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("Failed to delete team")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete team", "message": err.Error()})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{"message": "Team deleted successfully"})
}

// GetTeamMembers gets team members
func (c *TeamController) GetTeamMembers(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing team ID"})
		return
	}

	// Get team
	team, err := c.teamService.GetTeamByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("Failed to get team for members")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{
		"teamId":      team.ID,
		"teamName":    team.Name,
		"memberCount": len(team.Members),
		"members":     team.Members,
	})
}

// AddTeamMember adds a member to a team
func (c *TeamController) AddTeamMember(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing team ID"})
		return
	}

	// Get user ID from context
	userID := middleware.GetUserId(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse request
	var req models.AddTeamMemberRequest
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

	// Add member
	err := c.teamService.AddTeamMember(ctx, id, req, userID)
	if err != nil {
		log.Error().Err(err).Str("id", id).Interface("req", req).Msg("Failed to add team member")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add team member", "message": err.Error()})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{"message": "Team member added successfully"})
}

// UpdateTeamMember updates a team member
func (c *TeamController) UpdateTeamMember(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing team ID"})
		return
	}

	memberID := ctx.Param("memberId")
	if memberID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing member ID"})
		return
	}

	// Get user ID from context
	userID := middleware.GetUserId(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse request
	var req models.UpdateTeamMemberRequest
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

	// Update member
	err := c.teamService.UpdateTeamMember(ctx, id, memberID, req, userID)
	if err != nil {
		log.Error().Err(err).Str("id", id).Str("memberId", memberID).Interface("req", req).Msg("Failed to update team member")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update team member", "message": err.Error()})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{"message": "Team member updated successfully"})
}

// RemoveTeamMember removes a member from a team
func (c *TeamController) RemoveTeamMember(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing team ID"})
		return
	}

	memberID := ctx.Param("memberId")
	if memberID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing member ID"})
		return
	}

	// Get user ID from context
	userID := middleware.GetUserId(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Remove member
	err := c.teamService.RemoveTeamMember(ctx, id, memberID, userID)
	if err != nil {
		log.Error().Err(err).Str("id", id).Str("memberId", memberID).Msg("Failed to remove team member")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove team member", "message": err.Error()})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{"message": "Team member removed successfully"})
}

// GetUserTeams gets teams by user
func (c *TeamController) GetUserTeams(ctx *gin.Context) {
	// Get user ID from context
	userID := middleware.GetUserId(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse pagination parameters
	pageStr := ctx.DefaultQuery("page", "1")
	limitStr := ctx.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	// Get teams
	teams, total, err := c.teamService.GetTeamsByUser(ctx, userID, page, limit)
	if err != nil {
		log.Error().Err(err).Str("userId", userID).Int("page", page).Int("limit", limit).
			Msg("Failed to get user teams")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user teams", "message": err.Error()})
		return
	}

	// Convert to response
	teamResponses := make([]models.TeamResponse, len(teams))
	for i, team := range teams {
		teamResponses[i] = team.ToResponse(false)
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{
		"teams":      teamResponses,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": (total + int64(limit) - 1) / int64(limit),
	})
}

// GetOrganizationTeams gets teams by organization
func (c *TeamController) GetOrganizationTeams(ctx *gin.Context) {
	orgID := ctx.Param("orgId")
	if orgID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing organization ID"})
		return
	}

	// Get user ID from context
	userID := middleware.GetUserId(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse pagination parameters
	pageStr := ctx.DefaultQuery("page", "1")
	limitStr := ctx.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	// Get teams
	teams, total, err := c.teamService.GetTeamsByOrganization(ctx, orgID, page, limit)
	if err != nil {
		log.Error().Err(err).Str("orgId", orgID).Int("page", page).Int("limit", limit).
			Msg("Failed to get organization teams")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organization teams", "message": err.Error()})
		return
	}

	// Convert to response
	teamResponses := make([]models.TeamResponse, len(teams))
	for i, team := range teams {
		teamResponses[i] = team.ToResponse(false)
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{
		"teams":      teamResponses,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": (total + int64(limit) - 1) / int64(limit),
	})
}
