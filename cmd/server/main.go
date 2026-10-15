package main

import (
	"log"

	_ "durak-online/docs"
	"durak-online/internal/config"
	deliveryHttp "durak-online/internal/delivery/http"
	"durak-online/internal/notifier"
	"durak-online/internal/repository/postgres"
	"durak-online/internal/repository/postgres/resettoken"
	"durak-online/internal/repository/postgres/user"
	"durak-online/internal/service"
	"durak-online/pkg/hasher"
	"durak-online/pkg/token"
)

// @title Durak Online Core API
// @version 1.0
// @description Бэкенд-сервис карточной игры «Дурак Онлайн» (MVP).
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите "Bearer <токен>" для авторизации.
func main() {
	cfg := config.LoadConfig()

	log.Printf("Connecting to database...")
	pg, err := postgres.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer pg.SQL.Close()
	log.Printf("Database connected and schema verified.")

	userRepo := user.NewRepository(pg.SQL)
	resetRepo := resettoken.NewRepository(pg.SQL)
	bcryptHasher := hasher.NewBcryptHasher(10)
	jwtManager := token.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiration)
	consoleSender := notifier.NewConsoleResetLinkSender(cfg.ResetBaseURL)

	authService := service.NewAuthService(
		userRepo,
		resetRepo,
		bcryptHasher,
		jwtManager,
		consoleSender,
		cfg.ResetTokenTTLMin,
	)

	authHandler := deliveryHttp.NewAuthHandler(authService)
	router := deliveryHttp.SetupRouter(authHandler, authService, userRepo)

	log.Printf("Starting Durak Online HTTP Server on port :%s", cfg.ServerPort)
	log.Printf("Swagger documentation available at http://localhost:%s/swagger/index.html", cfg.ServerPort)

	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
