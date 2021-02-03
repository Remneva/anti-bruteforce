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

var isValidIP, isValidLogin, isValidPassword, white, black bool

type App struct {
	rdb           *redis.Client
	l             *zap.Logger
	listRepo      storage.ListStorage
	configRepo    storage.ConfigurationStorage
	loginLimit    int64
	passwordLimit int64
	ipLimit       int64
	config        configs.Config
	mu            sync.Mutex
}

func NewApp(ctx context.Context, db storage.BaseStorage, c configs.Config, rdb *redis.Client, l *zap.Logger) *App {
	limits, err := db.Configs().Get(ctx)
	if err != nil {
		l.Error("getting configuration error")
	}
	a := &App{
		config:        c,
		rdb:           rdb,
		l:             l,
		listRepo:      db.Lists(),
		configRepo:    db.Configs(),
		ipLimit:       limits["ipAttempts"],
		loginLimit:    limits["loginAttempts"],
		passwordLimit: limits["passwordAttempts"],
		mu:            sync.Mutex{},
	}
	return a
}

func (a *App) Validate(ctx context.Context, request storage.Auth) (bool, error) {
	ip := storage.IP{IP: request.IP}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		white = a.containsInWhiteList(ctx, request.IP)
	}()
	go func() {
		defer wg.Done()
		black = a.containsInBlackList(ctx, request.IP)
	}()
	wg.Wait()
	switch {
	case white:
		return true, nil
	case black:
		return false, nil
	case black && white:
		return false, nil
	}

	wg.Add(3)
	go func() {
		defer wg.Done()
		isValidIP = a.ipValidation(ctx, ip)
	}()
	go func() {
		defer wg.Done()
		isValidLogin = a.loginValidation(ctx, ip, request.Login)
	}()
	go func() {
		defer wg.Done()
		isValidPassword = a.passwordValidation(ctx, ip, request.Password)
	}()
	wg.Wait()
	if !isValidIP || !isValidLogin || !isValidPassword {
		a.l.Info("Anti-Fraud Protection", zap.String("login", request.Login))
		return false, nil
	}
	a.l.Info("Successful authorization", zap.String("login", request.Login))
	return true, nil
}

func (a *App) CleanBucket(u storage.User) error {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := a.rdb.CleanByKey(u.IP); err != nil {
			a.l.Error("clean bucket by IP", zap.Error(err))
		}
	}()
	go func() {
		defer wg.Done()
		if err := a.rdb.CleanByKey(u.Login); err != nil {
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
	count, err := a.rdb.GettingCount(ip.IP)
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
	count, err := a.rdb.GettingCount(login)
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
	count, err := a.rdb.GettingCount(pass)
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
	a.mu.Lock()
	defer a.mu.Unlock()
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
	a.mu.Lock()
	defer a.mu.Unlock()
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
