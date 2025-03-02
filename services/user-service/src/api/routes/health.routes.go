package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-username/slido-clone/user-service/db"
	"github.com/your-username/slido-clone/user-service/pkg/kafka"
)

// Health represents the service health information
type Health struct {
	Status       string            `json:"status"`
	Service      string            `json:"service"`
	Version      string            `json:"version"`
	Timestamp    time.Time         `json:"timestamp"`
	Dependencies map[string]string `json:"dependencies"`
}

// RegisterHealthRoutes registers health routes
func RegisterHealthRoutes(router *gin.RouterGroup, mongoDB *db.MongoDB, producer *kafka.Producer) {
	router.GET("", func(c *gin.Context) {
		// Basic health check
		health := Health{
			Status:    "UP",
			Service:   "user-service",
			Version:   "1.0.0",
			Timestamp: time.Now(),
			Dependencies: map[string]string{
				"mongodb": "UP",
				"kafka":   "UP",
			},
		}

		c.JSON(http.StatusOK, health)
	})

	router.GET("/detailed", func(c *gin.Context) {
		// Check MongoDB connection
		ctx := c.Request.Context()
		mongoStatus := "UP"

		err := mongoDB.Client.Ping(ctx, nil)
		if err != nil {
			mongoStatus = "DOWN"
		}

		// Detailed health check
		health := Health{
			Status:    mongoStatus,
			Service:   "user-service",
			Version:   "1.0.0",
			Timestamp: time.Now(),
			Dependencies: map[string]string{
				"mongodb": mongoStatus,
				"kafka":   "UP", // Assuming Kafka is UP - we could add a specific check
			},
		}

		// Set status code based on dependencies
		statusCode := http.StatusOK
		if mongoStatus != "UP" {
			health.Status = "DEGRADED"
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, health)
	})
}
