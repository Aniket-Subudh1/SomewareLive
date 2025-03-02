package validators

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

// Custom validation tags
const (
	// UUID4 validation
	uuid4Tag = "uuid4"
	// Role validation
	roleTag = "role"
	// Email validation
	emailTag = "email"
	// URL validation
	urlTag = "url"
	// Phone validation
	phoneTag = "phone"
)

// Initialize custom validators
func InitUserValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// Register validators
		registerUUID4Validator(v)
		registerRoleValidator(v)
		registerEmailValidator(v)
		registerURLValidator(v)
		registerPhoneValidator(v)

		// Register JSON tag name
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		log.Info().Msg("Custom validators registered")
	} else {
		log.Error().Msg("Failed to register custom validators")
	}
}

// registerUUID4Validator registers the uuid4 validator
func registerUUID4Validator(v *validator.Validate) {
	_ = v.RegisterValidation(uuid4Tag, func(fl validator.FieldLevel) bool {
		uuid4Regex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
		return uuid4Regex.MatchString(fl.Field().String())
	})
}

// registerRoleValidator registers the role validator
func registerRoleValidator(v *validator.Validate) {
	_ = v.RegisterValidation(roleTag, func(fl validator.FieldLevel) bool {
		role := fl.Field().String()
		validRoles := []string{"user", "presenter", "admin"}
		for _, r := range validRoles {
			if role == r {
				return true
			}
		}
		return false
	})
}

// registerEmailValidator registers the email validator
func registerEmailValidator(v *validator.Validate) {
	_ = v.RegisterValidation(emailTag, func(fl validator.FieldLevel) bool {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
		return emailRegex.MatchString(fl.Field().String())
	})
}

// registerURLValidator registers the URL validator
func registerURLValidator(v *validator.Validate) {
	_ = v.RegisterValidation(urlTag, func(fl validator.FieldLevel) bool {
		urlRegex := regexp.MustCompile(`^(https?):\/\/[^\s/$.?#].[^\s]*$`)
		return urlRegex.MatchString(fl.Field().String())
	})
}

// registerPhoneValidator registers the phone validator
func registerPhoneValidator(v *validator.Validate) {
	_ = v.RegisterValidation(phoneTag, func(fl validator.FieldLevel) bool {
		phoneRegex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
		return phoneRegex.MatchString(fl.Field().String())
	})
}

// UserRequestValidator is a middleware for validating user requests
func UserRequestValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add custom validation logic here if needed
		c.Next()
	}
}
