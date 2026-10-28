// Command beacon serves the dashboard and the probe scheduler.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the service catalog loaded from YAML.
type Config struct {
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
	Budget   time.Duration `yaml:"budget"`
	Services []Service     `yaml:"services"`
}

// Service is one endpoint the scheduler checks.
type Service struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
	Tags []string `yaml:"tags"`
}

func main() {
	var (
		addr   = flag.String("addr", "127.0.0.1:8099", "listen address")
		dbPath = flag.String("db", "beacon.db", "SQLite database path")
		config = flag.String("config", "config/services.yaml", "service catalog")
		debug  = flag.Bool("debug", false, "log every probe")
	)
	flag.Parse()

	level := slog.LevelInfo
	if *debug {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	cfg, err := loadConfig(*config)
	if err != nil {
		logger.Error("config", "err", err)
		os.Exit(1)
	}

	store, err := OpenStore(*dbPath)
	if err != nil {
		logger.Error("store", "err", err)
		os.Exit(1)
	}
	defer store.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	prober := NewProber(cfg.Timeout, cfg.Budget, logger)
	sched := NewScheduler(cfg.Services, cfg.Interval, prober, store, logger)
	go sched.Run(ctx)

	server := NewServer(cfg, store, logger)
	srv := &http.Server{
		Addr:              *addr,
		Handler:           server,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdown); err != nil {
			logger.Warn("shutdown", "err", err)
		}
	}()

	logger.Info("listening", "addr", *addr, "services", len(cfg.Services))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("serve", "err", err)
		os.Exit(1)
	}
}

// loadConfig reads and validates the service catalog.
func loadConfig(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", path, err)
	}

	if cfg.Interval <= 0 {
		cfg.Interval = 30 * time.Second
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 4 * time.Second
	}
	if cfg.Budget <= 0 {
		cfg.Budget = 750 * time.Millisecond
	}
	if len(cfg.Services) == 0 {
		return Config{}, errors.New("service catalog is empty")
	}

	for i := range cfg.Services {
		if cfg.Services[i].URL == "" {
			return Config{}, fmt.Errorf("service %q has no URL", cfg.Services[i].Name)
		}
	}

	return cfg, nil
}
