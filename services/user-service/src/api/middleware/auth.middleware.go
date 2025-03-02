package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/your-username/slido-clone/user-service/config"
	"github.com/your-username/slido-clone/user-service/pkg/utils"
)

// AuthMiddleware creates a Gin middleware for authentication
func AuthMiddleware(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")

		// Extract token
		token, err := utils.ExtractToken(authHeader)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to extract token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: " + err.Error(),
			})
			return
		}

		// Validate token
		claims, err := utils.ValidateToken(token, cfg)
		if err != nil {
			log.Debug().Err(err).Msg("Invalid token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: " + err.Error(),
			})
			return
		}

		// Store user information in context
		c.Set("userId", claims.Subject)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", claims.Role)

		if len(claims.Roles) > 0 {
			c.Set("userRoles", claims.Roles)
		} else if claims.Role != "" {
			// For backward compatibility
			c.Set("userRoles", []string{claims.Role})
		}

		// Continue
		c.Next()
	}
}

// RoleMiddleware creates a Gin middleware for role-based authorization
func RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user roles from context
		userRolesI, exists := c.Get("userRoles")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: user roles not found",
			})
			return
		}

		userRoles, ok := userRolesI.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error: invalid user roles format",
			})
			return
		}

		// Check if user has any of the required roles
		hasRole := false
		for _, role := range roles {
			for _, userRole := range userRoles {
				if strings.EqualFold(userRole, role) {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Forbidden: insufficient permissions",
			})
			return
		}

		// Continue
		c.Next()
	}
}

// OptionalAuthMiddleware creates a Gin middleware for optional authentication
func OptionalAuthMiddleware(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token, but that's ok
			c.Next()
			return
		}

		// Extract token
		token, err := utils.ExtractToken(authHeader)
		if err != nil {
			// Invalid format, but optional so continue
			log.Debug().Err(err).Msg("Failed to extract optional token")
			c.Next()
			return
		}

		// Validate token
		claims, err := utils.ValidateToken(token, cfg)
		if err != nil {
			// Invalid token, but optional so continue
			log.Debug().Err(err).Msg("Invalid optional token")
			c.Next()
			return
		}

		// Store user information in context
		c.Set("userId", claims.Subject)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", claims.Role)

		if len(claims.Roles) > 0 {
			c.Set("userRoles", claims.Roles)
		} else if claims.Role != "" {
			// For backward compatibility
			c.Set("userRoles", []string{claims.Role})
		}

		// Set authenticated flag
		c.Set("authenticated", true)

		// Continue
		c.Next()
	}
}

// GetUserId gets the user ID from the context
func GetUserId(c *gin.Context) string {
	userIdI, exists := c.Get("userId")
	if !exists {
		return ""
	}
	userId, ok := userIdI.(string)
	if !ok {
		return ""
	}
	return userId
}

// IsAuthenticated checks if the user is authenticated
func IsAuthenticated(c *gin.Context) bool {
	authenticatedI, exists := c.Get("authenticated")
	if !exists {
		return false
	}
	authenticated, ok := authenticatedI.(bool)
	if !ok {
		return false
	}
	return authenticated
}
