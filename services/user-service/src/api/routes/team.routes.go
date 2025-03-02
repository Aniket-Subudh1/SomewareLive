package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/your-username/slido-clone/user-service/api/controllers"
	"github.com/your-username/slido-clone/user-service/api/middleware"
	"github.com/your-username/slido-clone/user-service/config"
)

// RegisterTeamRoutes registers team routes
func RegisterTeamRoutes(router *gin.RouterGroup, teamController *controllers.TeamController, cfg *config.JWTConfig) {
	// All team routes require authentication
	protected := router.Group("")
	protected.Use(middleware.AuthMiddleware(cfg))

	// Team routes
	protected.GET("/teams", teamController.GetUserTeams)
	protected.POST("/teams", teamController.CreateTeam)
	protected.GET("/teams/:id", teamController.GetTeam)
	protected.PUT("/teams/:id", teamController.UpdateTeam)
	protected.DELETE("/teams/:id", teamController.DeleteTeam)

	// Team members routes
	protected.GET("/teams/:id/members", teamController.GetTeamMembers)
	protected.POST("/teams/:id/members", teamController.AddTeamMember)
	protected.PUT("/teams/:id/members/:memberId", teamController.UpdateTeamMember)
	protected.DELETE("/teams/:id/members/:memberId", teamController.RemoveTeamMember)

	// Organization teams
	protected.GET("/organizations/:orgId/teams", teamController.GetOrganizationTeams)
}
