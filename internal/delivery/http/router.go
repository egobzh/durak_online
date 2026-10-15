package http

import (
	"durak-online/internal/delivery/http/middleware"
	"durak-online/internal/domain"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(
	authHandler *AuthHandler,
	authService domain.AuthService,
	userRepo domain.UserRepository,
) *gin.Engine {
	r := gin.Default()

	r.Use(gin.Recovery())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		profile := api.Group("/profile")
		profile.Use(middleware.AuthMiddleware(authService, userRepo))
		{
			profile.GET("/me", authHandler.GetProfile)
		}
	}

	return r
}
