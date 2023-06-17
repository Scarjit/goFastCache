package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"goFastCache/pkg/routes"
)

func registerRoutes(router *gin.Engine) {
	router.GET("/:DOMAIN/:USER/:REPO/@v/list", routes.HandleList)
	router.GET("/:DOMAIN/:USER/:REPO/@v/:VERSION", VRouter)
	router.GET("/:DOMAIN/:USER/:REPO/@latest", routes.HandleLatest)
}

func VRouter(c *gin.Context) {
	// switch on ending of :VERSION

	d := c.Param("VERSION")

	if len(d) < 4 {
		_ = c.AbortWithError(400, errors.New("unknown file type"))
		return
	}

	switch d[len(d)-4:] {
	case ".mod":
		routes.HandleMod(c)
	case ".zip":
		routes.HandleZip(c)
	case ".info":
		routes.HandleInfo(c)
	default:
		_ = c.AbortWithError(400, errors.New("unknown file type: "+d[len(d)-4:]))
	}
}
