package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"goFastCache/pkg/blobstorage"
	"goFastCache/pkg/cache"
	"goFastCache/pkg/logger"
)

func main() {
	// Initialize logger
	logger.InitLogger()

	// Initialize blob storage
	blob, err := blobstorage.NewBlobstore()
	if err != nil {
		zap.S().Fatalf("Unable to connect to Minio: %v", err)
	}

	// Initialize cache
	var cacheX *cache.Cache
	cacheX, err = cache.NewCache()
	if err != nil {
		zap.S().Fatalf("Unable to connect to Redis: %v", err)
	}

	// Initialize router
	router := gin.Default()

	// Use middleware to store the db and minioClient in the context
	router.Use(func(c *gin.Context) {
		// c.Set("db", db)
		c.Set("blob", blob)
		c.Set("cache", cacheX)
		c.Next()
	})

	// Register routes
	registerRoutes(router)

	// Set error handler for missing routes
	router.NoRoute(func(c *gin.Context) {
		c.String(404, fmt.Sprintf("Route %s not found", c.Request.URL.Path))
	})

	// Start server
	router.Run()
}
