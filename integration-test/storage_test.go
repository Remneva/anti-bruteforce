package test

import (
	"context"
	"flag"
	"log"
	"testing"

	"github.com/Remneva/anti-bruteforce/internal/logger"
	"github.com/Remneva/anti-bruteforce/internal/storage"
	"github.com/Remneva/anti-bruteforce/internal/storage/sql"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

var dsn = "host=postgres port=5432 user=test password=test dbname=exampledb sslmode=disable"

func TestStorage(t *testing.T) {

	var z zapcore.Level
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logg, err := logger.NewLogger(z, "dev", "/dev/null")
	if err != nil {
		log.Fatal("failed to create logger")
	}
	store := sql.NewDB(logg)
	if err := store.Connect(ctx, dsn, logg); err != nil {
		logg.Fatal("failed connection")
	}
	var ip storage.IP

	t.Run("Get from White List", func(t *testing.T) {
		ip.IP = "192.1.1.0/25"
		err := store.AddToWhiteList(ctx, ip)
		require.NoError(t, err)

		ok, err := store.GetFromWhiteList(ip)
		require.Equal(t, ok, true)
		require.NoError(t, err)
	})

	t.Run("Get from Black List", func(t *testing.T) {
		ip.IP = "192.1.1.0/26"

		err = store.AddToBlackList(ctx, ip)
		require.NoError(t, err)

		ok, err := store.GetFromBlackList(ip)
		require.NoError(t, err)
		require.Equal(t, ok, true)

	})

	t.Run("Delete from Black List", func(t *testing.T) {
		ip.IP = "192.1.1.0/26"

		err = store.DeleteFromBlackList(ctx, ip)
		require.NoError(t, err)

		ok, err := store.GetFromBlackList(ip)
		require.Equal(t, ok, false)
		require.NoError(t, err)
	})

	t.Run("Delete from White List", func(t *testing.T) {
		ip.IP = "192.1.1.0/25"

		err = store.DeleteFromWhiteList(ctx, ip)
		require.NoError(t, err)

		ok, err := store.GetFromWhiteList(ip)
		require.Equal(t, ok, false)
		require.NoError(t, err)
	})
}
