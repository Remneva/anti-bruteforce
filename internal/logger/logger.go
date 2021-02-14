package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/dchest/safefile"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(level zapcore.Level, env string, outputfile string) (*zap.Logger, error) {
	err := mkFile(outputfile)
	if err != nil {
		return nil, errors.New("Error file creating")
	}
	var cfg zap.Config
	switch {
	case env == "prod":
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.EncodeLevel = customLevelEncoder
	default:
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	}
	cfg.EncoderConfig.EncodeTime = syslogTimeEncoder
	cfg.OutputPaths = []string{outputfile, "stdout"}
	cfg.ErrorOutputPaths = []string{"stderr"}
	cfg.Level = zap.NewAtomicLevelAt(level)

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("logger build failed: %w", err)
	}

	return logger, nil
}

func syslogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("Jan 2 15:04:05"))
}

func customLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + level.String() + "]")
}

func mkFile(path string) error {
	existFile := exists(path)
	if !existFile {
		err := os.Mkdir("/tmp/tmpdir", 0755)
		if err != nil {
			return fmt.Errorf("mkdir failed: %w", err)
		}
		tmpfile, err := safefile.Create(path, 0755)
		if err != nil {
			return fmt.Errorf("create tmpfile failed: %w", err)
		}
		defer tmpfile.Close()
		return nil
	}
	return nil
}

func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
