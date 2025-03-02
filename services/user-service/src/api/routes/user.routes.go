package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/your-username/slido-clone/user-service/api/controllers"
	"github.com/your-username/slido-clone/user-service/api/middleware"
	"github.com/your-username/slido-clone/user-service/config"
)

// RegisterUserRoutes registers user routes
func RegisterUserRoutes(router *gin.RouterGroup, userController *controllers.UserController, cfg *config.JWTConfig) {
	// Public routes
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "UP",
		})
	})

	// Protected routes
	protected := router.Group("")
	protected.Use(middleware.AuthMiddleware(cfg))

	// Current user routes
	protected.GET("/me", userController.GetCurrentUser)
	protected.PUT("/me", userController.UpdateCurrentUser)

	// User routes
	protected.GET("/users", userController.ListUsers)
	protected.POST("/users", userController.CreateUser)
	protected.GET("/users/:id", userController.GetUser)
	protected.PUT("/users/:id", userController.UpdateUser)
	protected.DELETE("/users/:id", userController.DeleteUser)
	protected.POST("/users/:id/deactivate", userController.DeactivateUser)
	protected.POST("/users/:id/activate", userController.ActivateUser)
}
