package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/app"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/config"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/logger"
	internalgrpc "github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/server/grpc"
	internalhttp "github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/server/http"
	storage "github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/storage/init"
)

var (
	configFile string
	wg         *sync.WaitGroup
)

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	cfg, err := config.Parse(configFile)
	if err != nil {
		log.Fatal(err)
	}

	logg := logger.New(cfg.Logger.Level)

	storage, err := storage.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	storage.Connect()
	defer storage.Close()

	calendar := app.New(storage)

	server := internalhttp.NewServer(logg, calendar, cfg)
	GRPCServer := internalgrpc.NewGRPCServer(logg, calendar, cfg)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}

		GRPCServer.Stop()
	}()

	logg.Info("calendar is running...")
	wg = &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.Start(ctx); err != nil {
			logg.Error("failed to start http server: " + err.Error())
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := GRPCServer.Start(ctx); err != nil {
			logg.Error("failed to start grpc server: " + err.Error())
		}
	}()

	wg.Wait()
	cancel()
	os.Exit(1) //nolint:gocritic
}
