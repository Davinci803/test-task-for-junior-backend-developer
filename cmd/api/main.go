package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	infrastructurepostgres "example.com/taskservice/internal/infrastructure/postgres"
	postgresrepo "example.com/taskservice/internal/repository/postgres"
	transporthttp "example.com/taskservice/internal/transport/http"
	swaggerdocs "example.com/taskservice/internal/transport/http/docs"
	httphandlers "example.com/taskservice/internal/transport/http/handlers"
	"example.com/taskservice/internal/usecase/task"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := loadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := infrastructurepostgres.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		logger.Error("open postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	taskRepo := postgresrepo.New(pool)
	scheduleRepo := postgresrepo.NewScheduleRepository(pool)
	taskUsecase := task.NewService(taskRepo)
	scheduleUsecase := task.NewScheduleService(scheduleRepo, taskRepo)
	taskHandler := httphandlers.NewTaskHandler(taskUsecase)
	scheduleHandler := httphandlers.NewScheduleHandler(scheduleUsecase)
	docsHandler := swaggerdocs.NewHandler()
	router := transporthttp.NewRouter(taskHandler, scheduleHandler, docsHandler)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if cfg.ScheduleGeneratorEnabled {
		go runScheduleGenerator(ctx, logger, scheduleUsecase, cfg)
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown http server", "error", err)
		}
	}()

	logger.Info("http server started", "addr", cfg.HTTPAddr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("listen and serve", "error", err)
		os.Exit(1)
	}
}

type config struct {
	HTTPAddr                    string
	DatabaseDSN                 string
	ScheduleGeneratorEnabled    bool
	ScheduleGeneratorInterval   time.Duration
	ScheduleGeneratorTimeout    time.Duration
	ScheduleGeneratorWindowDays int
}

func loadConfig() config {
	cfg := config{
		HTTPAddr:                  envOrDefault("HTTP_ADDR", ":8080"),
		DatabaseDSN:               envOrDefault("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/taskservice?sslmode=disable"),
		ScheduleGeneratorEnabled:  mustParseBool("SCHEDULE_GENERATOR_ENABLED", envOrDefault("SCHEDULE_GENERATOR_ENABLED", "true")),
		ScheduleGeneratorInterval: mustParseDuration("SCHEDULE_GENERATOR_INTERVAL", envOrDefault("SCHEDULE_GENERATOR_INTERVAL", "1h")),
		ScheduleGeneratorTimeout:  mustParseDuration("SCHEDULE_GENERATOR_TIMEOUT", envOrDefault("SCHEDULE_GENERATOR_TIMEOUT", "15s")),
		ScheduleGeneratorWindowDays: mustParsePositiveInt(
			"SCHEDULE_GENERATOR_WINDOW_DAYS",
			envOrDefault("SCHEDULE_GENERATOR_WINDOW_DAYS", "30"),
		),
	}

	if strings.TrimSpace(cfg.DatabaseDSN) == "" {
		panic(fmt.Errorf("DATABASE_DSN is required"))
	}

	return cfg
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func runScheduleGenerator(ctx context.Context, logger *slog.Logger, usecase task.ScheduleUsecase, cfg config) {
	run := func(trigger string) {
		from, to := generatorWindow(time.Now().UTC(), cfg.ScheduleGeneratorWindowDays)

		cycleCtx, cancel := context.WithTimeout(ctx, cfg.ScheduleGeneratorTimeout)
		defer cancel()

		created, err := usecase.GenerateTasks(cycleCtx, from, to)
		if err != nil {
			logger.Error("schedule generator cycle failed",
				"trigger", trigger,
				"from", from.Format(time.DateOnly),
				"to", to.Format(time.DateOnly),
				"error", err,
			)
			return
		}

		logger.Info("schedule generator cycle completed",
			"trigger", trigger,
			"from", from.Format(time.DateOnly),
			"to", to.Format(time.DateOnly),
			"created_tasks", created,
		)
	}

	run("startup")

	ticker := time.NewTicker(cfg.ScheduleGeneratorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("schedule generator stopped")
			return
		case <-ticker.C:
			run("ticker")
		}
	}
}

func generatorWindow(now time.Time, windowDays int) (time.Time, time.Time) {
	utc := now.UTC()
	from := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, windowDays)
	return from, to
}

func mustParseDuration(name, value string) time.Duration {
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		panic(fmt.Errorf("%s must be a positive duration", name))
	}
	return duration
}

func mustParsePositiveInt(name, value string) int {
	number, err := strconv.Atoi(value)
	if err != nil || number <= 0 {
		panic(fmt.Errorf("%s must be a positive integer", name))
	}
	return number
}

func mustParseBool(name, value string) bool {
	normalized := strings.TrimSpace(strings.ToLower(value))
	switch normalized {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		panic(fmt.Errorf("%s must be a boolean", name))
	}
}
