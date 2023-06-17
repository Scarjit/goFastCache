package database

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"os"
	"strings"
)

type Database struct {
	ConnectionPool *pgxpool.Pool
}

func NewDatabase() (*Database, error) {
	zap.S().Info("Connecting to Postgres")
	dbPassword, found := os.LookupEnv("DB_PASSWORD")
	if !found {
		return nil, errors.New("DB_PASSWORD not found")
	}
	dbPassword = strings.Trim(dbPassword, "\n\r")

	dbUser, found := os.LookupEnv("DB_USER")
	if !found {
		return nil, errors.New("DB_USER not found")
	}
	dbUser = strings.Trim(dbUser, "\n\r")

	dbName, found := os.LookupEnv("DB_NAME")
	if !found {
		return nil, errors.New("DB_NAME not found")
	}
	dbName = strings.Trim(dbName, "\n\r")

	dbHost, found := os.LookupEnv("DB_HOST")
	if !found {
		return nil, errors.New("DB_HOST not found")
	}
	dbHost = strings.Trim(dbHost, "\n\r")

	db := &Database{}

	var err error
	db.ConnectionPool, err = pgxpool.Connect(context.Background(), "postgres://"+dbUser+":"+dbPassword+"@"+dbHost+"/"+dbName)
	if err != nil {
		return nil, err
	}

	zap.S().Info("Connected to Postgres")
	return db, nil
}

func (d *Database) createTables() error {
	// Creat 'list' table (if not exists)
	_, err := d.ConnectionPool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS list (
		    			id SERIAL PRIMARY KEY,
		    			hash TEXT NOT NULL,
		    			version TEXT NOT NULL,
		    			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		    		);
	`)
	return err
}
