package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luisforni/nexus-logistics/backend/internal/api/handlers"
	"github.com/luisforni/nexus-logistics/backend/internal/api/middleware"
	"github.com/luisforni/nexus-logistics/backend/internal/config"
	"github.com/luisforni/nexus-logistics/backend/internal/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(cfg *config.Config, authSvc *service.AuthService, shipmentSvc *service.ShipmentService) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	r.Use(middleware.RequestID())
	r.Use(middleware.StructuredLogger())
	r.Use(middleware.Recovery())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORS(cfg.AppEnv))
	r.Use(middleware.RateLimit(200, time.Minute))

	r.GET("/health", handlers.Health)
	r.GET("/ready", handlers.Ready)

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := r.Group("/api/v1")

	authHandler := handlers.NewAuthHandler(authSvc)
	auth := v1.Group("/auth")
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.Refresh)

	protected := v1.Group("")
	protected.Use(middleware.JWTAuth(authSvc))

	shipHandler := handlers.NewShipmentHandler(shipmentSvc)
	shipments := protected.Group("/shipments")
	shipments.GET("", shipHandler.List)
	shipments.POST("", shipHandler.Create)
	shipments.GET("/:id", shipHandler.GetByID)
	shipments.PUT("/:id/status", shipHandler.UpdateStatus)
	shipments.GET("/:id/trace", shipHandler.GetBlockchainTrace)

	analyticsHandler := handlers.NewAnalyticsHandler()
	analytics := protected.Group("/analytics")
	analytics.GET("/forecast", analyticsHandler.DemandForecast)
	analytics.GET("/kpis", analyticsHandler.KPIs)

	optimizeHandler := handlers.NewOptimizerHandler(cfg.OptimizerHost, cfg.OptimizerPort)
	optimize := protected.Group("/optimize")
	optimize.POST("/route", optimizeHandler.OptimizeRoute)

	return r
}
