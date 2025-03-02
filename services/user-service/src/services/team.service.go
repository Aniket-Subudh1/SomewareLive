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

// TeamService is a service for teams
type TeamService struct {
	teamRepo *repositories.TeamRepository
	userRepo *repositories.UserRepository
	orgRepo  *repositories.OrganizationRepository
	producer *kafka.Producer
}

// NewTeamService creates a new team service
func NewTeamService(
	teamRepo *repositories.TeamRepository,
	userRepo *repositories.UserRepository,
	orgRepo *repositories.OrganizationRepository,
	producer *kafka.Producer,
) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
		orgRepo:  orgRepo,
		producer: producer,
	}
}

// CreateTeam creates a new team
func (s *TeamService) CreateTeam(ctx context.Context, req models.CreateTeamRequest, createdBy string) (*models.Team, error) {
	// Verify organization exists
	org, err := s.orgRepo.GetByID(ctx, req.OrganizationID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("organization not found")
		}
		log.Error().Err(err).Str("orgId", req.OrganizationID).Msg("Failed to get organization for team creation")
		return nil, err
	}

	// Verify user is member of the organization
	if !org.IsMember(createdBy) {
		return nil, errors.New("user is not a member of the organization")
	}

	// Create team
	team := models.NewTeam(req, createdBy)

	// Save to database
	err = s.teamRepo.Create(ctx, team)
	if err != nil {
		log.Error().Err(err).Interface("req", req).Msg("Failed to create team")
		return nil, err
	}

	// Add team to organization
	err = s.orgRepo.AddTeam(ctx, req.OrganizationID, team.ID)
	if err != nil {
		log.Error().Err(err).Str("teamId", team.ID).Str("orgId", req.OrganizationID).
			Msg("Failed to add team to organization")
		// Don't fail the team creation, but log the error
	}

	// Add team to creator's user profile
	err = s.userRepo.AddTeamToUser(ctx, createdBy, team.ID)
	if err != nil {
		log.Error().Err(err).Str("teamId", team.ID).Str("userId", createdBy).
			Msg("Failed to add team to user")
		// Don't fail the team creation, but log the error
	}

	// Publish event
	go func(t *models.Team) {
		err := s.producer.PublishTeamEvent(
			kafka.TeamCreated,
			models.TeamResponse{
				ID:             t.ID,
				Name:           t.Name,
				Description:    t.Description,
				LogoURL:        t.LogoURL,
				OrganizationID: t.OrganizationID,
				CreatedBy:      t.CreatedBy,
				CreatedAt:      t.CreatedAt,
				MemberCount:    len(t.Members),
			},
			t.ID,
			"",
		)
		if err != nil {
			log.Error().Err(err).Str("teamId", t.ID).Msg("Failed to publish team.created event")
		}
	}(team)

	return team, nil
}

// GetTeamByID gets a team by ID
func (s *TeamService) GetTeamByID(ctx context.Context, id string) (*models.Team, error) {
	team, err := s.teamRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("team not found")
		}
		log.Error().Err(err).Str("id", id).Msg("Failed to get team by ID")
		return nil, err
	}
	return team, nil
}

// GetTeamsByOrganization gets teams by organization ID
func (s *TeamService) GetTeamsByOrganization(ctx context.Context, organizationID string, page, limit int) ([]*models.Team, int64, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get teams
	teams, total, err := s.teamRepo.GetTeamsByOrganization(ctx, organizationID, page, limit)
	if err != nil {
		log.Error().Err(err).Str("orgId", organizationID).Int("page", page).Int("limit", limit).
			Msg("Failed to get teams by organization")
		return nil, 0, err
	}

	return teams, total, nil
}

// GetTeamsByUser gets teams by user ID
func (s *TeamService) GetTeamsByUser(ctx context.Context, userID string, page, limit int) ([]*models.Team, int64, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get teams
	teams, total, err := s.teamRepo.GetTeamsByUser(ctx, userID, page, limit)
	if err != nil {
		log.Error().Err(err).Str("userId", userID).Int("page", page).Int("limit", limit).
			Msg("Failed to get teams by user")
		return nil, 0, err
	}

	return teams, total, nil
}

// UpdateTeam updates a team
func (s *TeamService) UpdateTeam(ctx context.Context, id string, req models.UpdateTeamRequest, userID string) (*models.Team, error) {
	// Get team
	team, err := s.teamRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("team not found")
		}
		log.Error().Err(err).Str("id", id).Msg("Failed to get team for update")
		return nil, err
	}

	// Check permissions - must be admin or owner
	if !team.HasRole(userID, models.TeamRoleOwner, models.TeamRoleAdmin) {
		return nil, errors.New("insufficient permissions to update team")
	}

	// Apply changes
	team.Apply(req)

	// Save to database
	err = s.teamRepo.Update(ctx, team)
	if err != nil {
		log.Error().Err(err).Str("id", id).Interface("req", req).
			Msg("Failed to update team")
		return nil, err
	}

	// Publish event
	go func(t *models.Team) {
		err := s.producer.PublishTeamEvent(
			kafka.TeamUpdated,
			team.ToResponse(false),
			t.ID,
			"",
		)
		if err != nil {
			log.Error().Err(err).Str("teamId", t.ID).Msg("Failed to publish team.updated event")
		}
	}(team)

	return team, nil
}

// DeleteTeam deletes a team
func (s *TeamService) DeleteTeam(ctx context.Context, id string, userID string) error {
	// Get team
	team, err := s.teamRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("team not found")
		}
		log.Error().Err(err).Str("id", id).Msg("Failed to get team for deletion")
		return err
	}

	// Check permissions - must be owner
	if !team.HasRole(userID, models.TeamRoleOwner) {
		return errors.New("insufficient permissions to delete team")
	}

	// Delete team
	err = s.teamRepo.Delete(ctx, id)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("Failed to delete team")
		return err
	}

	// Remove team from organization
	err = s.orgRepo.RemoveTeam(ctx, team.OrganizationID, id)
	if err != nil {
		log.Error().Err(err).Str("teamId", id).Str("orgId", team.OrganizationID).
			Msg("Failed to remove team from organization")
		// Don't fail the team deletion, but log the error
	}

	// Remove team from all members
	for _, member := range team.Members {
		err = s.userRepo.RemoveTeamFromUser(ctx, member.UserID, id)
		if err != nil {
			log.Error().Err(err).Str("teamId", id).Str("userId", member.UserID).
				Msg("Failed to remove team from user")
			// Don't fail the team deletion, but log the error
		}
	}

	// Publish event
	go func(t *models.Team) {
		err := s.producer.PublishTeamEvent(
			kafka.TeamDeleted,
			models.TeamResponse{
				ID:             t.ID,
				Name:           t.Name,
				OrganizationID: t.OrganizationID,
			},
			t.ID,
			"",
		)
		if err != nil {
			log.Error().Err(err).Str("teamId", t.ID).Msg("Failed to publish team.deleted event")
		}
	}(team)

	return nil
}

// AddTeamMember adds a member to a team
func (s *TeamService) AddTeamMember(ctx context.Context, teamID string, req models.AddTeamMemberRequest, invitedBy string) error {
	// Get team
	team, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("team not found")
		}
		log.Error().Err(err).Str("id", teamID).Msg("Failed to get team for adding member")
		return err
	}

	// Check permissions - must be admin or owner
	if !team.HasRole(invitedBy, models.TeamRoleOwner, models.TeamRoleAdmin) {
		return errors.New("insufficient permissions to add team member")
	}

	// Verify user exists
	user, err := s.userRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("user not found")
		}
		log.Error().Err(err).Str("userId", req.UserID).Msg("Failed to get user for adding to team")
		return err
	}

	// Verify user is member of the organization
	org, err := s.orgRepo.GetByID(ctx, team.OrganizationID)
	if err != nil {
		log.Error().Err(err).Str("orgId", team.OrganizationID).Msg("Failed to get organization for team member")
		return err
	}

	if !org.IsMember(req.UserID) {
		return errors.New("user is not a member of the organization")
	}

	// Add member to team
	err = s.teamRepo.AddMember(ctx, teamID, req.UserID, req.Role, invitedBy)
	if err != nil {
		log.Error().Err(err).Str("teamId", teamID).Str("userId", req.UserID).
			Msg("Failed to add member to team")
		return err
	}

	// Add team to user
	err = s.userRepo.AddTeamToUser(ctx, req.UserID, teamID)
	if err != nil {
		log.Error().Err(err).Str("teamId", teamID).Str("userId", req.UserID).
			Msg("Failed to add team to user")
		// Don't fail the operation, but log the error
	}

	// Refresh team data
	team, err = s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		log.Error().Err(err).Str("id", teamID).Msg("Failed to refresh team data after adding member")
		// Don't fail the operation, but log the error
	}

	// Publish event
	go func(t *models.Team, userID string, role models.TeamMemberRole) {
		if t == nil {
			return
		}

		// Find the added member
		var addedMember *models.TeamMember
		for _, m := range t.Members {
			if m.UserID == userID {
				addedMember = &m
				break
			}
		}

		if addedMember == nil {
			log.Error().Str("teamId", teamID).Str("userId", userID).
				Msg("Failed to find added member for event")
			return
		}

		err := s.producer.PublishTeamEvent(
			kafka.TeamMemberAdded,
			map[string]interface{}{
				"teamId":    t.ID,
				"teamName":  t.Name,
				"userId":    userID,
				"role":      role,
				"invitedBy": invitedBy,
				"joinedAt":  addedMember.JoinedAt,
			},
			t.ID,
			"",
		)
		if err != nil {
			log.Error().Err(err).Str("teamId", t.ID).Str("userId", userID).
				Msg("Failed to publish team.member.added event")
		}
	}(team, req.UserID, req.Role)

	return nil
}

// UpdateTeamMember updates a team member's role
func (s *TeamService) UpdateTeamMember(ctx context.Context, teamID, memberID string, req models.UpdateTeamMemberRequest, updatedBy string) error {
	// Get team
	team, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("team not found")
		}
		log.Error().Err(err).Str("id", teamID).Msg("Failed to get team for updating member")
		return err
	}

	// Check permissions - must be admin or owner
	if !team.HasRole(updatedBy, models.TeamRoleOwner, models.TeamRoleAdmin) {
		return errors.New("insufficient permissions to update team member")
	}

	// If updating an owner, only an owner can do that
	currentMember := team.GetMember(memberID)
	if currentMember != nil && currentMember.Role == models.TeamRoleOwner && !team.HasRole(updatedBy, models.TeamRoleOwner) {
		return errors.New("only a team owner can change the role of another owner")
	}

	// Check if the user is trying to update their own role to a lower one
	if memberID == updatedBy && (currentMember.Role == models.TeamRoleOwner && req.Role != models.TeamRoleOwner) {
		// Count owners
		ownerCount := 0
		for _, m := range team.Members {
			if m.Role == models.TeamRoleOwner {
				ownerCount++
			}
		}

		// If this is the only owner, don't allow role change
		if ownerCount <= 1 {
			return errors.New("cannot change role: team must have at least one owner")
		}
	}

	// Update member role
	err = s.teamRepo.AddMember(ctx, teamID, memberID, req.Role, updatedBy)
	if err != nil {
		log.Error().Err(err).Str("teamId", teamID).Str("userId", memberID).
			Msg("Failed to update team member")
		return err
	}

	// Refresh team data
	team, err = s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		log.Error().Err(err).Str("id", teamID).Msg("Failed to refresh team data after updating member")
		// Don't fail the operation, but log the error
	}

	// Publish event
	go func(t *models.Team, userID string, role models.TeamMemberRole) {
		if t == nil {
			return
		}

		err := s.producer.PublishTeamEvent(
			kafka.TeamMemberUpdated,
			map[string]interface{}{
				"teamId":    t.ID,
				"teamName":  t.Name,
				"userId":    userID,
				"role":      role,
				"updatedBy": updatedBy,
				"updatedAt": time.Now(),
			},
			t.ID,
			"",
		)
		if err != nil {
			log.Error().Err(err).Str("teamId", t.ID).Str("userId", userID).
				Msg("Failed to publish team.member.updated event")
		}
	}(team, memberID, req.Role)

	return nil
}

// RemoveTeamMember removes a member from a team
func (s *TeamService) RemoveTeamMember(ctx context.Context, teamID, memberID string, removedBy string) error {
	// Get team
	team, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("team not found")
		}
		log.Error().Err(err).Str("id", teamID).Msg("Failed to get team for removing member")
		return err
	}

	// Get member being removed
	memberToRemove := team.GetMember(memberID)
	if memberToRemove == nil {
		return errors.New("member not found in team")
	}

	// Check permissions
	// 1. Team owners can remove anyone
	// 2. Team admins can remove regular members and other admins
	// 3. A user can remove themselves
	isOwner := team.HasRole(removedBy, models.TeamRoleOwner)
	isAdmin := team.HasRole(removedBy, models.TeamRoleAdmin)
	isSelf := removedBy == memberID

	if !isOwner && !isSelf && (memberToRemove.Role == models.TeamRoleOwner || (!isAdmin && memberToRemove.Role == models.TeamRoleAdmin)) {
		return errors.New("insufficient permissions to remove this team member")
	}

	// If trying to remove the last owner, prevent it
	if memberToRemove.Role == models.TeamRoleOwner {
		// Count owners
		ownerCount := 0
		for _, m := range team.Members {
			if m.Role == models.TeamRoleOwner {
				ownerCount++
			}
		}

		// If this is the only owner, don't allow removal
		if ownerCount <= 1 {
			return errors.New("cannot remove the only team owner")
		}
	}

	// Remove member from team
	err = s.teamRepo.RemoveMember(ctx, teamID, memberID)
	if err != nil {
		log.Error().Err(err).Str("teamId", teamID).Str("userId", memberID).
			Msg("Failed to remove member from team")
		return err
	}

	// Remove team from user
	err = s.userRepo.RemoveTeamFromUser(ctx, memberID, teamID)
	if err != nil {
		log.Error().Err(err).Str("teamId", teamID).Str("userId", memberID).
			Msg("Failed to remove team from user")
		// Don't fail the operation, but log the error
	}

	// Publish event
	go func(t *models.Team, userID string) {
		err := s.producer.PublishTeamEvent(
			kafka.TeamMemberRemoved,
			map[string]interface{}{
				"teamId":    t.ID,
				"teamName":  t.Name,
				"userId":    userID,
				"removedBy": removedBy,
				"removedAt": time.Now(),
			},
			t.ID,
			"",
		)
		if err != nil {
			log.Error().Err(err).Str("teamId", t.ID).Str("userId", userID).
				Msg("Failed to publish team.member.removed event")
		}
	}(team, memberID)

	return nil
}
