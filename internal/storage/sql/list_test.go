package sql

import (
	"context"
	"fmt"
	"testing"

	"github.com/Remneva/anti-bruteforce/internal/logger"
	"github.com/Remneva/anti-bruteforce/internal/storage"
	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestStoreSuite(t *testing.T) {
	suite.Run(t, new(StoreSuite))
}

type StoreSuite struct {
	suite.Suite
	st  *Storage
	ctx context.Context
	log *zap.Logger
}

func (s *StoreSuite) SetupTest() {
	s.ctx = context.Background()

	var z zapcore.Level
	logg, _ := logger.NewLogger(z, "dev", "/dev/null")
	store := NewDB(logg)
	if err := store.Connect(s.ctx, "host=localhost port=5432 user=mary password=mary dbname=exampledb sslmode=disable", logg); err != nil {
		logg.Fatal("failed connection")
	}
	s.st = store
}

func (s *StoreSuite) TestValidationBlackList() {
	var ip storage.IP
	ip.IP = faker.IPv4()

	list, _ := s.st.GetAllFromBlackList(s.ctx)
	fmt.Println(list)
}
