package app

import (
	"context"

	"github.com/Remneva/anti-bruteforce/internal/storage"
)

type InterfaceApp interface {
	Validate(ctx context.Context, request storage.Auth) (bool, error)
	CleanBucket(ctx context.Context, u storage.User) error
	AddToWhiteList(ctx context.Context, ip storage.IP) error
	AddToBlackList(ctx context.Context, ip storage.IP) error
	DeleteFromWhiteList(ctx context.Context, ip storage.IP) error
	DeleteFromBlackList(ctx context.Context, ip storage.IP) error
	ipValidation(ctx context.Context, ip storage.IP) bool
	loginValidation(ctx context.Context, ip storage.IP, login string) bool
	passwordValidation(ctx context.Context, ip storage.IP, pass string) bool
}
