package storage

import (
	"context"
	"go.uber.org/zap"
)

type BaseStorage interface {
	Connect(ctx context.Context, dsn string, l *zap.Logger) error
	Close() error
}
