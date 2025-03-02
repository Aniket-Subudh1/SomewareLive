package models

import (
	"time"

	"github.com/google/uuid"
)

// OrganizationMemberRole represents an organization member role
type OrganizationMemberRole string

// Organization member roles
const (
	OrgRoleOwner  OrganizationMemberRole = "owner"
	OrgRoleAdmin  OrganizationMemberRole = "admin"
	OrgRoleMember OrganizationMemberRole = "member"
)

// Organization represents an organization in the system
type Organization struct {
	ID          string               `bson:"_id,omitempty" json:"id"`
	Name        string               `bson:"name" json:"name"`
	Description string               `bson:"description,omitempty" json:"description,omitempty"`
	LogoURL     string               `bson:"logoUrl,omitempty" json:"logoUrl,omitempty"`
	Website     string               `bson:"website,omitempty" json:"website,omitempty"`
	Industry    string               `bson:"industry,omitempty" json:"industry,omitempty"`
	Size        string               `bson:"size,omitempty" json:"size,omitempty"`
	Location    string               `bson:"location,omitempty" json:"location,omitempty"`
	CreatedBy   string               `bson:"createdBy" json:"createdBy"`
	CreatedAt   time.Time            `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time            `bson:"updatedAt" json:"updatedAt"`
	Members     []OrganizationMember `bson:"members" json:"members"`
	TeamIDs     []string             `bson:"teamIds,omitempty" json:"teamIds,omitempty"`
	Settings    OrganizationSettings `bson:"settings" json:"settings"`
}

// OrganizationMember represents a member of an organization
type OrganizationMember struct {
	UserID    string                 `bson:"userId" json:"userId"`
	Role      OrganizationMemberRole `bson:"role" json:"role"`
	JoinedAt  time.Time              `bson:"joinedAt" json:"joinedAt"`
	InvitedBy string                 `bson:"invitedBy,omitempty" json:"invitedBy,omitempty"`
}

// OrganizationSettings represents settings for an organization
type OrganizationSettings struct {
	DefaultUserRole OrganizationMemberRole `bson:"defaultUserRole" json:"defaultUserRole"`
	Features        struct {
		AllowPublicEvents  bool `bson:"allowPublicEvents" json:"allowPublicEvents"`
		AllowExternalUsers bool `bson:"allowExternalUsers" json:"allowExternalUsers"`
		EnableTeams        bool `bson:"enableTeams" json:"enableTeams"`
	} `bson:"features" json:"features"`
	Branding struct {
		PrimaryColor   string `bson:"primaryColor,omitempty" json:"primaryColor,omitempty"`
		SecondaryColor string `bson:"secondaryColor,omitempty" json:"secondaryColor,omitempty"`
		LogoURL        string `bson:"logoUrl,omitempty" json:"logoUrl,omitempty"`
		FaviconURL     string `bson:"faviconUrl,omitempty" json:"faviconUrl,omitempty"`
	} `bson:"branding" json:"branding"`
}

// CreateOrganizationRequest represents a request to create a new organization
type CreateOrganizationRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"max=500"`
	LogoURL     string `json:"logoUrl" validate:"omitempty,url"`
	Website     string `json:"website" validate:"omitempty,url"`
	Industry    string `json:"industry" validate:"max=100"`
	Size        string `json:"size" validate:"omitempty,oneof=1-10 11-50 51-200 201-500 501-1000 1001+"`
	Location    string `json:"location" validate:"max=100"`
}

// UpdateOrganizationRequest represents a request to update an organization
type UpdateOrganizationRequest struct {
	Name        *string                     `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Description *string                     `json:"description,omitempty" validate:"omitempty,max=500"`
	LogoURL     *string                     `json:"logoUrl,omitempty" validate:"omitempty,url"`
	Website     *string                     `json:"website,omitempty" validate:"omitempty,url"`
	Industry    *string                     `json:"industry,omitempty" validate:"omitempty,max=100"`
	Size        *string                     `json:"size,omitempty" validate:"omitempty,oneof=1-10 11-50 51-200 201-500 501-1000 1001+"`
	Location    *string                     `json:"location,omitempty" validate:"omitempty,max=100"`
	Settings    *UpdateOrganizationSettings `json:"settings,omitempty"`
}

// UpdateOrganizationSettings represents a request to update organization settings
type UpdateOrganizationSettings struct {
	DefaultUserRole *OrganizationMemberRole `json:"defaultUserRole,omitempty" validate:"omitempty,oneof=owner admin member"`
	Features        *struct {
		AllowPublicEvents  *bool `json:"allowPublicEvents,omitempty"`
		AllowExternalUsers *bool `json:"allowExternalUsers,omitempty"`
		EnableTeams        *bool `json:"enableTeams,omitempty"`
	} `json:"features,omitempty"`
	Branding *struct {
		PrimaryColor   *string `json:"primaryColor,omitempty" validate:"omitempty,hexcolor"`
		SecondaryColor *string `json:"secondaryColor,omitempty" validate:"omitempty,hexcolor"`
		LogoURL        *string `json:"logoUrl,omitempty" validate:"omitempty,url"`
		FaviconURL     *string `json:"faviconUrl,omitempty" validate:"omitempty,url"`
	} `json:"branding,omitempty"`
}

// AddOrganizationMemberRequest represents a request to add a member to an organization
type AddOrganizationMemberRequest struct {
	UserID string                 `json:"userId" validate:"required"`
	Role   OrganizationMemberRole `json:"role" validate:"required,oneof=owner admin member"`
}

// UpdateOrganizationMemberRequest represents a request to update an organization member
type UpdateOrganizationMemberRequest struct {
	Role OrganizationMemberRole `json:"role" validate:"required,oneof=owner admin member"`
}

// OrganizationResponse represents an organization response
type OrganizationResponse struct {
	ID          string                     `json:"id"`
	Name        string                     `json:"name"`
	Description string                     `json:"description,omitempty"`
	LogoURL     string                     `json:"logoUrl,omitempty"`
	Website     string                     `json:"website,omitempty"`
	Industry    string                     `json:"industry,omitempty"`
	Size        string                     `json:"size,omitempty"`
	Location    string                     `json:"location,omitempty"`
	CreatedBy   string                     `json:"createdBy"`
	CreatedAt   time.Time                  `json:"createdAt"`
	MemberCount int                        `json:"memberCount"`
	TeamCount   int                        `json:"teamCount"`
	Members     []OrganizationMemberDetail `json:"members,omitempty"`
	Settings    OrganizationSettings       `json:"settings,omitempty"`
}

// OrganizationMemberDetail represents detailed information about an organization member
type OrganizationMemberDetail struct {
	UserID    string                 `json:"userId"`
	Email     string                 `json:"email"`
	FirstName string                 `json:"firstName"`
	LastName  string                 `json:"lastName"`
	FullName  string                 `json:"fullName"`
	Role      OrganizationMemberRole `json:"role"`
	JoinedAt  time.Time              `json:"joinedAt"`
}

// NewOrganization creates a new organization from a request
func NewOrganization(req CreateOrganizationRequest, createdBy string) *Organization {
	now := time.Now()
	return &Organization{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		LogoURL:     req.LogoURL,
		Website:     req.Website,
		Industry:    req.Industry,
		Size:        req.Size,
		Location:    req.Location,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
		Members: []OrganizationMember{
			{
				UserID:   createdBy,
				Role:     OrgRoleOwner,
				JoinedAt: now,
			},
		},
		Settings: OrganizationSettings{
			DefaultUserRole: OrgRoleMember,
			Features: struct {
				AllowPublicEvents  bool `bson:"allowPublicEvents" json:"allowPublicEvents"`
				AllowExternalUsers bool `bson:"allowExternalUsers" json:"allowExternalUsers"`
				EnableTeams        bool `bson:"enableTeams" json:"enableTeams"`
			}{
				AllowPublicEvents:  true,
				AllowExternalUsers: false,
				EnableTeams:        true,
			},
			Branding: struct {
				PrimaryColor   string `bson:"primaryColor,omitempty" json:"primaryColor,omitempty"`
				SecondaryColor string `bson:"secondaryColor,omitempty" json:"secondaryColor,omitempty"`
				LogoURL        string `bson:"logoUrl,omitempty" json:"logoUrl,omitempty"`
				FaviconURL     string `bson:"faviconUrl,omitempty" json:"faviconUrl,omitempty"`
			}{
				PrimaryColor:   "#3f51b5",
				SecondaryColor: "#f50057",
			},
		},
	}
}

// ToResponse converts an organization to a response
func (o *Organization) ToResponse(includeMembers bool, includeSettings bool) OrganizationResponse {
	response := OrganizationResponse{
		ID:          o.ID,
		Name:        o.Name,
		Description: o.Description,
		LogoURL:     o.LogoURL,
		Website:     o.Website,
		Industry:    o.Industry,
		Size:        o.Size,
		Location:    o.Location,
		CreatedBy:   o.CreatedBy,
		CreatedAt:   o.CreatedAt,
		MemberCount: len(o.Members),
		TeamCount:   len(o.TeamIDs),
	}

	if includeMembers {
		response.Members = make([]OrganizationMemberDetail, 0, len(o.Members))
		for _, member := range o.Members {
			// Note: In a real implementation, you would lookup user details
			// from the user repository. This is a simplified version.
			response.Members = append(response.Members, OrganizationMemberDetail{
				UserID:   member.UserID,
				Role:     member.Role,
				JoinedAt: member.JoinedAt,
			})
		}
	}

	if includeSettings {
		response.Settings = o.Settings
	}

	return response
}

// Apply applies an update request to an organization
func (o *Organization) Apply(req UpdateOrganizationRequest) {
	o.UpdatedAt = time.Now()

	if req.Name != nil {
		o.Name = *req.Name
	}
	if req.Description != nil {
		o.Description = *req.Description
	}
	if req.LogoURL != nil {
		o.LogoURL = *req.LogoURL
	}
	if req.Website != nil {
		o.Website = *req.Website
	}
	if req.Industry != nil {
		o.Industry = *req.Industry
	}
	if req.Size != nil {
		o.Size = *req.Size
	}
	if req.Location != nil {
		o.Location = *req.Location
	}

	// Update settings
	if req.Settings != nil {
		if req.Settings.DefaultUserRole != nil {
			o.Settings.DefaultUserRole = *req.Settings.DefaultUserRole
		}

		// Update features
		if req.Settings.Features != nil {
			if req.Settings.Features.AllowPublicEvents != nil {
				o.Settings.Features.AllowPublicEvents = *req.Settings.Features.AllowPublicEvents
			}
			if req.Settings.Features.AllowExternalUsers != nil {
				o.Settings.Features.AllowExternalUsers = *req.Settings.Features.AllowExternalUsers
			}
			if req.Settings.Features.EnableTeams != nil {
				o.Settings.Features.EnableTeams = *req.Settings.Features.EnableTeams
			}
		}

		// Update branding
		if req.Settings.Branding != nil {
			if req.Settings.Branding.PrimaryColor != nil {
				o.Settings.Branding.PrimaryColor = *req.Settings.Branding.PrimaryColor
			}
			if req.Settings.Branding.SecondaryColor != nil {
				o.Settings.Branding.SecondaryColor = *req.Settings.Branding.SecondaryColor
			}
			if req.Settings.Branding.LogoURL != nil {
				o.Settings.Branding.LogoURL = *req.Settings.Branding.LogoURL
			}
			if req.Settings.Branding.FaviconURL != nil {
				o.Settings.Branding.FaviconURL = *req.Settings.Branding.FaviconURL
			}
		}
	}
}

// AddMember adds a member to the organization
func (o *Organization) AddMember(userID string, role OrganizationMemberRole, invitedBy string) bool {
	// Check if the user is already a member
	for i, member := range o.Members {
		if member.UserID == userID {
			// Update the member's role if it's different
			if member.Role != role {
				o.Members[i].Role = role
				o.UpdatedAt = time.Now()
				return true
			}
			return false
		}
	}

	// Add the new member
	o.Members = append(o.Members, OrganizationMember{
		UserID:    userID,
		Role:      role,
		JoinedAt:  time.Now(),
		InvitedBy: invitedBy,
	})
	o.UpdatedAt = time.Now()
	return true
}

// UpdateMember updates an organization member's role
func (o *Organization) UpdateMember(userID string, role OrganizationMemberRole) bool {
	for i, member := range o.Members {
		if member.UserID == userID {
			// Update the member's role if it's different
			if member.Role != role {
				o.Members[i].Role = role
				o.UpdatedAt = time.Now()
				return true
			}
			return false
		}
	}
	return false
}

// RemoveMember removes a member from the organization
func (o *Organization) RemoveMember(userID string) bool {
	for i, member := range o.Members {
		if member.UserID == userID {
			// Remove the member
			o.Members = append(o.Members[:i], o.Members[i+1:]...)
			o.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// GetMember gets a member from the organization
func (o *Organization) GetMember(userID string) *OrganizationMember {
	for _, member := range o.Members {
		if member.UserID == userID {
			return &member
		}
	}
	return nil
}

// IsMember checks if a user is a member of the organization
func (o *Organization) IsMember(userID string) bool {
	return o.GetMember(userID) != nil
}

// HasRole checks if a user has a specific role in the organization
func (o *Organization) HasRole(userID string, roles ...OrganizationMemberRole) bool {
	member := o.GetMember(userID)
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

// AddTeam adds a team to the organization
func (o *Organization) AddTeam(teamID string) bool {
	// Check if the team is already in the organization
	for _, id := range o.TeamIDs {
		if id == teamID {
			return false
		}
	}

	// Add the team
	o.TeamIDs = append(o.TeamIDs, teamID)
	o.UpdatedAt = time.Now()
	return true
}

// RemoveTeam removes a team from the organization
func (o *Organization) RemoveTeam(teamID string) bool {
	for i, id := range o.TeamIDs {
		if id == teamID {
			// Remove the team
			o.TeamIDs = append(o.TeamIDs[:i], o.TeamIDs[i+1:]...)
			o.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}
