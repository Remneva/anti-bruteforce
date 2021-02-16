package storage

import (
	"context"

	"go.uber.org/zap"
)

//go:generate mockgen -destination=mock_db_test.go -package=storage . BaseStorage
type BaseStorage interface {
	Connect(ctx context.Context, dsn string, l *zap.Logger) error
	Close() error
	Configs() Configurations
	List() Lists
}

type Configurations interface {
	Get(ctx context.Context) (map[string]int64, error)
}

type Lists interface {
	AddToWhiteList(ctx context.Context, ip IP) error
	AddToBlackList(ctx context.Context, ip IP) error
	DeleteFromWhiteList(ctx context.Context, ip IP) error
	DeleteFromBlackList(ctx context.Context, ip IP) error
	GetFromWhiteList(ip IP) (bool, error)
	GetFromBlackList(ip IP) (bool, error)
	GetAllFromWhiteList(ctx context.Context) ([]string, error)
	GetAllFromBlackList(ctx context.Context) ([]string, error)
}

type IP struct {
	IP   string
	Mask string
}

type Config struct {
	Key   string
	Value int64
}

type Auth struct {
	Login    string
	Password string
	IP       string
}

type User struct {
	Login string
	IP    string
}

type Address struct {
	ID   int64
	IP   string
	Mask string
}
