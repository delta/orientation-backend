package main

import (
	"fmt"
	"regexp"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/delta/orientation-backend/auth"
	"github.com/delta/orientation-backend/config"
	"github.com/delta/orientation-backend/core"
	"github.com/delta/orientation-backend/models"
)

func allowOrigin(origin string) (bool, error) {
	return regexp.MatchString(`^http:\/\/localhost:3000((\/).*)?$`, origin)
}

func main() {
	config.InitConfig()

	config.DB.AutoMigrate(&models.User{}, &models.SpriteSheet{})
	// Create dummy spritesheet for testing
	// for i := 1; i < 5; i++ {
	// 	config.DB.Create(&models.SpriteSheet{ID: i})
	// }

	port := config.Config("PORT")
	addr := fmt.Sprintf(":%s", port)

	e := echo.New()
	e.Validator = core.NewValidator()
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOriginFunc: allowOrigin,
		AllowMethods:    []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
		AllowHeaders: []string{
			echo.HeaderAccessControlRequestMethod,
			echo.HeaderAccessControlRequestHeaders,
			echo.HeaderContentType,
			echo.HeaderAccessControlAllowOrigin,
		},
		AllowCredentials: true,
		ExposeHeaders: []string{
			echo.HeaderAccessControlAllowHeaders,
			echo.HeaderAccessControlAllowOrigin,
			echo.HeaderAccessControlAllowMethods,
			echo.HeaderAccessControlAllowCredentials,
		},
	}))

	apiGroup := e.Group("/api", auth.AuthMiddlewareWrapper(auth.AuthMiddlewareConfig{
		Skipper: auth.SkipperFunc,
	}))

	core.RegisterRoutes(apiGroup)
	authGroup := apiGroup.Group("/auth")

	auth.RegisterRoutes(authGroup)

	e.Logger.Fatal(e.Start(addr))
}
