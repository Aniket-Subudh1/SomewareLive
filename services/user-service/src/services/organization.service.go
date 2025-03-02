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

// OrganizationService is a service for organizations
type OrganizationService struct {
	orgRepo  *repositories.OrganizationRepository
	userRepo *repositories.UserRepository
	teamRepo *repositories.TeamRepository
	producer *kafka.Producer
}

// NewOrganizationService creates a new organization service
func NewOrganizationService(
	orgRepo *repositories.OrganizationRepository,
	userRepo *repositories.UserRepository,
	teamRepo *repositories.TeamRepository,
	producer *kafka.Producer,
) *OrganizationService {
	return &OrganizationService{
		orgRepo:  orgRepo,
		userRepo: userRepo,
		teamRepo: teamRepo,
		producer: producer,
	}
}

// CreateOrganization creates a new organization
func (s *OrganizationService) CreateOrganization(ctx context.Context, req models.CreateOrganizationRequest, createdBy string) (*models.Organization, error) {
	// Create organization
	org := models.NewOrganization(req, createdBy)

	// Save to database
	err := s.orgRepo.Create(ctx, org)
	if err != nil {
		log.Error().Err(err).Interface("req", req).Msg("Failed to create organization")
		return nil, err
	}

	// Add organization to creator's user profile
	err = s.userRepo.AddOrganizationToUser(ctx, createdBy, org.ID)
	if err != nil {
		log.Error().Err(err).Str("orgId", org.ID).Str("userId", createdBy).
			Msg("Failed to add organization to user")
		// Don't fail the organization creation, but log the error
	}

	// Publish event
	go func(o *models.Organization) {
		err := s.producer.PublishUserEvent(
			kafka.OrganizationCreated,
			o.ToResponse(false, false),
			o.ID,
			"",
		)
		if err != nil {
			log.Error().Err(err).Str("orgId", o.ID).Msg("Failed to publish organization.created event")
		}
	}(org)

	return org, nil
}

// GetOrganizationByID gets an organization by ID
func (s *OrganizationService) GetOrganizationByID(ctx context.Context, id string) (*models.Organization, error) {
	org, err := s.orgRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("organization not found")
		}
		log.Error().Err(err).Str("id", id).Msg("Failed to get organization by ID")
		return nil, err
	}
	return org, nil
}

// GetOrganizationsByUser gets organizations by user ID
func (s *OrganizationService) GetOrganizationsByUser(ctx context.Context, userID string, page, limit int) ([]*models.Organization, int64, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get organizations
	orgs, total, err := s.orgRepo.GetOrganizationsByUser(ctx, userID, page, limit)
	if err != nil {
		log.Error().Err(err).Str("userId", userID).Int("page", page).Int("limit", limit).
			Msg("Failed to get organizations by user")
		return nil, 0, err
	}

	return orgs, total, nil
}

// ListOrganizations lists all organizations with pagination
func (s *OrganizationService) ListOrganizations(ctx context.Context, page, limit int) ([]*models.Organization, int64, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get organizations
	orgs, total, err := s.orgRepo.ListOrganizations(ctx, page, limit)
	if err != nil {
		log.Error().Err(err).Int("page", page).Int("limit", limit).
			Msg("Failed to list organizations")
		return nil, 0, err
	}

	return orgs, total, nil
}

// UpdateOrganization updates an organization
func (s *OrganizationService) UpdateOrganization(ctx context.Context, id string, req models.UpdateOrganizationRequest, userID string) (*models.Organization, error) {
	// Get organization
	org, err := s.orgRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("organization not found")
		}
		log.Error().Err(err).Str("id", id).Msg("Failed to get organization for update")
		return nil, err
	}

	// Check permissions - must be admin or owner
	if !org.HasRole(userID, models.OrgRoleOwner, models.OrgRoleAdmin) {
		return nil, errors.New("insufficient permissions to update organization")
	}

	// Apply changes
	org.Apply(req)

	// Save to database
	err = s.orgRepo.Update(ctx, org)
	if err != nil {
		log.Error().Err(err).Str("id", id).Interface("req", req).
			Msg("Failed to update organization")
		return nil, err
	}

	// Publish event
	go func(o *models.Organization) {
		err := s.producer.PublishUserEvent(
			kafka.OrganizationUpdated,
			o.ToResponse(false, true),
			o.ID,
			"",
		)
		if err != nil {
			log.Error().Err(err).Str("orgId", o.ID).Msg("Failed to publish organization.updated event")
		}
	}(org)

	return org, nil
}

// DeleteOrganization deletes an organization
func (s *OrganizationService) DeleteOrganization(ctx context.Context, id string, userID string) error {
	// Get organization
	org, err := s.orgRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("organization not found")
		}
		log.Error().Err(err).Str("id", id).Msg("Failed to get organization for deletion")
		return err
	}

	// Check permissions - must be owner
	if !org.HasRole(userID, models.OrgRoleOwner) {
		return errors.New("insufficient permissions to delete organization")
	}

	// Delete all teams in the organization
	for _, teamID := range org.TeamIDs {
		err = s.teamRepo.Delete(ctx, teamID)
		if err != nil {
			log.Error().Err(err).Str("teamId", teamID).Str("orgId", id).
				Msg("Failed to delete team during organization deletion")
			// Continue with other teams
		}
	}

	// Delete organization
	err = s.orgRepo.Delete(ctx, id)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("Failed to delete organization")
		return err
	}

	// Remove organization from all members
	for _, member := range org.Members {
		err = s.userRepo.RemoveOrganizationFromUser(ctx, member.UserID, id)
		if err != nil {
			log.Error().Err(err).Str("orgId", id).Str("userId", member.UserID).
				Msg("Failed to remove organization from user")
			// Don't fail the organization deletion, but log the error
		}
	}

	// Publish event
	go func(o *models.Organization) {
		err := s.producer.PublishUserEvent(
			kafka.OrganizationDeleted,
			models.OrganizationResponse{
				ID:   o.ID,
				Name: o.Name,
			},
			o.ID,
			"",
		)
		if err != nil {
			log.Error().Err(err).Str("orgId", o.ID).Msg("Failed to publish organization.deleted event")
		}
	}(org)

	return nil
}

// AddOrganizationMember adds a member to an organization
func (s *OrganizationService) AddOrganizationMember(ctx context.Context, orgID string, req models.AddOrganizationMemberRequest, invitedBy string) error {
	// Get organization
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("organization not found")
		}
		log.Error().Err(err).Str("id", orgID).Msg("Failed to get organization for adding member")
		return err
	}

	// Check permissions - must be admin or owner
	if !org.HasRole(invitedBy, models.OrgRoleOwner, models.OrgRoleAdmin) {
		return errors.New("insufficient permissions to add organization member")
	}

	// Verify user exists
	user, err := s.userRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("user not found")
		}
		log.Error().Err(err).Str("userId", req.UserID).Msg("Failed to get user for adding to organization")
		return err
	}

	// Add member to organization
	err = s.orgRepo.AddMember(ctx, orgID, req.UserID, req.Role, invitedBy)
	if err != nil {
		log.Error().Err(err).Str("orgId", orgID).Str("userId", req.UserID).
			Msg("Failed to add member to organization")
		return err
	}

	// Add organization to user
	err = s.userRepo.AddOrganizationToUser(ctx, req.UserID, orgID)
	if err != nil {
		log.Error().Err(err).Str("orgId", orgID).Str("userId", req.UserID).
			Msg("Failed to add organization to user")
		// Don't fail the operation, but log the error
	}

	// Refresh organization data
	org, err = s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		log.Error().Err(err).Str("id", orgID).Msg("Failed to refresh organization data after adding member")
		// Don't fail the operation, but log the error
	}

	// Publish event
	go func(o *models.Organization, userID string, role models.OrganizationMemberRole) {
		if o == nil {
			return
		}

		// Find the added member
		var addedMember *models.OrganizationMember
		for _, m := range o.Members {
			if m.UserID == userID {
				addedMember = &m
				break
			}
		}

		if addedMember == nil {
			log.Error().Str("orgId", orgID).Str("userId", userID).
				Msg("Failed to find added member for event")
			return
		}

		err := s.producer.PublishUserEvent(
			kafka.OrganizationMemberAdded,
			map[string]interface{}{
				"orgId":     o.ID,
				"orgName":   o.Name,
				"userId":    userID,
				"userEmail": user.Email,
				"userName":  user.FirstName + " " + user.LastName,
				"role":      role,
				"invitedBy": invitedBy,
				"joinedAt":  addedMember.JoinedAt,
			},
			o.ID,
			"",
		)
		if err != nil {
			log.Error().Err(err).Str("orgId", o.ID).Str("userId", userID).
				Msg("Failed to publish organization.member.added event")
		}
	}(org, req.UserID, req.Role)

	return nil
}

// UpdateOrganizationMember updates an organization member's role
func (s *OrganizationService) UpdateOrganizationMember(ctx context.Context, orgID, memberID string, req models.UpdateOrganizationMemberRequest, updatedBy string) error {
	// Get organization
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("organization not found")
		}
		log.Error().Err(err).Str("id", orgID).Msg("Failed to get organization for updating member")
		return err
	}

	// Check permissions - must be admin or owner
	if !org.HasRole(updatedBy, models.OrgRoleOwner, models.OrgRoleAdmin) {
		return errors.New("insufficient permissions to update organization member")
	}

	// If updating an owner, only an owner can do that
	currentMember := org.GetMember(memberID)
	if currentMember != nil && currentMember.Role == models.OrgRoleOwner && !org.HasRole(updatedBy, models.OrgRoleOwner) {
		return errors.New("only an organization owner can change the role of another owner")
	}

	// Check if the user is trying to update their own role to a lower one
	if memberID == updatedBy && (currentMember.Role == models.OrgRoleOwner && req.Role != models.OrgRoleOwner) {
		// Count owners
		ownerCount := 0
		for _, m := range org.Members {
			if m.Role == models.OrgRoleOwner {
				ownerCount++
			}
		}

		// If this is the only owner, don't allow role change
		if ownerCount <= 1 {
			return errors.New("cannot change role: organization must have at least one owner")
		}
	}

	// Update member role
	err = s.orgRepo.AddMember(ctx, orgID, memberID, req.Role, updatedBy)
	if err != nil {
		log.Error().Err(err).Str("orgId", orgID).Str("userId", memberID).
			Msg("Failed to update organization member")
		return err
	}

	// Refresh organization data
	org, err = s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		log.Error().Err(err).Str("id", orgID).Msg("Failed to refresh organization data after updating member")
		// Don't fail the operation, but log the error
	}

	// Publish event
	go func(o *models.Organization, userID string, role models.OrganizationMemberRole) {
		if o == nil {
			return
		}

		err := s.producer.PublishUserEvent(
			kafka.OrganizationMemberUpdated,
			map[string]interface{}{
				"orgId":     o.ID,
				"orgName":   o.Name,
				"userId":    userID,
				"role":      role,
				"updatedBy": updatedBy,
				"updatedAt": time.Now(),
			},
			o.ID,
			"",
		)
		if err != nil {
			log.Error().Err(err).Str("orgId", o.ID).Str("userId", userID).
				Msg("Failed to publish organization.member.updated event")
		}
	}(org, memberID, req.Role)

	return nil
}

// RemoveOrganizationMember removes a member from an organization
func (s *OrganizationService) RemoveOrganizationMember(ctx context.Context, orgID, memberID string, removedBy string) error {
	// Get organization
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("organization not found")
		}
		log.Error().Err(err).Str("id", orgID).Msg("Failed to get organization for removing member")
		return err
	}

	// Get member being removed
	memberToRemove := org.GetMember(memberID)
	if memberToRemove == nil {
		return errors.New("member not found in organization")
	}

	// Check permissions
	// 1. Organization owners can remove anyone
	// 2. Organization admins can remove regular members and other admins
	// 3. A user can remove themselves
	isOwner := org.HasRole(removedBy, models.OrgRoleOwner)
	isAdmin := org.HasRole(removedBy, models.OrgRoleAdmin)
	isSelf := removedBy == memberID

	if !isOwner && !isSelf && (memberToRemove.Role == models.OrgRoleOwner || (!isAdmin && memberToRemove.Role == models.OrgRoleAdmin)) {
		return errors.New("insufficient permissions to remove this organization member")
	}

	// If trying to remove the last owner, prevent it
	if memberToRemove.Role == models.OrgRoleOwner {
		// Count owners
		ownerCount := 0
		for _, m := range org.Members {
			if m.Role == models.OrgRoleOwner {
				ownerCount++
			}
		}

		// If this is the only owner, don't allow removal
		if ownerCount <= 1 {
			return errors.New("cannot remove the only organization owner")
		}
	}

	// Remove member from organization
	err = s.orgRepo.RemoveMember(ctx, orgID, memberID)
	if err != nil {
		log.Error().Err(err).Str("orgId", orgID).Str("userId", memberID).
			Msg("Failed to remove member from organization")
		return err
	}

	// Remove organization from user
	err = s.userRepo.RemoveOrganizationFromUser(ctx, memberID, orgID)
	if err != nil {
		log.Error().Err(err).Str("orgId", orgID).Str("userId", memberID).
			Msg("Failed to remove organization from user")
		// Don't fail the operation, but log the error
	}

	// Publish event
	go func(o *models.Organization, userID string) {
		err := s.producer.PublishUserEvent(
			kafka.OrganizationMemberRemoved,
			map[string]interface{}{
				"orgId":     o.ID,
				"orgName":   o.Name,
				"userId":    userID,
				"removedBy": removedBy,
				"removedAt": time.Now(),
			},
			o.ID,
			"",
		)
		if err != nil {
			log.Error().Err(err).Str("orgId", o.ID).Str("userId", userID).
				Msg("Failed to publish organization.member.removed event")
		}
	}(org, memberID)

	return nil
}

// GetOrganizationTeams gets teams in an organization
func (s *OrganizationService) GetOrganizationTeams(ctx context.Context, orgID string, page, limit int, userID string) ([]*models.Team, int64, error) {
	// Verify organization exists
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, 0, errors.New("organization not found")
		}
		log.Error().Err(err).Str("id", orgID).Msg("Failed to get organization for teams")
		return nil, 0, err
	}

	// Verify user is member of the organization
	if !org.IsMember(userID) {
		return nil, 0, errors.New("user is not a member of the organization")
	}

	// Get teams
	return s.teamRepo.GetTeamsByOrganization(ctx, orgID, page, limit)
}
