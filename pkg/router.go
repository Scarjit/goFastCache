package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"goFastCache/pkg/routes"
	"regexp"
)

func registerRoutes(router *gin.Engine) {
	router.GET("/*TRAIL", Router)
}

var uriRegex = regexp.MustCompile(`^/?([A-Za-z0-9_.\-~/]+)`)
var uriType = regexp.MustCompile(`(/@v/(.*).(zip|mod|info|list))$|(@latest)$`)

type Type int

const (
	RAW Type = iota
	LIST
	LATEST
	INFO
	MOD
	ZIP
	SUMDB
)

func getURIParts(rawUrl string) (uri string, version string, t Type, err error) {
	// Check if url starts with /sumdb/
	if rawUrl[:7] == "/sumdb/" {
		t = SUMDB
		uri = rawUrl[7:]
		return
	}

	// This is used to get the uri from the rawUrl
	matchesUri := uriRegex.FindStringSubmatch(rawUrl)
	if len(matchesUri) == 0 {
		err = errors.New("invalid path")
		return
	}
	uri = matchesUri[1]

	// This is used to get the type from the rawUrl
	// For zip, mod, info & list group 2 has the version and group 3 has the extension
	// For latest group 4 is set

	matchesType := uriType.FindStringSubmatch(rawUrl)
	if len(matchesType) == 0 {
		t = RAW
		return
	}
	if matchesType[4] != "" {
		t = LATEST
		return
	}
	switch matchesType[3] {
	case "zip":
		t = ZIP
	case "mod":
		t = MOD
	case "info":
		t = INFO
	case "list":
		t = LIST
	default:
		err = errors.New("invalid path")
		return
	}
	version = matchesType[2]
	return
}

func Router(c *gin.Context) {
	trail := c.Param("TRAIL")
	uri, version, t, err := getURIParts(trail)
	if err != nil {
		return
	}
	zap.S().Debugf("URI: %s, Version: %s, Type: %d", uri, version, t)
	switch t {
	case RAW:
		// TODO: Handle raw
	case LIST:
		routes.HandleList(c, uri)
	case LATEST:
		routes.HandleLatest(c, uri)
	case INFO:
		routes.HandleInfo(c, uri, version)
	case MOD:
		routes.HandleMod(c, uri, version)
	case ZIP:
		routes.HandleZip(c, uri, version)
	case SUMDB:
		routes.HandleSumdb(c, uri)

	}
}
