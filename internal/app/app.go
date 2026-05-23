package app

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/horlakz/jarakey.assessment/internal/authorization"
	"github.com/horlakz/jarakey.assessment/internal/config"
	"github.com/horlakz/jarakey.assessment/internal/database"
	"github.com/horlakz/jarakey.assessment/internal/handlers"
	"github.com/horlakz/jarakey.assessment/internal/middleware"
	"github.com/horlakz/jarakey.assessment/internal/repositories"
	"github.com/horlakz/jarakey.assessment/internal/services"
	"github.com/horlakz/jarakey.assessment/internal/utils"
	"gorm.io/gorm"
)

type Application struct {
	Fiber  *fiber.App
	DB     *gorm.DB
	Config config.Config
}

func Build(cfg config.Config) (*Application, error) {
	db, err := database.Connect(cfg)
	if err != nil {
		return nil, err
	}
	if _, err := database.SeedDefaults(db, cfg); err != nil {
		return nil, err
	}

	jwtManager := utils.NewJWTManager(cfg.JWTSecret)
	userRepo := repositories.NewUserRepository(db)
	authzRepo := repositories.NewAuthorizationRepository(db)
	auditRepo := repositories.NewAuditRepository(db)

	authorizer := authorization.NewService(authzRepo)
	authService := services.NewAuthService(userRepo, jwtManager)
	userService := services.NewUserService(userRepo)
	gateService := services.NewGateService(authorizer, auditRepo)
	debugService := services.NewDebugService(authorizer)

	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	gateHandler := handlers.NewGateHandler(gateService)
	debugHandler := handlers.NewDebugHandler(debugService)

	server := fiber.New(fiber.Config{
		ErrorHandler: handlers.ErrorHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	})

	server.Use(middleware.RequestID())
	server.Use(middleware.StructuredLogger())
	server.Use(fiberrecover.New())

	server.Post("/auth/login", authHandler.Login)

	protected := server.Group("/", middleware.AuthRequired(jwtManager))
	protected.Get("/me", userHandler.Me)
	protected.Post("/gate/open", gateHandler.Open)
	protected.Post("/debug/downgrade-role", debugHandler.DowngradeRole)

	return &Application{
		Fiber:  server,
		DB:     db,
		Config: cfg,
	}, nil
}

func (a *Application) Shutdown(ctx context.Context) error {
	sqlDB, err := a.DB.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}

	shutdownErr := a.Fiber.ShutdownWithContext(ctx)
	closeErr := sqlDB.Close()

	if shutdownErr != nil {
		return shutdownErr
	}
	return closeErr
}
