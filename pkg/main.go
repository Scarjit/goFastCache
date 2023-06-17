package main

import (
	"fmt"
	"goFastCache/pkg/blobstorage"
	"goFastCache/pkg/database"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	logger *zap.Logger
)

func main() {
	// Initialize logger
	initLogger()

	// Initialize blob storage
	blob, err := blobstorage.NewBlobstore()
	if err != nil {
		zap.S().Fatalf("Unable to connect to Minio: %v", err)
	}

	// Initialize database
	var db *database.Database
	db, err = database.NewDatabase()
	if err != nil {
		zap.S().Fatalf("Unable to connect to Postgres: %v", err)
	}

	// Initialize router
	router := gin.Default()

	// Use middleware to store the db and minioClient in the context
	router.Use(func(c *gin.Context) {
		//c.Set("db", db)
		c.Set("blob", blob)
		c.Set("db", db)
		c.Next()
	})

	// Register routes
	registerRoutes(router)

	// Start server
	router.Run()
}

func initLogger() {
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		fmt.Printf("Can't create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)
}
