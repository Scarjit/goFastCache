package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/united-manufacturing-hub/expiremap/pkg/expiremap"
	"go.uber.org/zap"
	"goFastCache/pkg/blobstorage"
	"goFastCache/pkg/cache"
	"goFastCache/pkg/hash"
	"goFastCache/pkg/upstream"
	"time"
)

var sumMap = expiremap.NewEx[string, string](time.Minute, time.Minute)

func SumDBRouter(c *gin.Context) {
	domain := c.Param("DOMAIN")
	trail := c.Param("TRAIL")
	cX := c.MustGet("cache").(*cache.Cache)
	bX := c.MustGet("blob").(*blobstorage.Blobstore)

	var found bool
	var sumX *string
	var err error
	sumX, found = sumMap.Get(hash.GetMinioSumPath(domain, trail))
	if found {
		c.Header("X-From-Cache", "true")
		c.Header("X-From-Cache-Reason", "memory")
		c.Data(200, "text/plain; charset=utf-8", []byte(*sumX))

		go func() {
			err = bX.PutSumObject(domain, trail, []byte(*sumX))
			if err != nil {
				zap.S().Error(err)
				return
			}
			err = cX.SetSumObj(domain, trail, []byte(*sumX))
			if err != nil {
				zap.S().Error(err)
				return
			}
		}()
		return
	}

	var sum string
	sum, err = cX.GetSumObj(domain, trail)
	if err == nil {
		c.Header("X-From-Cache", "true")
		c.Header("X-From-Cache-Reason", "cache")
		c.Data(200, "text/plain; charset=utf-8", []byte(sum))

		go func() {
			err = cX.SetSumObj(domain, trail, []byte(sum))
			if err != nil {
				zap.S().Error(err)
				return
			}
			err = bX.PutSumObject(domain, trail, []byte(sum))
			if err != nil {
				zap.S().Error(err)
				return
			}
		}()
		return
	}

	db, found, err := bX.GetSumObject(domain, trail)
	if err != nil {
		return
	}

	if found {
		c.Header("X-From-Cache", "true")
		c.Header("X-From-Cache-Reason", "blob")
		c.Data(200, "text/plain; charset=utf-8", db)
		go func() {
			err = cX.SetSumObj(domain, trail, db)
			if err != nil {
				zap.S().Error(err)
				return
			}
			sumMap.Set(hash.GetMinioSumPath(domain, trail), string(db))
		}()
		return
	}

	var status int
	db, err, status = upstream.CallUpstreamSumDB(domain, trail)
	if err != nil {
		_ = c.AbortWithError(500, err)
		return
	}

	if status != 200 {
		c.Data(status, "text/plain; charset=utf-8", db)
		return
	}
	c.Header("X-From-Cache", "false")
	c.Data(200, "text/plain; charset=utf-8", db)

	go func() {
		err = bX.PutSumObject(domain, trail, db)
		if err != nil {
			zap.S().Error(err)
			return
		}
		err = cX.SetSumObj(domain, trail, db)
		if err != nil {
			zap.S().Error(err)
			return
		}
		sumMap.Set(hash.GetMinioSumPath(domain, trail), string(db))
	}()
}
