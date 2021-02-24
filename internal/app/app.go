package app

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/Remneva/anti-bruteforce/configs"
	"github.com/Remneva/anti-bruteforce/internal/redis"
	storage "github.com/Remneva/anti-bruteforce/internal/storage"

	// Postgres driver.
	_ "github.com/jackc/pgx/v4/stdlib"
	"go.uber.org/zap"
)

var _ InterfaceApp = (*App)(nil)

type App struct {
	rdb           redis.InterfaceRedis
	l             *zap.Logger
	listRepo      storage.Lists
	configRepo    storage.Configurations
	loginLimit    int64
	passwordLimit int64
	ipLimit       int64
	config        configs.Config
	mu            sync.Mutex
}

func NewApp(ctx context.Context, db storage.BaseStorage, c configs.Config, rdb *redis.Client, l *zap.Logger) (*App, error) { // nolint:interfacer
	limits, err := db.Configs().Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting configuration error: %w", err)
	}
	a := &App{
		config:        c,
		rdb:           rdb,
		l:             l,
		listRepo:      db.List(),
		configRepo:    db.Configs(),
		ipLimit:       limits["ipAttempts"],
		loginLimit:    limits["loginAttempts"],
		passwordLimit: limits["passwordAttempts"],
		mu:            sync.Mutex{},
	}
	return a, nil
}

func (a *App) Validate(ctx context.Context, request storage.Auth) (bool, error) {
	var isValidIP, isValidLogin, isValidPassword, white, black bool
	ip := storage.IP{IP: request.IP}

	white = a.containsInWhiteList(ctx, request.IP)
	black = a.containsInBlackList(ctx, request.IP)

	switch {
	case white:
		return true, nil
	case black:
		return false, nil
	case black && white:
		return false, nil
	}

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()
		if isValidIP = a.ipValidation(ctx, ip); !isValidIP {
			a.l.Info("Anti-Fraud Protection", zap.String("ip", request.IP))
		}
	}()

	go func() {
		defer wg.Done()
		if isValidLogin = a.loginValidation(ctx, ip, request.Login); !isValidLogin {
			a.l.Info("Anti-Fraud Protection", zap.String("login", request.Login))
		}
	}()

	go func() {
		defer wg.Done()
		if isValidPassword = a.passwordValidation(ctx, ip, request.Password); !isValidPassword {
			a.l.Info("Anti-Fraud Protection", zap.String("password", request.Password))
		}
	}()
	wg.Wait()

	if !isValidIP || !isValidLogin || !isValidPassword {
		return false, nil
	}
	a.l.Info("Successful authorization", zap.String("login", request.Login))
	return true, nil
}

func (a *App) CleanBucket(ctx context.Context, u storage.User) error {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := a.rdb.CleanByKey(ctx, u.IP); err != nil {
			a.l.Error("clean bucket by IP", zap.Error(err))
		}
	}()

	go func() {
		defer wg.Done()
		if err := a.rdb.CleanByKey(ctx, u.Login); err != nil {
			a.l.Error("clean bucket by Login", zap.Error(err))
		}
	}()
	wg.Wait()
	return nil
}

func (a *App) AddToWhiteList(ctx context.Context, ip storage.IP) error {
	a.l.Info("ip", zap.String("ip", ip.IP))
	if err := a.listRepo.AddToWhiteList(ctx, ip); err != nil {
		a.l.Error("add to white list error", zap.Error(err))
		return fmt.Errorf("add to white list error: %w", err)
	}
	return nil
}

func (a *App) AddToBlackList(ctx context.Context, ip storage.IP) error {
	if err := a.listRepo.AddToBlackList(ctx, ip); err != nil {
		a.l.Error("add to black list error", zap.Error(err))
		return fmt.Errorf("add to black list error: %w", err)
	}
	return nil
}

func (a *App) DeleteFromWhiteList(ctx context.Context, ip storage.IP) error {
	if err := a.listRepo.DeleteFromWhiteList(ctx, ip); err != nil {
		a.l.Error("delete from white list error", zap.Error(err))
		return fmt.Errorf("delete from white list error: %w", err)
	}
	return nil
}

func (a *App) DeleteFromBlackList(ctx context.Context, ip storage.IP) error {
	if err := a.listRepo.DeleteFromBlackList(ctx, ip); err != nil {
		a.l.Error("delete from black list error", zap.Error(err))
		return fmt.Errorf("delete from black list error: %w", err)
	}
	return nil
}

func (a *App) GetFromBlackList(ip storage.IP) bool {
	black, err := a.listRepo.GetFromBlackList(ip)
	if err != nil {
		a.l.Error("error while checking black list", zap.Error(err))
	}
	if black {
		a.l.Info("User in black list", zap.String("ip", ip.IP))
	}
	return black
}

func (a *App) GetFromWhiteList(ip storage.IP) bool {
	white, err := a.listRepo.GetFromWhiteList(ip)
	if err != nil {
		a.l.Error("error while checking white list", zap.Error(err))
	}
	if white {
		a.l.Info("User in white list", zap.String("ip", ip.IP))
	}
	return white
}

func (a *App) ipValidation(ctx context.Context, ip storage.IP) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	count, err := a.rdb.GettingCount(ctx, ip.IP)
	if err != nil {
		a.l.Error("getting count error", zap.Error(err))
		return false
	}
	if count > a.ipLimit {
		if err := a.listRepo.AddToBlackList(ctx, ip); err != nil {
			a.l.Error("adding to black list error", zap.Error(err))
			return false
		}
		return false
	}
	return true
}

func (a *App) loginValidation(ctx context.Context, ip storage.IP, login string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	count, err := a.rdb.GettingCount(ctx, login)
	if err != nil {
		a.l.Error("getting count error", zap.Error(err))
		return false
	}
	if count > a.loginLimit {
		if err := a.listRepo.AddToBlackList(ctx, ip); err != nil {
			a.l.Error("adding to black list error", zap.Error(err))
			return false
		}
		return false
	}
	return true
}

func (a *App) passwordValidation(ctx context.Context, ip storage.IP, pass string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	count, err := a.rdb.GettingCount(ctx, pass)
	if err != nil {
		a.l.Error("getting count error", zap.Error(err))
		return false
	}
	if count > a.loginLimit {
		if err := a.listRepo.AddToBlackList(ctx, ip); err != nil {
			a.l.Error("adding to black list error", zap.Error(err))
			return false
		}
		return false
	}
	return true
}

func parseAddress(ip string) (net.IPNet, error) {
	_, ipnet, err := net.ParseCIDR(ip)
	if err != nil {
		return *ipnet, fmt.Errorf("parse address error: %w", err)
	}
	return *ipnet, nil
}

func (a *App) containsInWhiteList(ctx context.Context, ip string) bool {
	ipnet, _ := parseAddress(ip)
	list, err := a.listRepo.GetAllFromWhiteList(ctx)
	if err != nil {
		a.l.Error("get list error", zap.Error(err))
		return false
	}
	for _, pr := range list {
		p := net.ParseIP(pr)
		result := ipnet.Contains(p)
		if result {
			return true
		}
	}
	return false
}
func (a *App) containsInBlackList(ctx context.Context, ip string) bool {
	ipnet, _ := parseAddress(ip)
	list, err := a.listRepo.GetAllFromBlackList(ctx)
	if err != nil {
		a.l.Error("get list error", zap.Error(err))
		return false
	}
	for _, pr := range list {
		p := net.ParseIP(pr)
		result := ipnet.Contains(p)
		if result {
			return true
		}
	}
	return false
}
