package routes

import (
	"PropertyListingSys/handlers"
	"PropertyListingSys/middleware"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo) {
	e.GET("/health", handlers.HealthCheck)

	userController := handlers.NewUserController()
	propertyController := handlers.NewPropertyController()
	favoriteController := handlers.NewFavoriteController()
	recommendationController := handlers.NewRecommendationController()

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
	users.GET("/search", userController.SearchUserByEmail)

	properties := api.Group("/properties")
	properties.POST("", propertyController.CreateProperty)
	properties.PUT("/:id", propertyController.UpdateProperty)
	properties.DELETE("/:id", propertyController.DeleteProperty)
	e.GET("/properties", propertyController.ListProperties)
	e.GET("/properties/:id", propertyController.GetProperty)

	favorites := api.Group("/favorites")
	favorites.POST("", favoriteController.CreateFavorite)
	favorites.GET("", favoriteController.GetFavorites)
	favorites.DELETE("/:propertyId", favoriteController.DeleteFavorite)

	recommendations := api.Group("/recommendations")
	recommendations.POST("", recommendationController.CreateRecommendation)
	recommendations.GET("/received", recommendationController.GetReceivedRecommendations)
}
