package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/your-username/slido-clone/user-service/api/controllers"
	"github.com/your-username/slido-clone/user-service/api/middleware"
	"github.com/your-username/slido-clone/user-service/config"
)

// RegisterOrganizationRoutes registers organization routes
func RegisterOrganizationRoutes(router *gin.RouterGroup, orgController *controllers.OrganizationController, cfg *config.JWTConfig) {
	// All organization routes require authentication
	protected := router.Group("")
	protected.Use(middleware.AuthMiddleware(cfg))

	// Organization routes
	protected.GET("/organizations", orgController.GetUserOrganizations)
	protected.POST("/organizations", orgController.CreateOrganization)
	protected.GET("/organizations/:id", orgController.GetOrganization)
	protected.PUT("/organizations/:id", orgController.UpdateOrganization)
	protected.DELETE("/organizations/:id", orgController.DeleteOrganization)

	// Organization members routes
	protected.GET("/organizations/:id/members", orgController.GetOrganizationMembers)
	protected.POST("/organizations/:id/members", orgController.AddOrganizationMember)
	protected.PUT("/organizations/:id/members/:memberId", orgController.UpdateOrganizationMember)
	protected.DELETE("/organizations/:id/members/:memberId", orgController.RemoveOrganizationMember)

	// Organization teams routes
	protected.GET("/organizations/:id/teams", orgController.GetOrganizationTeams)

	// Admin routes
	admin := router.Group("")
	admin.Use(middleware.AuthMiddleware(cfg), middleware.RoleMiddleware("admin"))
	admin.GET("/admin/organizations", orgController.ListOrganizations)
}
