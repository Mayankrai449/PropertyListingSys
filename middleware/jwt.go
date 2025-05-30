package middleware

import (
	"PropertyListingSys/utils"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func JWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Authorization header is required",
				})
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid authorization header format",
				})
			}

			tokenString := tokenParts[1]
			claims, err := utils.ValidateJWT(tokenString)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid token",
				})
			}

			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)
			c.Set("user_role", claims.Role)

			return next(c)
		}
	}
}
