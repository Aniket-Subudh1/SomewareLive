package models

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents a user role
type UserRole string

// User roles
const (
	RoleUser      UserRole = "user"
	RolePresenter UserRole = "presenter"
	RoleAdmin     UserRole = "admin"
)

// UserStatus represents a user status
type UserStatus string

// User statuses
const (
	StatusActive   UserStatus = "active"
	StatusInactive UserStatus = "inactive"
	StatusPending  UserStatus = "pending"
)

// User represents a user in the system
type User struct {
	ID              string            `bson:"_id,omitempty" json:"id"`
	UserID          string            `bson:"userId" json:"userId"`
	Email           string            `bson:"email" json:"email"`
	FirstName       string            `bson:"firstName" json:"firstName"`
	LastName        string            `bson:"lastName" json:"lastName"`
	Role            UserRole          `bson:"role" json:"role"`
	Status          UserStatus        `bson:"status" json:"status"`
	ProfilePicture  string            `bson:"profilePicture,omitempty" json:"profilePicture,omitempty"`
	Bio             string            `bson:"bio,omitempty" json:"bio,omitempty"`
	JobTitle        string            `bson:"jobTitle,omitempty" json:"jobTitle,omitempty"`
	Company         string            `bson:"company,omitempty" json:"company,omitempty"`
	Location        string            `bson:"location,omitempty" json:"location,omitempty"`
	Phone           string            `bson:"phone,omitempty" json:"phone,omitempty"`
	Website         string            `bson:"website,omitempty" json:"website,omitempty"`
	SocialLinks     map[string]string `bson:"socialLinks,omitempty" json:"socialLinks,omitempty"`
	Preferences     UserPreferences   `bson:"preferences" json:"preferences"`
	LastLogin       *time.Time        `bson:"lastLogin,omitempty" json:"lastLogin,omitempty"`
	CreatedAt       time.Time         `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time         `bson:"updatedAt" json:"updatedAt"`
	OrganizationIDs []string          `bson:"organizationIds,omitempty" json:"organizationIds,omitempty"`
	TeamIDs         []string          `bson:"teamIds,omitempty" json:"teamIds,omitempty"`
}

// UserPreferences represents user preferences
type UserPreferences struct {
	Language             string `bson:"language" json:"language"`
	Theme                string `bson:"theme" json:"theme"`
	Timezone             string `bson:"timezone" json:"timezone"`
	NotificationSettings struct {
		Email bool `bson:"email" json:"email"`
		Push  bool `bson:"push" json:"push"`
		InApp bool `bson:"inApp" json:"inApp"`
	} `bson:"notificationSettings" json:"notificationSettings"`
	Privacy struct {
		ShowProfileToEveryone bool `bson:"showProfileToEveryone" json:"showProfileToEveryone"`
		ShowEmailToEveryone   bool `bson:"showEmailToEveryone" json:"showEmailToEveryone"`
	} `bson:"privacy" json:"privacy"`
}

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	UserID    string   `json:"userId" validate:"required"`
	Email     string   `json:"email" validate:"required,email"`
	FirstName string   `json:"firstName" validate:"required"`
	LastName  string   `json:"lastName" validate:"required"`
	Role      UserRole `json:"role" validate:"required,oneof=user presenter admin"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	FirstName      *string            `json:"firstName,omitempty"`
	LastName       *string            `json:"lastName,omitempty"`
	Status         *UserStatus        `json:"status,omitempty" validate:"omitempty,oneof=active inactive pending"`
	ProfilePicture *string            `json:"profilePicture,omitempty"`
	Bio            *string            `json:"bio,omitempty"`
	JobTitle       *string            `json:"jobTitle,omitempty"`
	Company        *string            `json:"company,omitempty"`
	Location       *string            `json:"location,omitempty"`
	Phone          *string            `json:"phone,omitempty" validate:"omitempty,e164"`
	Website        *string            `json:"website,omitempty" validate:"omitempty,url"`
	SocialLinks    map[string]string  `json:"socialLinks,omitempty"`
	Preferences    *UpdatePreferences `json:"preferences,omitempty"`
}

// UpdatePreferences represents a request to update user preferences
type UpdatePreferences struct {
	Language             *string `json:"language,omitempty"`
	Theme                *string `json:"theme,omitempty"`
	Timezone             *string `json:"timezone,omitempty"`
	NotificationSettings *struct {
		Email *bool `json:"email,omitempty"`
		Push  *bool `json:"push,omitempty"`
		InApp *bool `json:"inApp,omitempty"`
	} `json:"notificationSettings,omitempty"`
	Privacy *struct {
		ShowProfileToEveryone *bool `json:"showProfileToEveryone,omitempty"`
		ShowEmailToEveryone   *bool `json:"showEmailToEveryone,omitempty"`
	} `json:"privacy,omitempty"`
}

// UserResponse represents a user response
type UserResponse struct {
	ID             string            `json:"id"`
	Email          string            `json:"email"`
	FirstName      string            `json:"firstName"`
	LastName       string            `json:"lastName"`
	FullName       string            `json:"fullName"`
	Role           UserRole          `json:"role"`
	Status         UserStatus        `json:"status"`
	ProfilePicture string            `json:"profilePicture,omitempty"`
	Bio            string            `json:"bio,omitempty"`
	JobTitle       string            `json:"jobTitle,omitempty"`
	Company        string            `json:"company,omitempty"`
	Location       string            `json:"location,omitempty"`
	SocialLinks    map[string]string `json:"socialLinks,omitempty"`
	LastLogin      *time.Time        `json:"lastLogin,omitempty"`
	CreatedAt      time.Time         `json:"createdAt"`
}

// NewUser creates a new user from a request
func NewUser(req CreateUserRequest) *User {
	now := time.Now()
	return &User{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      req.Role,
		Status:    StatusActive,
		Preferences: UserPreferences{
			Language: "en",
			Theme:    "light",
			Timezone: "UTC",
			NotificationSettings: struct {
				Email bool `bson:"email" json:"email"`
				Push  bool `bson:"push" json:"push"`
				InApp bool `bson:"inApp" json:"inApp"`
			}{
				Email: true,
				Push:  true,
				InApp: true,
			},
			Privacy: struct {
				ShowProfileToEveryone bool `bson:"showProfileToEveryone" json:"showProfileToEveryone"`
				ShowEmailToEveryone   bool `bson:"showEmailToEveryone" json:"showEmailToEveryone"`
			}{
				ShowProfileToEveryone: true,
				ShowEmailToEveryone:   false,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// ToResponse converts a user to a response
func (u *User) ToResponse() UserResponse {
	return UserResponse{
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
	}
}

// Apply applies an update request to a user
func (u *User) Apply(req UpdateUserRequest) {
	u.UpdatedAt = time.Now()

	if req.FirstName != nil {
		u.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		u.LastName = *req.LastName
	}
	if req.Status != nil {
		u.Status = *req.Status
	}
	if req.ProfilePicture != nil {
		u.ProfilePicture = *req.ProfilePicture
	}
	if req.Bio != nil {
		u.Bio = *req.Bio
	}
	if req.JobTitle != nil {
		u.JobTitle = *req.JobTitle
	}
	if req.Company != nil {
		u.Company = *req.Company
	}
	if req.Location != nil {
		u.Location = *req.Location
	}
	if req.Phone != nil {
		u.Phone = *req.Phone
	}
	if req.Website != nil {
		u.Website = *req.Website
	}
	if req.SocialLinks != nil {
		u.SocialLinks = req.SocialLinks
	}

	// Update preferences
	if req.Preferences != nil {
		if req.Preferences.Language != nil {
			u.Preferences.Language = *req.Preferences.Language
		}
		if req.Preferences.Theme != nil {
			u.Preferences.Theme = *req.Preferences.Theme
		}
		if req.Preferences.Timezone != nil {
			u.Preferences.Timezone = *req.Preferences.Timezone
		}

		// Update notification settings
		if req.Preferences.NotificationSettings != nil {
			if req.Preferences.NotificationSettings.Email != nil {
				u.Preferences.NotificationSettings.Email = *req.Preferences.NotificationSettings.Email
			}
			if req.Preferences.NotificationSettings.Push != nil {
				u.Preferences.NotificationSettings.Push = *req.Preferences.NotificationSettings.Push
			}
			if req.Preferences.NotificationSettings.InApp != nil {
				u.Preferences.NotificationSettings.InApp = *req.Preferences.NotificationSettings.InApp
			}
		}

		// Update privacy settings
		if req.Preferences.Privacy != nil {
			if req.Preferences.Privacy.ShowProfileToEveryone != nil {
				u.Preferences.Privacy.ShowProfileToEveryone = *req.Preferences.Privacy.ShowProfileToEveryone
			}
			if req.Preferences.Privacy.ShowEmailToEveryone != nil {
				u.Preferences.Privacy.ShowEmailToEveryone = *req.Preferences.Privacy.ShowEmailToEveryone
			}
		}
	}
}
