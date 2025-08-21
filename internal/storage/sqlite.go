package storage

import (
	"context"
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Store struct{ DB *gorm.DB }

func Open(ctx context.Context, path string) (*Store, error) {
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_synchronous=NORMAL", path)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.WithContext(ctx).AutoMigrate(&Session{}, &Command{}, &Output{}); err != nil {
		return nil, err
	}
	return &Store{DB: db}, nil
}
