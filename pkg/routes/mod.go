package routes

import (
	"github.com/gin-gonic/gin"
	"goFastCache/pkg/blobstorage"
	"goFastCache/pkg/upstream"
)

func HandleMod(c *gin.Context, version string, repo string) {
	// Get the cache from the context
	bX := c.MustGet("blob").(*blobstorage.Blobstore)

	// Get :DOMAIN, :USER, :REPO from the context
	domain := c.Param("DOMAIN")
	user := c.Param("USER")

	// Get the info from blob
	mod, found, err := bX.GetModObject(domain, user, repo, version)

	if err != nil {
		_ = c.AbortWithError(500, err)
		return
	}

	if !found {
		// Call upstream
		var status int
		mod, err, status = upstream.CallUpstreamMod(domain, user, repo, version)
		if err != nil {
			_ = c.AbortWithError(500, err)
			return
		}
		if status != 200 {
			c.Data(status, "text/plain; charset=utf-8", mod)
			return
		}

		// Set the mod in blob
		err = bX.PutModObject(domain, user, repo, version, mod)
		if err != nil {
			_ = c.AbortWithError(500, err)
			return
		}
	}

	// Return the mod
	c.Data(200, "text/plain; charset=utf-8", mod)
}
