package main

import (
	"backend/sicpapi/routes"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Terjadi kesalahan")
		}
	}()
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		ContentTypeNosniff: "nosniff",
	}))

	// mengatur cors
	e.Use(middleware.CORS())

	e.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(20),
				Burst:     30,
				ExpiresIn: 3 * time.Minute,
			},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			return ctx.RealIP(), nil
		},
		ErrorHandler: func(ctx echo.Context, err error) error {
			return ctx.JSON(http.StatusForbidden, map[string]interface{}{
				"success": false,
				"message": "Terjadi kesalahan saat memproses rate limit",
				"code":    http.StatusForbidden,
			})
		},
		DenyHandler: func(ctx echo.Context, identifier string, err error) error {
			return ctx.JSON(http.StatusTooManyRequests, map[string]interface{}{
				"success": false,
				"message": "Terlalu banyak request, silakan coba lagi nanti",
				"code":    http.StatusTooManyRequests,
			})
		},
	}))

	routes.SetupRoutes(e)

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		e.Logger.Fatal("Error loading .env file : ", err.Error())
	}

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	e.Logger.Fatal(e.Start(":" + port))
}
