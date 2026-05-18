package database

import (
	"database/sql"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

type Database struct {
	DB    *gorm.DB
	sqlDB *sql.DB
}

func Connect() (*Database, error) {
	dsn := "host=postgres user=postgres password=admin dbname=filmsdb port=5432 sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()

	if err != nil {
		return nil, fmt.Errorf("error to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Println("Successfully to connect database")

	return &Database{
		DB:    db,
		sqlDB: sqlDB,
	}, nil
}

func (d *Database) Close() error {
	if d.sqlDB != nil {
		return d.sqlDB.Close()
	}
	return nil
}
