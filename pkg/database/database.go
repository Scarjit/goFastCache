package database

import (
	"errors"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	postgres *gorm.DB
}

type Gomodule struct {
	gorm.Model
	Index   int    `gorm:"primaryKey"`
	Path    string `gorm:"unique"`
	Version string
}

func NewDatabase() (*Database, error) {
	postgresUser, found := os.LookupEnv("POSTGRES_USER")
	if !found {
		return nil, errors.New("POSTGRES_USER not found")
	}

	postgresPassword, found := os.LookupEnv("POSTGRES_PASSWORD")
	if !found {
		return nil, errors.New("POSTGRES_PASSWORD not found")
	}

	postgresHost, found := os.LookupEnv("POSTGRES_HOST")
	if !found {
		return nil, errors.New("POSTGRES_HOST not found")
	}

	postgresDB, found := os.LookupEnv("POSTGRES_DB")
	if !found {
		return nil, errors.New("POSTGRES_DB not found")
	}

	dsn := fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=disable", postgresUser, postgresPassword, postgresHost, postgresDB)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Create the table if it doesn't exist
	err = db.AutoMigrate(&Gomodule{})
	if err != nil {
		return nil, err
	}

	return &Database{
		postgres: db,
	}, nil
}

func (db *Database) InsertGomodule(gomodule Gomodule) error {
	result := db.postgres.Create(&gomodule)
	return result.Error
}

func (db *Database) GetGomoduleByPath(path string) (Gomodule, error) {
	var gomodule Gomodule
	result := db.postgres.First(&gomodule, "path = ?", path)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return Gomodule{}, nil
		}
		return Gomodule{}, result.Error
	}

	return gomodule, nil
}
