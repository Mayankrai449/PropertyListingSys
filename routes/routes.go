package routes

import (
	"PropertyListingSys/handlers"
	"PropertyListingSys/middleware"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo) {
	e.GET("/health", handlers.HealthCheck)

	userController := handlers.NewUserController()

	auth := e.Group("/api/auth")
	auth.POST("/register", userController.Register)
	auth.POST("/login", userController.Login)

	api := e.Group("/api")
	api.Use(middleware.JWTMiddleware())

	users := api.Group("/users")
	users.GET("/profile", userController.GetProfile)
	users.PUT("/profile", userController.UpdateProfile)
	users.DELETE("/profile", userController.DeleteAccount)
	users.GET("", userController.GetAllUsers)
}
