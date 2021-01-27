package configs

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Logger LoggerConfig
	PSQL   PSQLConfig
	Port   PortConfig
	Redis  RedisConfig
}

type ModeConfig struct {
	MemMode bool
}

type LoggerConfig struct {
	Level zapcore.Level
	Path  string
}

type PSQLConfig struct {
	DSN string
}

type PortConfig struct {
	HTTP string
	Grpc string
}

type RedisConfig struct {
	Address      string
	Password     string
	ExpiryPeriod time.Duration
}

func Read(path string) (c Config, err error) {
	_, err = toml.DecodeFile(path, &c)
	if err != nil {
		return Config{}, fmt.Errorf("decodeFile failed: %w", err)
	}
	return
}
