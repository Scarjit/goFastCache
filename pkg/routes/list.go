package routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"goFastCache/pkg/cache"
	"goFastCache/pkg/upstream"
)

func HandleList(c *gin.Context, repo string) {
	// Get the cache from the context
	cX := c.MustGet("cache").(*cache.Cache)

	// Get :DOMAIN, :USER, :REPO from the context
	domain := c.Param("DOMAIN")
	user := c.Param("USER")

	// Get the list from the cache
	list, err := cX.GetList(domain, user, repo)
	if err != nil {
		var upstreamBytes []byte
		var status int
		upstreamBytes, err, status = upstream.CallUpstreamList(domain, user, repo)
		if err != nil {
			zap.S().Errorf("Error calling upstream: %s", err.Error())
			_ = c.AbortWithError(500, err)
			return
		}
		if status != 200 {
			c.Data(status, "text/plain; charset=utf-8", upstreamBytes)
			return
		}
		// Set the list in the cache
		err = cX.SetList(domain, user, repo, string(upstreamBytes))
		if err != nil {
			zap.S().Errorf("Error setting list in cache: %s", err.Error())
			_ = c.AbortWithError(500, err)
			return
		}
		list = string(upstreamBytes)
	}

	// Return the list
	c.Data(200, "text/plain; charset=utf-8", []byte(list))
}
