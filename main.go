package main

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/delta/orientation-backend/auth"
	"github.com/delta/orientation-backend/config"
	"github.com/delta/orientation-backend/core"
	"github.com/delta/orientation-backend/leaderboard"
	"github.com/delta/orientation-backend/models"
	"github.com/delta/orientation-backend/videocall"
	"github.com/delta/orientation-backend/webhooks"
	"github.com/delta/orientation-backend/ws"
)

func main() {
	config.InitConfig()
	models.Init()
	ws.InitRooms()

	// broadcasts users position to each room every *x* seconds
	go ws.RoomBroadcast()

	port := config.Config("PORT")
	addr := fmt.Sprintf(":%s", port)

	e := echo.New()
	e.Validator = core.NewValidator()
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{config.Config("FRONTEND_URL")},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
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
	ws.RegisterRoutes(apiGroup)
	webhooks.RegisterRoutes(apiGroup)

	authGroup := apiGroup.Group("/auth")
	auth.RegisterRoutes(authGroup)
	leaderboard.RegisterRoutes(apiGroup)
	videocall.RegisterRoutes(apiGroup)
	e.Logger.Fatal(e.Start(addr))
}
