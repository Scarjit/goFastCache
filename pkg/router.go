package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"goFastCache/pkg/routes"
	"regexp"
	"strings"
)

var regexModZipInfo = regexp.MustCompile(`(v\d+\.\d+\.\d+(-\w+)?)\.(mod|zip|info)`)
var regexRepo = regexp.MustCompile(`^/(.+)/@`)

func registerRoutes(router *gin.Engine) {
	router.GET("/sumdb/:DOMAIN/*TRAIL", routes.SumDBRouter)
	router.GET("/:DOMAIN/:USER/*TRAIL", CustomRouter)
}

func CustomRouter(context *gin.Context) {
	// Get trail from context
	trail := context.Param("TRAIL")
	// Get repo from trail
	repoX := regexRepo.FindStringSubmatch(trail)
	if len(repoX) == 0 {
		_ = context.AbortWithError(400, errors.New("invalid path"))
		return
	}
	repo := repoX[1]

	if strings.HasSuffix(trail, "@latest") {
		routes.HandleLatest(context, repo)
		return
	} else if strings.HasSuffix(trail, "list") {
		routes.HandleList(context, repo)
		return
	}

	zap.S().Infof("trail: %s", trail)
	matches := regexModZipInfo.FindStringSubmatch(trail)
	if len(matches) == 0 {
		_ = context.AbortWithError(400, errors.New("invalid path"))
	} else {
		version := matches[1]
		extension := matches[3]
		VersionRouter(context, version, extension, repo)
	}
}

func VersionRouter(c *gin.Context, version, extension, repo string) {

	switch extension {
	case "mod":
		routes.HandleMod(c, version, repo)
		return
	case "zip":
		routes.HandleZip(c, version, repo)
		return
	case "info":
		routes.HandleInfo(c, version, repo)
		return
	default:
		_ = c.AbortWithError(400, errors.New("unknown file type: "+extension))
	}
}
