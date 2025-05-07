package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	ServeDataBaseName = "gd-tools-serve.db"
)

var (
	ServeDB *gorm.DB
)

func InitServeDB() error {
	var err error
	ServeDB, err = gorm.Open(sqlite.Open(ServeDataBaseName), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, err := ServeDB.DB()
	if err != nil {
		return err
	}

	if err := sqlDB.Ping(); err != nil {
		return err
	}

	return nil
}
