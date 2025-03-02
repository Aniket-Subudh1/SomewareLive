package models

import (
	"time"

	"github.com/google/uuid"
)

// TeamMemberRole represents a team member role
type TeamMemberRole string

// Team member roles
const (
	TeamRoleOwner  TeamMemberRole = "owner"
	TeamRoleAdmin  TeamMemberRole = "admin"
	TeamRoleMember TeamMemberRole = "member"
	TeamRoleViewer TeamMemberRole = "viewer"
)

// Team represents a team in the system
type Team struct {
	ID             string       `bson:"_id,omitempty" json:"id"`
	Name           string       `bson:"name" json:"name"`
	Description    string       `bson:"description,omitempty" json:"description,omitempty"`
	LogoURL        string       `bson:"logoUrl,omitempty" json:"logoUrl,omitempty"`
	OrganizationID string       `bson:"organizationId" json:"organizationId"`
	CreatedBy      string       `bson:"createdBy" json:"createdBy"`
	CreatedAt      time.Time    `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time    `bson:"updatedAt" json:"updatedAt"`
	Members        []TeamMember `bson:"members" json:"members"`
}

// TeamMember represents a member of a team
type TeamMember struct {
	UserID    string         `bson:"userId" json:"userId"`
	Role      TeamMemberRole `bson:"role" json:"role"`
	JoinedAt  time.Time      `bson:"joinedAt" json:"joinedAt"`
	InvitedBy string         `bson:"invitedBy,omitempty" json:"invitedBy,omitempty"`
}

// CreateTeamRequest represents a request to create a new team
type CreateTeamRequest struct {
	Name           string `json:"name" validate:"required,min=3,max=50"`
	Description    string `json:"description" validate:"max=500"`
	LogoURL        string `json:"logoUrl" validate:"omitempty,url"`
	OrganizationID string `json:"organizationId" validate:"required"`
}

// UpdateTeamRequest represents a request to update a team
type UpdateTeamRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=3,max=50"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	LogoURL     *string `json:"logoUrl,omitempty" validate:"omitempty,url"`
}

// AddTeamMemberRequest represents a request to add a member to a team
type AddTeamMemberRequest struct {
	UserID string         `json:"userId" validate:"required"`
	Role   TeamMemberRole `json:"role" validate:"required,oneof=owner admin member viewer"`
}

// UpdateTeamMemberRequest represents a request to update a team member
type UpdateTeamMemberRequest struct {
	Role TeamMemberRole `json:"role" validate:"required,oneof=owner admin member viewer"`
}

// TeamResponse represents a team response
type TeamResponse struct {
	ID             string             `json:"id"`
	Name           string             `json:"name"`
	Description    string             `json:"description,omitempty"`
	LogoURL        string             `json:"logoUrl,omitempty"`
	OrganizationID string             `json:"organizationId"`
	CreatedBy      string             `json:"createdBy"`
	CreatedAt      time.Time          `json:"createdAt"`
	MemberCount    int                `json:"memberCount"`
	Members        []TeamMemberDetail `json:"members,omitempty"`
}

// TeamMemberDetail represents detailed information about a team member
type TeamMemberDetail struct {
	UserID    string         `json:"userId"`
	Email     string         `json:"email"`
	FirstName string         `json:"firstName"`
	LastName  string         `json:"lastName"`
	FullName  string         `json:"fullName"`
	Role      TeamMemberRole `json:"role"`
	JoinedAt  time.Time      `json:"joinedAt"`
}

// NewTeam creates a new team from a request
func NewTeam(req CreateTeamRequest, createdBy string) *Team {
	now := time.Now()
	return &Team{
		ID:             uuid.New().String(),
		Name:           req.Name,
		Description:    req.Description,
		LogoURL:        req.LogoURL,
		OrganizationID: req.OrganizationID,
		CreatedBy:      createdBy,
		CreatedAt:      now,
		UpdatedAt:      now,
		Members: []TeamMember{
			{
				UserID:   createdBy,
				Role:     TeamRoleOwner,
				JoinedAt: now,
			},
		},
	}
}

// ToResponse converts a team to a response
func (t *Team) ToResponse(includeMembers bool) TeamResponse {
	response := TeamResponse{
		ID:             t.ID,
		Name:           t.Name,
		Description:    t.Description,
		LogoURL:        t.LogoURL,
		OrganizationID: t.OrganizationID,
		CreatedBy:      t.CreatedBy,
		CreatedAt:      t.CreatedAt,
		MemberCount:    len(t.Members),
	}

	if includeMembers {
		response.Members = make([]TeamMemberDetail, 0, len(t.Members))
		for _, member := range t.Members {
			// Note: In a real implementation, you would lookup user details
			// from the user repository. This is a simplified version.
			response.Members = append(response.Members, TeamMemberDetail{
				UserID:   member.UserID,
				Role:     member.Role,
				JoinedAt: member.JoinedAt,
			})
		}
	}

	return response
}

// Apply applies an update request to a team
func (t *Team) Apply(req UpdateTeamRequest) {
	t.UpdatedAt = time.Now()

	if req.Name != nil {
		t.Name = *req.Name
	}
	if req.Description != nil {
		t.Description = *req.Description
	}
	if req.LogoURL != nil {
		t.LogoURL = *req.LogoURL
	}
}

// AddMember adds a member to the team
func (t *Team) AddMember(userID string, role TeamMemberRole, invitedBy string) bool {
	// Check if the user is already a member
	for i, member := range t.Members {
		if member.UserID == userID {
			// Update the member's role if it's different
			if member.Role != role {
				t.Members[i].Role = role
				t.UpdatedAt = time.Now()
				return true
			}
			return false
		}
	}

	// Add the new member
	t.Members = append(t.Members, TeamMember{
		UserID:    userID,
		Role:      role,
		JoinedAt:  time.Now(),
		InvitedBy: invitedBy,
	})
	t.UpdatedAt = time.Now()
	return true
}

// UpdateMember updates a team member's role
func (t *Team) UpdateMember(userID string, role TeamMemberRole) bool {
	for i, member := range t.Members {
		if member.UserID == userID {
			// Update the member's role if it's different
			if member.Role != role {
				t.Members[i].Role = role
				t.UpdatedAt = time.Now()
				return true
			}
			return false
		}
	}
	return false
}

// RemoveMember removes a member from the team
func (t *Team) RemoveMember(userID string) bool {
	for i, member := range t.Members {
		if member.UserID == userID {
			// Remove the member
			t.Members = append(t.Members[:i], t.Members[i+1:]...)
			t.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// GetMember gets a member from the team
func (t *Team) GetMember(userID string) *TeamMember {
	for _, member := range t.Members {
		if member.UserID == userID {
			return &member
		}
	}
	return nil
}

// IsMember checks if a user is a member of the team
func (t *Team) IsMember(userID string) bool {
	return t.GetMember(userID) != nil
}

// HasRole checks if a user has a specific role in the team
func (t *Team) HasRole(userID string, roles ...TeamMemberRole) bool {
	member := t.GetMember(userID)
	if member == nil {
		return false
	}

	for _, role := range roles {
		if member.Role == role {
			return true
		}
	}
	return false
}
