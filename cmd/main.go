package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Remneva/anti-bruteforce/configs"
	"github.com/Remneva/anti-bruteforce/internal/app"
	"github.com/Remneva/anti-bruteforce/internal/cli"
	"github.com/Remneva/anti-bruteforce/internal/logger"
	"github.com/Remneva/anti-bruteforce/internal/redis"
	"github.com/Remneva/anti-bruteforce/internal/server"
	grpc2 "github.com/Remneva/anti-bruteforce/internal/server/grpc"
	"github.com/Remneva/anti-bruteforce/internal/storage/sql"
)

var config string
var env string

func init() {
	flag.StringVar(&config, "config", "./configs/config.toml", "Path to configuration file")
	flag.StringVar(&env, "env", "prod", "environmental")
}

func main() {
	flag.Parse()
	config, err := configs.Read(config)
	if err != nil {
		log.Fatal("failed to read config")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logg, err := logger.NewLogger(config.Logger.Level, env, config.Logger.Path)
	if err != nil {
		logg.Error("failed to create logger")
	}
	redisClient := redis.NewClient(logg, config.Redis.ExpiryPeriod)
	redisClient, err = redisClient.RdbConnect(ctx, config.Redis.Address, config.Redis.Password)
	if err != nil {
		logg.Error("failed to get redis connection")
	}
	storage := sql.NewDB(logg)
	if err := storage.Connect(ctx, config.PSQL.DSN, logg); err != nil {
		logg.Fatal("failed connection")
	}
	defer storage.Close()
	application := app.NewApp(ctx, storage, config, redisClient, logg)

	grpc, _ := grpc2.NewServer(application, logg, config.Port.Grpc)
	client := cli.New(application)
	go signalChan(grpc, client)
	go client.RunCli()
	if err := grpc.Start(); err != nil {
		logg.Fatal("failed to start grpc")
	}
}
func signalChan(srv ...server.Stopper) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	fmt.Printf("Got %v...\n", <-signals)

	for _, s := range srv {
		s.Stop()
	}
}
