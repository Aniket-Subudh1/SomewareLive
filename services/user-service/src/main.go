package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/your-username/slido-clone/user-service/api/controllers"
	"github.com/your-username/slido-clone/user-service/api/middleware"
	"github.com/your-username/slido-clone/user-service/api/routes"
	"github.com/your-username/slido-clone/user-service/api/validators"
	"github.com/your-username/slido-clone/user-service/config"
	"github.com/your-username/slido-clone/user-service/db"
	"github.com/your-username/slido-clone/user-service/pkg/kafka"
	"github.com/your-username/slido-clone/user-service/pkg/logger"
	"github.com/your-username/slido-clone/user-service/repositories"
	"github.com/your-username/slido-clone/user-service/services"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(cfg)
	log.Info().Msg("Starting User Service")
	log.Debug().Interface("config", cfg.String()).Msg("Configuration loaded")

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Create context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to MongoDB
	mongoDB, err := db.New(&cfg.MongoDB)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer mongoDB.Close()

	// Create Kafka producer
	producer, err := kafka.NewProducer(&cfg.Kafka)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka producer")
	}
	defer producer.Close()

	// Create Kafka consumer
	consumer, err := kafka.NewConsumer(&cfg.Kafka)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka consumer")
	}
	defer consumer.Close()

	// Initialize repositories
	userRepo := repositories.NewUserRepository(mongoDB)
	teamRepo := repositories.NewTeamRepository(mongoDB)
	orgRepo := repositories.NewOrganizationRepository(mongoDB)

	// Initialize services
	userService := services.NewUserService(userRepo, producer)
	teamService := services.NewTeamService(teamRepo, userRepo, orgRepo, producer)
	orgService := services.NewOrganizationService(orgRepo, userRepo, teamRepo, producer)

	// Register Kafka event handlers
	consumer.RegisterHandler(
		cfg.Kafka.Topics.AuthEvents,
		kafka.UserCreated,
		userService.ProcessAuthUserCreated,
	)

	// Start Kafka consumer
	if err := consumer.Start(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to start Kafka consumer")
	}

	// Initialize controllers
	userController := controllers.NewUserController(userService)
	teamController := controllers.NewTeamController(teamService)
	orgController := controllers.NewOrganizationController(orgService)
	profileController := controllers.NewProfileController(userService, teamService, orgService)

	// Initialize validators
	validators.InitUserValidators()
	validators.InitTeamValidators()

	// Create router
	router := gin.New()

	// Add middlewares
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.RequestID())

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.CORS.AllowedOrigins},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Register routes
	apiGroup := router.Group("/api")
	routes.RegisterUserRoutes(apiGroup, userController, &cfg.JWT)
	routes.RegisterTeamRoutes(apiGroup, teamController, &cfg.JWT)
	routes.RegisterOrganizationRoutes(apiGroup, orgController, &cfg.JWT)
	routes.RegisterProfileRoutes(apiGroup, profileController, &cfg.JWT)
	routes.RegisterHealthRoutes(router.Group("/health"), mongoDB, producer)

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	log.Info().Str("port", cfg.Server.Port).Msg("Server started")

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Cancel context to stop Kafka consumer and other background tasks
	cancel()

	// Create a deadline for server shutdown
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exiting")
}
