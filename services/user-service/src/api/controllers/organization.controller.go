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

// OrganizationController handles organization-related requests
type OrganizationController struct {
	orgService *services.OrganizationService
	validator  *validator.Validate
}

// NewOrganizationController creates a new organization controller
func NewOrganizationController(orgService *services.OrganizationService) *OrganizationController {
	return &OrganizationController{
		orgService: orgService,
		validator:  validator.New(),
	}
}

// GetOrganization gets an organization by ID
func (c *OrganizationController) GetOrganization(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing organization ID"})
		return
	}

	// Get organization
	org, err := c.orgService.GetOrganizationByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("Failed to get organization")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	// Return response
	includeMembers := ctx.Query("includeMembers") == "true"
	includeSettings := ctx.Query("includeSettings") == "true"
	ctx.JSON(http.StatusOK, org.ToResponse(includeMembers, includeSettings))
}

// CreateOrganization creates a new organization
func (c *OrganizationController) CreateOrganization(ctx *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserId(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse request
	var req models.CreateOrganizationRequest
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

	// Create organization
	org, err := c.orgService.CreateOrganization(ctx, req, userID)
	if err != nil {
		log.Error().Err(err).Interface("req", req).Msg("Failed to create organization")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create organization", "message": err.Error()})
		return
	}

	// Return response
	ctx.JSON(http.StatusCreated, org.ToResponse(true, true))
}

// UpdateOrganization updates an organization
func (c *OrganizationController) UpdateOrganization(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing organization ID"})
		return
	}

	// Get user ID from context
	userID := middleware.GetUserId(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse request
	var req models.UpdateOrganizationRequest
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

	// Update organization
	org, err := c.orgService.UpdateOrganization(ctx, id, req, userID)
	if err != nil {
		log.Error().Err(err).Str("id", id).Interface("req", req).Msg("Failed to update organization")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update organization", "message": err.Error()})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, org.ToResponse(true, true))
}

// DeleteOrganization deletes an organization
func (c *OrganizationController) DeleteOrganization(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing organization ID"})
		return
	}

	// Get user ID from context
	userID := middleware.GetUserId(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Delete organization
	err := c.orgService.DeleteOrganization(ctx, id, userID)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("Failed to delete organization")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete organization", "message": err.Error()})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{"message": "Organization deleted successfully"})
}

// GetOrganizationMembers gets organization members
func (c *OrganizationController) GetOrganizationMembers(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing organization ID"})
		return
	}

	// Get organization
	org, err := c.orgService.GetOrganizationByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("Failed to get organization for members")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{
		"organizationId":   org.ID,
		"organizationName": org.Name,
		"memberCount":      len(org.Members),
		"members":          org.Members,
	})
}

// AddOrganizationMember adds a member to an organization
func (c *OrganizationController) AddOrganizationMember(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing organization ID"})
		return
	}

	// Get user ID from context
	userID := middleware.GetUserId(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse request
	var req models.AddOrganizationMemberRequest
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
	err := c.orgService.AddOrganizationMember(ctx, id, req, userID)
	if err != nil {
		log.Error().Err(err).Str("id", id).Interface("req", req).Msg("Failed to add organization member")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add organization member", "message": err.Error()})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{"message": "Organization member added successfully"})
}

// UpdateOrganizationMember updates an organization member
func (c *OrganizationController) UpdateOrganizationMember(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing organization ID"})
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
	var req models.UpdateOrganizationMemberRequest
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
	err := c.orgService.UpdateOrganizationMember(ctx, id, memberID, req, userID)
	if err != nil {
		log.Error().Err(err).Str("id", id).Str("memberId", memberID).Interface("req", req).Msg("Failed to update organization member")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update organization member", "message": err.Error()})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{"message": "Organization member updated successfully"})
}

// RemoveOrganizationMember removes a member from an organization
func (c *OrganizationController) RemoveOrganizationMember(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing organization ID"})
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
	err := c.orgService.RemoveOrganizationMember(ctx, id, memberID, userID)
	if err != nil {
		log.Error().Err(err).Str("id", id).Str("memberId", memberID).Msg("Failed to remove organization member")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove organization member", "message": err.Error()})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{"message": "Organization member removed successfully"})
}

// GetUserOrganizations gets organizations by user
func (c *OrganizationController) GetUserOrganizations(ctx *gin.Context) {
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

	// Get organizations
	orgs, total, err := c.orgService.GetOrganizationsByUser(ctx, userID, page, limit)
	if err != nil {
		log.Error().Err(err).Str("userId", userID).Int("page", page).Int("limit", limit).
			Msg("Failed to get user organizations")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user organizations", "message": err.Error()})
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
		"page":          page,
		"limit":         limit,
		"totalPages":    (total + int64(limit) - 1) / int64(limit),
	})
}

// GetOrganizationTeams gets teams in an organization
func (c *OrganizationController) GetOrganizationTeams(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
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
	teams, total, err := c.orgService.GetOrganizationTeams(ctx, id, page, limit, userID)
	if err != nil {
		log.Error().Err(err).Str("orgId", id).Int("page", page).Int("limit", limit).
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

// ListOrganizations lists all organizations
func (c *OrganizationController) ListOrganizations(ctx *gin.Context) {
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

	// Get organizations
	orgs, total, err := c.orgService.ListOrganizations(ctx, page, limit)
	if err != nil {
		log.Error().Err(err).Int("page", page).Int("limit", limit).
			Msg("Failed to list organizations")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list organizations", "message": err.Error()})
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
		"page":          page,
		"limit":         limit,
		"totalPages":    (total + int64(limit) - 1) / int64(limit),
	})
}
