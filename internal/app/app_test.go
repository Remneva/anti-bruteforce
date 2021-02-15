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
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
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
	cancel     context.CancelFunc
	log        *zap.Logger
	rdb        *redis.Client
	app        *App
	mr         *miniredis.Miniredis
}

func (s *StoreSuite) TestValidationBlackList() {
	var ip storage.IP
	ip.IP = "192.1.1.0/25"
	list1 := []string{"192.1.1.0", "255.255.255.128"}
	list2 := []string{"0.0.0"}
	auth := &storage.Auth{
		Login: faker.Name(),
		IP:    ip.IP,
	}

	s.mockListDB.EXPECT().GetAllFromBlackList(s.ctx).Return(list1, nil).Times(1)
	s.mockListDB.EXPECT().GetAllFromWhiteList(s.ctx).Return(list2, nil).Times(1)

	result, err := s.app.Validate(s.ctx, *auth)
	s.Require().NoError(err)
	s.Require().False(result)
}

func (s *StoreSuite) TestValidationWhiteListOk() {
	var ip storage.IP
	ip.IP = "192.1.1.0/25"
	list2 := []string{"192.1.1.0", "255.255.255.128"}
	list1 := []string{"0.0.0"}
	auth := &storage.Auth{
		Login: faker.Name(),
		IP:    ip.IP,
	}

	s.mockListDB.EXPECT().GetAllFromBlackList(s.ctx).Return(list1, nil).Times(1)
	s.mockListDB.EXPECT().GetAllFromWhiteList(s.ctx).Return(list2, nil).Times(1)

	result, err := s.app.Validate(s.ctx, *auth)
	s.Require().NoError(err)
	s.Require().True(result)
}

func (s *StoreSuite) TestValidation() {
	var ip storage.IP
	var result bool
	var err error
	ip.IP = faker.IPv4() + "/25"
	list1 := []string{"0.0.0"}
	auth := &storage.Auth{
		Login: faker.Name(),
		IP:    ip.IP,
	}
	count := 3
	s.mockListDB.EXPECT().GetAllFromBlackList(s.ctx).Return(list1, nil).Times(3)
	s.mockListDB.EXPECT().GetAllFromWhiteList(s.ctx).Return(list1, nil).Times(3)
	s.mockListDB.EXPECT().AddToBlackList(gomock.Any(), ip).Return(nil).Times(3)

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
	ip.IP = faker.IPv4() + "/25"
	list1 := []string{"0.0.0"}
	auth := &storage.Auth{
		IP: ip.IP,
	}
	count := 3
	s.mockListDB.EXPECT().GetAllFromBlackList(s.ctx).Return(list1, nil).Times(3)
	s.mockListDB.EXPECT().GetAllFromWhiteList(s.ctx).Return(list1, nil).Times(3)
	s.mockListDB.EXPECT().AddToBlackList(gomock.Any(), ip).Return(nil).Times(3)

	for i := 0; i < count; i++ {
		auth.Password = faker.Password()
		auth.Login = faker.Name()
		result, err = s.app.Validate(s.ctx, *auth)
		s.Require().NoError(err)
	}
	s.Require().False(result)
}

func (s *StoreSuite) TestValidationPassword() {
	var result bool
	var err error
	list1 := []string{"0.0.0"}
	auth := &storage.Auth{
		Login:    faker.Name(),
		Password: faker.Password(),
	}
	count := 3
	s.mockListDB.EXPECT().GetAllFromBlackList(s.ctx).Return(list1, nil).Times(3)
	s.mockListDB.EXPECT().GetAllFromWhiteList(s.ctx).Return(list1, nil).Times(3)
	s.mockListDB.EXPECT().AddToBlackList(gomock.Any(), gomock.Any()).Return(nil).Times(3)

	for i := 0; i < count; i++ {
		auth.IP = faker.IPv4() + "/25"
		result, err = s.app.Validate(s.ctx, *auth)
		s.Require().NoError(err)
	}
	s.Require().False(result)
}

func (s *StoreSuite) TestValidationLogin() {
	var result bool
	var err error
	auth := &storage.Auth{
		IP:       "192.1.1.0/25",
		Password: faker.Password(),
	}
	list1 := []string{"0.0.0", "192.1.1.0"}
	list2 := []string{"0.0.0"}
	count := 3
	s.mockListDB.EXPECT().GetAllFromBlackList(s.ctx).Return(list1, nil).Times(3)
	s.mockListDB.EXPECT().GetAllFromWhiteList(s.ctx).Return(list2, nil).Times(3)
	s.mockListDB.EXPECT().AddToBlackList(gomock.Any(), gomock.Any()).Return(nil).Times(3)

	for i := 0; i < count; i++ {
		auth.Login = faker.Name()
		result, err = s.app.Validate(s.ctx, *auth)
		s.Require().NoError(err)
	}
	s.Require().False(result)
}

func (s *StoreSuite) TestValidationOk() {
	var ip storage.IP
	ip.IP = faker.IPv4() + "/25"
	auth := &storage.Auth{
		Login:    faker.Name(),
		IP:       ip.IP,
		Password: faker.Password(),
	}
	list1 := []string{"0.0.0"}

	s.mockListDB.EXPECT().GetAllFromBlackList(s.ctx).Return(list1, nil).Times(1)
	s.mockListDB.EXPECT().GetAllFromWhiteList(s.ctx).Return(list1, nil).Times(1)

	result, err := s.app.Validate(s.ctx, *auth)
	s.Require().NoError(err)
	s.Require().True(result)
}

func (s *StoreSuite) TestCleanBucket() {
	var ip storage.IP
	ip.IP = faker.IPv4() + "/25"
	user := storage.User{
		Login: faker.Name(),
		IP:    ip.IP,
	}

	s.mockListDB.EXPECT().GetFromBlackList(ip).Return(false, nil).Times(1)
	s.mockListDB.EXPECT().GetFromWhiteList(ip).Return(false, nil).Times(1)

	err := s.app.CleanBucket(s.ctx, user)
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

func (s *StoreSuite) TestContainsIpBlackListOk() {
	var ip storage.IP
	ip.IP = "192.1.1.0/25"
	list := []string{"192.1.1.0", "255.255.255.128"}

	s.mockListDB.EXPECT().GetAllFromBlackList(s.ctx).Return(list, nil).Times(1)

	result := s.app.containsInBlackList(s.ctx, ip.IP)
	s.Require().True(result)
}

func (s *StoreSuite) TestContainsIpBlackListNegative() {
	var ip storage.IP
	ip.IP = "192.1.1.0/25"
	list := []string{"1.1.1.0", "2.2.2.2"}

	s.mockListDB.EXPECT().GetAllFromBlackList(s.ctx).Return(list, nil).Times(1)

	result := s.app.containsInBlackList(s.ctx, ip.IP)
	s.Require().False(result)
}

func (s *StoreSuite) TestContainsIpWhiteListOk() {
	var ip storage.IP
	ip.IP = "192.1.1.0/25"
	list := []string{"192.1.1.0", "255.255.255.128"}

	s.mockListDB.EXPECT().GetAllFromWhiteList(s.ctx).Return(list, nil).Times(1)

	result := s.app.containsInWhiteList(s.ctx, ip.IP)
	s.Require().True(result)
}

func (s *StoreSuite) TestContainsIpWhiteListNegative() {
	var ip storage.IP
	ip.IP = "192.1.1.0/25"
	list := []string{"1.1.1.0", "2.2.2.2"}

	s.mockListDB.EXPECT().GetAllFromWhiteList(s.ctx).Return(list, nil).Times(1)

	result := s.app.containsInWhiteList(s.ctx, ip.IP)
	s.Require().False(result)
}

func (s *StoreSuite) TeardownTest() {
	goleak.VerifyNone(s.T())
	s.mockCtl.Finish()
	s.rdb.Close(s.ctx)
	s.mr.Close()
	s.cancel()
}

func (s *StoreSuite) SetupTest() {
	s.mockCtl = gomock.NewController(s.T())
	s.mockBaseDB = NewMockBaseStorage(s.mockCtl)
	s.mockListDB = NewMockListStorage(s.mockCtl)
	s.ctx, s.cancel = context.WithCancel(context.Background())
	mr, err := miniredis.Run()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	var z zapcore.Level
	logg, _ := logger.NewLogger(z, "dev", "/dev/null")
	rdb := redis.NewClient(logg, 2*time.Millisecond)
	rdbClient, _ := rdb.RdbConnect(s.ctx, mr.Addr(), "")
	s.mr = mr
	s.rdb = rdbClient
	s.app = &App{
		loginLimit:    2,
		passwordLimit: 2,
		ipLimit:       2,
		rdb:           rdb,
		l:             logg,
		listRepo:      s.mockListDB,
		configRepo:    NewMockConfigurationStorage(s.mockCtl),
	}
}
