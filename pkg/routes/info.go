package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
	"goFastCache/pkg/blobstorage"
	"goFastCache/pkg/upstream"
)

func HandleInfo(c *gin.Context, version string, repo string) {
	// Get the cache from the context
	bX := c.MustGet("blob").(*blobstorage.Blobstore)

	// Get :DOMAIN, :USER, :REPO from the context
	domain := c.Param("DOMAIN")
	user := c.Param("USER")

	// Get the info from blob
	info, found, err := bX.GetInfoObject(domain, user, repo, version)

	if err != nil {
		_ = c.AbortWithError(500, err)
		return
	}

	if !found {
		// Get from upstream
		var upstreamBytes []byte
		var status int
		upstreamBytes, err, status = upstream.CallUpstreamInfo(domain, user, repo, version)
		if err != nil {
			_ = c.AbortWithError(500, err)
			return
		}
		if status != 200 {
			c.Data(status, "text/plain; charset=utf-8", upstreamBytes)
			return
		}
		// Unmarshal the info
		err = json.Unmarshal(upstreamBytes, &info)
		if err != nil {
			zap.S().Errorf("Error unmarshalling upstream info: %s", err.Error())
			zap.S().Errorf("Upstream info: %s", upstreamBytes)
			_ = c.AbortWithError(500, err)
			return
		}
		// Set the info in blob
		err = bX.PutInfoObject(domain, user, repo, version, upstreamBytes)
		if err != nil {
			_ = c.AbortWithError(500, err)
			return
		}
		// Set "source" key
		info["source"] = "upstream"
	} else {
		// Set "source" key
		info["source"] = "blob"
	}

	// Return the list
	c.JSON(200, info)
}
