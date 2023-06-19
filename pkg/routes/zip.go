package routes

import (
	"github.com/gin-gonic/gin"
	"goFastCache/pkg/blobstorage"
	"goFastCache/pkg/upstream"
)

func HandleZip(c *gin.Context, version string, repo string, isShortRepo bool) {
	// Get the cache from the context
	bX := c.MustGet("blob").(*blobstorage.Blobstore)

	// Get :DOMAIN, :USER, :REPO from the context
	domain := c.Param("DOMAIN")
	user := c.Param("USER")

	// Get the info from blob
	mod, found, err := bX.GetModuleSourceObject(domain, user, repo, version)

	if err != nil {
		_ = c.AbortWithError(500, err)
		return
	}

	if !found {
		// Call upstream
		var status int
		mod, err, status = upstream.CallUpstreamModuleSource(domain, user, repo, version, isShortRepo)
		if err != nil {
			_ = c.AbortWithError(500, err)
			return
		}
		if status != 200 {
			c.Data(status, "text/plain; charset=utf-8", mod)
			return
		}

		// Set the mod in blob
		err = bX.PutModuleSourceObject(domain, user, repo, version, mod)
		if err != nil {
			_ = c.AbortWithError(500, err)
			return
		}
		c.Header("X-From-Cache", "false")
	} else {
		c.Header("X-From-Cache", "true")
		c.Header("X-From-Cache-Reason", "blob")
	}

	// Return the zip
	c.Data(200, "application/zip", mod)
}
