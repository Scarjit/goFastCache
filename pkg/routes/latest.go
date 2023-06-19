package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/united-manufacturing-hub/expiremap/pkg/expiremap"
	"goFastCache/pkg/hash"
	"goFastCache/pkg/upstream"
	"time"
)

var cacheMap = expiremap.NewEx[string, []byte](time.Minute, time.Second*30)

func HandleLatest(c *gin.Context, repo string) {
	// Get :DOMAIN, :USER, :REPO from the context
	domain := c.Param("DOMAIN")
	user := c.Param("USER")

	hashX := hash.GetLatestHash(domain, user, repo)
	// Get the list from the cache
	latest, found := cacheMap.Get(hashX)
	if found {
		c.Header("X-From-Cache", "true")
		c.Header("X-From-Cache-Reason", "memory")
		c.Data(200, "text/plain; charset=utf-8", *latest)
		return
	}

	// Get latest version from upstream
	latestX, err, status := upstream.CallUpstreamLatest(domain, user, repo)
	if err != nil {
		_ = c.AbortWithError(500, err)
		return
	}

	if status != 200 {
		c.Data(status, "text/plain; charset=utf-8", latestX)
		return
	}

	cacheMap.Set(hashX, latestX)

	c.Header("X-From-Cache", "false")
	c.Data(200, "text/plain; charset=utf-8", latestX)
}
