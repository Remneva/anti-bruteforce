package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestLogger(t *testing.T) {
	var z zapcore.Level

	t.Run("NewLogger create", func(t *testing.T) {
		l, err := NewLogger(z, "prod", "/dev/null")
		require.NoError(t, err)
		require.NotNil(t, l)
	})
}
