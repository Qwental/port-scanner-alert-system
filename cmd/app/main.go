package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Qwental/port-scanner-alert-system/internal/config"
	"github.com/Qwental/port-scanner-alert-system/internal/diff"
	"github.com/Qwental/port-scanner-alert-system/internal/logger"
	"github.com/Qwental/port-scanner-alert-system/internal/notifier"
	"github.com/Qwental/port-scanner-alert-system/internal/report"
	"github.com/Qwental/port-scanner-alert-system/internal/scanner"
	"github.com/Qwental/port-scanner-alert-system/internal/scheduler"
	"github.com/Qwental/port-scanner-alert-system/internal/storage/sqlite"
)

func main() {
	log := logger.InitLogger()
	defer log.Sync()

	if os.Geteuid() != 0 {
		log.Error("This application must be run as ROOT (sudo)!")
		log.Sync()
		os.Exit(1)
	}

	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "config/config.yaml"
	}

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Errorf("Failed to load config: %v", err)
		log.Sync()
		os.Exit(1)
	}

	storage, err := sqlite.NewStorage(cfg.Database.Path, log)
	if err != nil {
		log.Errorf("Failed to init database: %v", err)
		log.Sync()
		os.Exit(1)
	}
	defer storage.Close()

	scn := scanner.NewMasscanWrapper(cfg.Masscan, log)

	var tg *notifier.TelegramNotifier
	if cfg.Telegram.Enabled && cfg.Telegram.Token != "" {
		tg = notifier.NewTelegramNotifier(cfg.Telegram.Token, cfg.Telegram.ChatID, log)
		log.Info("Telegram notifier enabled")
	}

	var em *notifier.EmailNotifier
	if cfg.SMTP.Enabled && cfg.SMTP.User != "" {
		em = notifier.NewEmailNotifier(
			cfg.SMTP.Host, cfg.SMTP.Port,
			cfg.SMTP.User, cfg.SMTP.Password,
			cfg.SMTP.From, log,
		)
		log.Info("Email notifier enabled")
	}

	task := func(ctx context.Context) error {
		previous, err := storage.GetAll()
		if err != nil {
			return fmt.Errorf("load previous state: %w", err)
		}

		results, err := scn.Run(ctx, cfg.Targets)
		if err != nil {
			return fmt.Errorf("scan: %w", err)
		}

		d := diff.Compare(results, previous)
		log.Infof("Results: %d ports | Diff: %d new, %d changed, %d closed",
			len(results), len(d.New), len(d.Changed), len(d.Closed))

		if len(results) > 0 {
			if err := storage.Upsert(results); err != nil {
				log.Errorf("Save failed: %v", err)
			}
		}

		fmt.Println(report.BuildScanReport(results))
		diffText := report.BuildDiffReport(d)
		fmt.Println(diffText)

		hasChanges := len(d.New) > 0 || len(d.Changed) > 0 || len(d.Closed) > 0
		if !hasChanges {
			log.Info("No changes, skipping notifications")
			return nil
		}

		if tg != nil {
			if err := tg.Send(diffText); err != nil {
				log.Errorf("Telegram failed: %v", err)
			}
		}

		if em != nil {
			htmlBody := report.BuildDiffHTML(d)
			if err := em.Send(cfg.SMTP.To, "Port Scanner Alert", htmlBody); err != nil {
				log.Errorf("Email failed: %v", err)
			}
		}

		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Info("Shutting down...")
		cancel()
	}()

	if cfg.Scheduler.Enabled {
		interval, err := time.ParseDuration(cfg.Scheduler.Interval)
		if err != nil {
			log.Errorf("Invalid interval %q: %v", cfg.Scheduler.Interval, err)
			log.Sync()
			os.Exit(1)
		}
		log.Infof("Scheduler mode: every %s", interval)
		s := scheduler.New(interval, task, log)
		s.Run(ctx)
	} else {
		log.Info("Single scan mode")
		if err := task(ctx); err != nil {
			log.Errorf("Task failed: %v", err)
		}
	}

	log.Info("Goodbye!")
}