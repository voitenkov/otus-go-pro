package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/app"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/config"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/logger"
	queue "github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/queue/init"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/scheduler"
	storage "github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/storage/init"
)

var (
	configFile string
	wg         *sync.WaitGroup
)

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/scheduler_config.yaml", "Path to configuration file")
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

	if cfg.DB.Type != "sql" {
		log.Fatal("unsupported db type selected, sql is to be used")
	}

	storage.Connect()
	queue, err := queue.New(cfg)
	if err != nil {
		storage.Close()
		log.Fatal(err)
	}
	defer storage.Close()

	calendar := app.New(storage)
	scheduler := scheduler.New(logg, calendar, queue, cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()
		fmt.Println(ctx.Err())
		os.Exit(1)
	}()

	logg.Info("calendar scheduler is running...")
	wg = &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := scheduler.Start(ctx); err != nil {
			logg.Error("failed to start scheduler: " + err.Error())
		}
	}()

	wg.Wait()
	cancel()
	os.Exit(1) //nolint:gocritic
}
