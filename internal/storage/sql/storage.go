package sql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Remneva/anti-bruteforce/internal/storage"

	// Postgres driver.
	_ "github.com/jackc/pgx/v4/stdlib"
	"go.uber.org/zap"
)

var _ storage.BaseStorage = (*Storage)(nil)

type Storage struct {
	DB *sql.DB
	l  *zap.Logger
	storage.BaseStorage
	storage.Lists
	storage.Configurations
}

func NewDB(l *zap.Logger) *Storage {
	db := &Storage{
		l: l,
	}
	return db
}
func (s *Storage) Connect(ctx context.Context, dsn string, l *zap.Logger) (err error) {
	s.DB, err = sql.Open("pgx", dsn)
	if err != nil {
		s.l.Error("Error", zap.String("Open connection", err.Error()))
		return fmt.Errorf("open connection error %w", err)
	}
	err = s.DB.PingContext(ctx)
	if err != nil {
		s.l.Error("Error", zap.String("Ping", err.Error()))
		return fmt.Errorf("ping error: %w", err)
	}
	l.Info("PSQL connection")
	return nil
}

func (s *Storage) Close() error {
	return s.DB.Close()
}
