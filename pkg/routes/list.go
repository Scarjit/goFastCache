package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/united-manufacturing-hub/expiremap/pkg/expiremap"
	"go.uber.org/zap"
	"goFastCache/pkg/cache"
	"goFastCache/pkg/hash"
	"goFastCache/pkg/upstream"
	"time"
)

var listMap = expiremap.NewEx[string, string](time.Minute, time.Second*30)

func HandleList(c *gin.Context, repo string, isShortRepo bool) {
	// Get the cache from the context
	cX := c.MustGet("cache").(*cache.Cache)

	// Get :DOMAIN, :USER, :REPO from the context
	domain := c.Param("DOMAIN")
	user := c.Param("USER")

	// Check if is in map
	var found bool
	var listX *string
	listX, found = listMap.Get(string(hash.GetHash(domain, user, repo)))
	if found {
		c.Header("X-From-Cache", "true")
		c.Header("X-From-Cache-Reason", "memory")
		c.Data(200, "text/plain; charset=utf-8", []byte(*listX))
		go func() {
			err := cX.SetList(domain, user, repo, *listX)
			if err != nil {
				zap.S().Error(err)
				return
			}
		}()
		return
	}

	// Get the list from the cache
	list, err := cX.GetList(domain, user, repo)
	if err != nil {
		var upstreamBytes []byte
		var status int
		upstreamBytes, err, status = upstream.CallUpstreamList(domain, user, repo, isShortRepo)
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
		c.Header("X-From-Cache", "false")
	} else {
		c.Header("X-From-Cache", "true")
		c.Header("X-From-Cache-Reason", "cache")
	}

	// Return the list
	c.Data(200, "text/plain; charset=utf-8", []byte(list))

	// Set the list in the map
	go listMap.Set(string(hash.GetHash(domain, user, repo)), list)
}
