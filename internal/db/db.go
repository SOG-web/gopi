package db

import (
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"gopi.com/config"
)

func NewSqliteDb(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(cfg.DBAddress), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func NewMysqlDb(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DBAddress), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func NewPostgresDb(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DBAddress), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}