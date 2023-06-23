package routes

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/united-manufacturing-hub/expiremap/pkg/expiremap"
	"go.uber.org/zap"
	"goFastCache/pkg/blobstorage"
	"goFastCache/pkg/cache"
	"goFastCache/pkg/database"
	"goFastCache/pkg/hash"
	"goFastCache/pkg/upstream"
	"time"
)

var ListExpireMap = expiremap.NewEx[string, []byte](time.Minute, time.Second*30)
var LatestExpireMap = expiremap.NewEx[string, []byte](time.Minute, time.Second*30)

var ThirtySeconds = time.Second * 30
var OneMinute = time.Minute

func HandleList(c *gin.Context, uri string) {
	blob := c.MustGet("blob").(*blobstorage.Blobstore)
	cacheX := c.MustGet("cache").(*cache.Cache)
	list, err, status := GetList(uri, ListExpireMap, cacheX, blob)
	if list != nil {
		c.Data(200, "text/plain; charset=utf-8", list)
		return
	}
	if err != nil {
		c.Data(status, "text/plain; charset=utf-8", []byte(err.Error()))
	}
	c.Status(404)
}

func HandleLatest(c *gin.Context, uri string) {
	cacheX := c.MustGet("cache").(*cache.Cache)
	list, err, status := GetLatest(uri, LatestExpireMap, cacheX)
	if list != nil {
		c.Data(200, "text/plain; charset=utf-8", list)
		return
	}
	if err != nil {
		c.Data(status, "text/plain; charset=utf-8", []byte(err.Error()))
	}
	c.Status(404)
}

func HandleInfo(c *gin.Context, uri, version string) {
	blob := c.MustGet("blob").(*blobstorage.Blobstore)
	db := c.MustGet("db").(*database.Database)

	info, err, status := GetInfo(uri, version, db, blob)
	if info != nil {
		c.Data(200, "application/json; charset=utf-8", info)
		return
	}
	if err != nil {
		c.Data(status, "text/plain; charset=utf-8", []byte(err.Error()))
	}
	c.Status(404)
}

func HandleMod(c *gin.Context, uri, version string) {
	blob := c.MustGet("blob").(*blobstorage.Blobstore)
	db := c.MustGet("db").(*database.Database)

	mod, err, status := GetMod(uri, version, db, blob)
	if mod != nil {
		c.Data(200, "text/plain; charset=utf-8", mod)
		return
	}
	if err != nil {
		c.Data(status, "text/plain; charset=utf-8", []byte(err.Error()))
	}
	c.Status(404)
}

func HandleZip(c *gin.Context, uri, version string) {
	blob := c.MustGet("blob").(*blobstorage.Blobstore)
	db := c.MustGet("db").(*database.Database)

	zip, err, status := GetZip(uri, version, db, blob)
	if zip != nil {
		c.Data(200, "application/zip", zip)
		return
	}
	if err != nil {
		c.Data(status, "text/plain; charset=utf-8", []byte(err.Error()))
	}
	c.Status(404)
}

func HandleSumdb(c *gin.Context, uri string) {
	zap.S().Debugf("sumdb request: %s", uri)
	c.AbortWithStatus(404)
}

func GetInfo(uri string, version string, db *database.Database, blob *blobstorage.Blobstore) ([]byte, error, int) {
	return GetX(uri, version, upstream.CallUpstreamInfo, nil, nil, blob, nil, db)
}

func GetList(uri string, memcache *expiremap.ExpireMap[string, []byte], cacheX *cache.Cache, blob *blobstorage.Blobstore) ([]byte, error, int) {
	return GetXNoVersion(uri, upstream.CallUpstreamList, memcache, cacheX, blob, &OneMinute)
}

func GetLatest(uri string, memcache *expiremap.ExpireMap[string, []byte], cacheX *cache.Cache) ([]byte, error, int) {
	return GetXNoVersion(uri, upstream.CallUpstreamLatest, memcache, cacheX, nil, &ThirtySeconds)
}

func GetMod(uri string, version string, db *database.Database, blob *blobstorage.Blobstore) ([]byte, error, int) {
	return GetX(uri, version, upstream.CallUpstreamMod, nil, nil, blob, nil, db)
}

func GetZip(uri string, version string, db *database.Database, blob *blobstorage.Blobstore) ([]byte, error, int) {
	return GetX(uri, version, upstream.CallUpstreamZip, nil, nil, blob, nil, db)
}

func GetXNoVersion(uri string, upstreamHandler func(uri string) ([]byte, error, int), memcache *expiremap.ExpireMap[string, []byte], cacheX *cache.Cache, blob *blobstorage.Blobstore, cacheTTL *time.Duration) ([]byte, error, int) {
	cacheKey := hash.GetListPath(uri)
	list, foundInCache := CachedLookup(cacheKey, memcache, cacheX, blob)
	if foundInCache {
		return list, nil, 200
	}
	upstreamList, err, status := upstreamHandler(uri)
	if err != nil {
		return nil, err, status
	}
	if status != 200 {
		return nil, fmt.Errorf("upstream returned status %d (%s)", status, upstreamList), status
	}

	SetCache(cacheKey, upstreamList, memcache, cacheX, blob, cacheTTL)

	return upstreamList, nil, status
}

func GetX(uri, version string, upstreamHandler func(uri string, version string) ([]byte, error, int), memcache *expiremap.ExpireMap[string, []byte], cacheX *cache.Cache, blob *blobstorage.Blobstore, cacheTTL *time.Duration, db *database.Database) ([]byte, error, int) {
	cacheKey := hash.GetListPath(uri)
	list, foundInCache := CachedLookup(cacheKey, memcache, cacheX, blob)
	if foundInCache {
		return list, nil, 200
	}
	upstreamList, err, status := upstreamHandler(uri, version)
	if err != nil {
		return nil, err, status
	}
	if status != 200 {
		return nil, fmt.Errorf("upstream returned status %d (%s)", status, upstreamList), status
	}

	SetCache(cacheKey, upstreamList, memcache, cacheX, blob, cacheTTL)

	if db != nil {
		// Upsert into database
		err = db.UpsertGoModule(database.Gomodule{
			Path:    uri,
			Version: version,
		})
		if err != nil {
			zap.S().Warnw("Failed to upsert gomodule", "error", err)
		}
	}

	return upstreamList, nil, status
}

func CachedLookup(cacheKey string, memcache *expiremap.ExpireMap[string, []byte], cacheX *cache.Cache, blob *blobstorage.Blobstore) ([]byte, bool) {
	if memcache != nil {
		if k, found := ListExpireMap.Get(cacheKey); found {
			kX := *k
			return kX, true
		}
	}
	if cacheX != nil {
		if k, found, _ := cacheX.Get(cacheKey); found {
			return k, true
		}
	}
	if blob != nil {
		if k, found := blob.Get(cacheKey); found {
			return k, true
		}
	}
	return nil, false
}

func SetCache(cacheKey string, value []byte, memcache *expiremap.ExpireMap[string, []byte], cacheX *cache.Cache, blob *blobstorage.Blobstore, cacheTTL *time.Duration) {
	if memcache != nil {
		memcache.Set(cacheKey, value)
	}
	var err error
	if cacheX != nil {
		if cacheTTL != nil {
			cacheTTLX := time.Minute
			cacheTTL = &cacheTTLX
		}
		err = cacheX.Set(cacheKey, value, *cacheTTL)
		if err != nil {
			zap.S().Errorf("Error setting cache: %s", err.Error())
		}
	}
	if blob != nil {
		err = blob.Put(cacheKey, value)
		if err != nil {
			zap.S().Errorf("Error setting cache: %s", err.Error())
		}
	}
	return
}
