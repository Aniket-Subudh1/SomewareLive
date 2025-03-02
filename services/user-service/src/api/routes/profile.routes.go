package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/your-username/slido-clone/user-service/api/controllers"
	"github.com/your-username/slido-clone/user-service/api/middleware"
	"github.com/your-username/slido-clone/user-service/config"
)

// RegisterProfileRoutes registers profile routes
func RegisterProfileRoutes(router *gin.RouterGroup, profileController *controllers.ProfileController, cfg *config.JWTConfig) {
	// All profile routes require authentication
	protected := router.Group("")
	protected.Use(middleware.AuthMiddleware(cfg))

	// Profile routes
	protected.GET("/profile", profileController.GetProfile)
	protected.PUT("/profile", profileController.UpdateProfile)
	protected.GET("/profile/teams", profileController.GetUserTeams)
	protected.GET("/profile/organizations", profileController.GetUserOrganizations)
	protected.GET("/profile/full", profileController.GetFullProfile)
	protected.PUT("/profile/preferences", profileController.UpdateUserPreferences)
}
