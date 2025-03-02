package validators

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

// Custom validation tags
const (
	// Team role validation
	teamRoleTag = "teamrole"
	// Organization role validation
	orgRoleTag = "orgrole"
)

// Initialize custom validators for teams
func InitTeamValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// Register validators
		registerTeamRoleValidator(v)
		registerOrgRoleValidator(v)

		// Register JSON tag name if not already registered
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		log.Info().Msg("Custom team validators registered")
	} else {
		log.Error().Msg("Failed to register custom team validators")
	}
}

// registerTeamRoleValidator registers the team role validator
func registerTeamRoleValidator(v *validator.Validate) {
	_ = v.RegisterValidation(teamRoleTag, func(fl validator.FieldLevel) bool {
		role := fl.Field().String()
		validRoles := []string{"owner", "admin", "member", "viewer"}
		for _, r := range validRoles {
			if role == r {
				return true
			}
		}
		return false
	})
}

// registerOrgRoleValidator registers the organization role validator
func registerOrgRoleValidator(v *validator.Validate) {
	_ = v.RegisterValidation(orgRoleTag, func(fl validator.FieldLevel) bool {
		role := fl.Field().String()
		validRoles := []string{"owner", "admin", "member"}
		for _, r := range validRoles {
			if role == r {
				return true
			}
		}
		return false
	})
}

// TeamValidation struct for team validation
type TeamValidation struct {
	// Add custom validation fields here
}

// New creates a new TeamValidation instance
func NewTeamValidation() *TeamValidation {
	return &TeamValidation{}
}

// ValidateTeamName validates a team name
func (v *TeamValidation) ValidateTeamName(name string) bool {
	// Team name should be between 3 and 50 characters
	if len(name) < 3 || len(name) > 50 {
		return false
	}
	return true
}

// ValidateTeamDescription validates a team description
func (v *TeamValidation) ValidateTeamDescription(description string) bool {
	// Team description should be at most 500 characters
	if len(description) > 500 {
		return false
	}
	return true
}

// ValidateTeamRole validates a team role
func (v *TeamValidation) ValidateTeamRole(role string) bool {
	validRoles := []string{"owner", "admin", "member", "viewer"}
	for _, r := range validRoles {
		if role == r {
			return true
		}
	}
	return false
}

// ValidateOrgRole validates an organization role
func (v *TeamValidation) ValidateOrgRole(role string) bool {
	validRoles := []string{"owner", "admin", "member"}
	for _, r := range validRoles {
		if role == r {
			return true
		}
	}
	return false
}

// TeamRequestValidator is a middleware for validating team requests
func TeamRequestValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add custom validation logic here if needed
		c.Next()
	}
}

// OrganizationRequestValidator is a middleware for validating organization requests
func OrganizationRequestValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add custom validation logic here if needed
		c.Next()
	}
}
