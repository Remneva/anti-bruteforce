package app

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/Remneva/anti-bruteforce/internal/logger"
	"github.com/Remneva/anti-bruteforce/internal/redis"
	"github.com/Remneva/anti-bruteforce/internal/storage"
	"github.com/alicebob/miniredis"
	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestStoreSuite(t *testing.T) {
	suite.Run(t, new(StoreSuite))
}

type StoreSuite struct {
	suite.Suite
	mockCtl    *gomock.Controller
	mockBaseDB *MockBaseStorage
	mockListDB *MockListStorage
	ctx        context.Context
	log        *zap.Logger
	rdb        *redis.Client
	app        *App
	mr         *miniredis.Miniredis
}

func (s *StoreSuite) TestValidationBlackList() {
	var ip storage.IP
	ip.IP = faker.IPv4()
	auth := &storage.Auth{
		Login: faker.Name(),
		IP:    ip.IP,
	}

	s.mockListDB.EXPECT().GetFromBlackList(ip).Return(true, nil)
	s.mockListDB.EXPECT().GetFromWhiteList(ip).Return(false, nil)

	result, err := s.app.Validate(s.ctx, *auth)
	s.Require().NoError(err)
	s.Require().False(result)
}

func (s *StoreSuite) TestValidationWitheList() {
	var ip storage.IP
	ip.IP = faker.IPv4()
	auth := &storage.Auth{
		Login: faker.Name(),
		IP:    ip.IP,
	}

	s.mockListDB.EXPECT().GetFromBlackList(ip).Return(false, nil)
	s.mockListDB.EXPECT().GetFromWhiteList(ip).Return(true, nil)

	result, err := s.app.Validate(s.ctx, *auth)
	s.Require().NoError(err)
	s.Require().True(result)
}

func (s *StoreSuite) TestValidation() {
	var ip storage.IP
	var result bool
	var err error
	ip.IP = faker.IPv4()
	auth := &storage.Auth{
		Login: faker.Name(),
		IP:    ip.IP,
	}
	count := 4
	s.mockListDB.EXPECT().GetFromBlackList(ip).Return(false, nil).Times(4)
	s.mockListDB.EXPECT().GetFromWhiteList(ip).Return(false, nil).Times(4)
	s.mockListDB.EXPECT().AddToBlackList(gomock.Any(), ip).Return(nil).Times(4)

	for i := 0; i < count; i++ {
		result, err = s.app.Validate(s.ctx, *auth)
		s.Require().NoError(err)
	}
	s.Require().False(result)
}

func (s *StoreSuite) TestValidationIP() {
	var ip storage.IP
	var result bool
	var err error
	ip.IP = faker.IPv4()
	auth := &storage.Auth{
		IP: ip.IP,
	}
	count := 4
	s.mockListDB.EXPECT().GetFromBlackList(ip).Return(false, nil).Times(4)
	s.mockListDB.EXPECT().GetFromWhiteList(ip).Return(false, nil).Times(4)
	s.mockListDB.EXPECT().AddToBlackList(gomock.Any(), ip).Return(nil).Times(4)

	for i := 0; i < count; i++ {
		auth.Password = faker.Password()
		auth.Login = faker.Name()
		result, err = s.app.Validate(s.ctx, *auth)
		s.Require().NoError(err)
	}
	s.Require().False(result)
}

func (s *StoreSuite) TestValidationPass() {
	var result bool
	var err error
	auth := &storage.Auth{
		Login:    faker.Name(),
		Password: faker.Password(),
	}
	count := 5
	s.mockListDB.EXPECT().GetFromBlackList(gomock.Any()).Return(false, nil).Times(3)
	s.mockListDB.EXPECT().GetFromBlackList(gomock.Any()).Return(true, nil).Times(2)
	s.mockListDB.EXPECT().GetFromWhiteList(gomock.Any()).Return(false, nil).Times(5)
	s.mockListDB.EXPECT().AddToBlackList(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	for i := 0; i < count; i++ {
		auth.IP = faker.IPv4()
		result, err = s.app.Validate(s.ctx, *auth)
		s.Require().NoError(err)
	}
	s.Require().False(result)
}

func (s *StoreSuite) TestValidationLogin() {
	var result bool
	var err error
	auth := &storage.Auth{
		IP:       faker.IPv4(),
		Password: faker.Password(),
	}
	count := 5
	s.mockListDB.EXPECT().GetFromBlackList(gomock.Any()).Return(false, nil).Times(3)
	s.mockListDB.EXPECT().GetFromBlackList(gomock.Any()).Return(true, nil).Times(2)
	s.mockListDB.EXPECT().GetFromWhiteList(gomock.Any()).Return(false, nil).Times(5)
	s.mockListDB.EXPECT().AddToBlackList(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	for i := 0; i < count; i++ {
		auth.Login = faker.Name()
		result, err = s.app.Validate(s.ctx, *auth)
		s.Require().NoError(err)
	}
	s.Require().False(result)
}

func (s *StoreSuite) TestValidationSuccess() {
	var ip storage.IP
	ip.IP = faker.IPv4()
	auth := &storage.Auth{
		Login:    faker.Name(),
		IP:       ip.IP,
		Password: faker.Password(),
	}

	s.mockListDB.EXPECT().GetFromBlackList(ip).Return(false, nil).Times(1)
	s.mockListDB.EXPECT().GetFromWhiteList(ip).Return(false, nil).Times(1)

	result, err := s.app.Validate(s.ctx, *auth)
	s.Require().NoError(err)
	s.Require().True(result)
}

func (s *StoreSuite) TestCleanBucket() {
	var ip storage.IP
	ip.IP = faker.IPv4()
	user := storage.User{
		Login: faker.Name(),
		IP:    ip.IP,
	}

	s.mockListDB.EXPECT().GetFromBlackList(ip).Return(false, nil).Times(1)
	s.mockListDB.EXPECT().GetFromWhiteList(ip).Return(false, nil).Times(1)

	err := s.app.CleanBucket(user)
	s.Require().NoError(err)
}

func (s *StoreSuite) TestAddToWhiteList() {
	var ip storage.IP
	ip.IP = faker.IPv4()

	s.mockListDB.EXPECT().AddToWhiteList(s.ctx, ip).Return(nil).Times(1)

	err := s.app.AddToWhiteList(s.ctx, ip)
	s.Require().NoError(err)
}

func (s *StoreSuite) TestAddToBlackList() {
	var ip storage.IP
	ip.IP = faker.IPv4()

	s.mockListDB.EXPECT().AddToBlackList(s.ctx, ip).Return(nil).Times(1)

	err := s.app.AddToBlackList(s.ctx, ip)
	s.Require().NoError(err)
}

func (s *StoreSuite) TeardownTest() {
	s.mockCtl.Finish()
	s.mr.Close()
}

func (s *StoreSuite) SetupTest() {
	s.mockCtl = gomock.NewController(s.T())
	s.mockBaseDB = NewMockBaseStorage(s.mockCtl)
	s.mockListDB = NewMockListStorage(s.mockCtl)
	s.ctx = context.Background()
	mr, err := miniredis.Run()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	address := mr.Addr()
	var z zapcore.Level
	logg, _ := logger.NewLogger(z, "dev", "/dev/null")
	rdb := redis.NewClient(logg, 25*time.Millisecond)
	rdbClient, _ := rdb.RdbConnect(s.ctx, address, "")
	s.rdb = rdbClient
	s.app = &App{
		loginLimit:    3,
		passwordLimit: 3,
		ipLimit:       3,
		rdb:           rdb,
		l:             logg,
		listRepo:      s.mockListDB,
		configRepo:    NewMockConfigurationStorage(s.mockCtl),
	}
}
