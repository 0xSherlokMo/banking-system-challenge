package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthRouter struct{}

func InstallHealthRouter(engine *gin.Engine) HealthRouter {
	healthRouter := HealthRouter{}

	healthRouter.install(
		engine.Group("/health"),
	)

	return healthRouter
}

func (h *HealthRouter) install(router *gin.RouterGroup) {
	router.GET("/", h.ping)
}

func (h *HealthRouter) ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
