package main

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

func TestEnvAndParseHelpers(t *testing.T) {
	t.Setenv("X_TEST_ENV", "value")
	if got := envOrDefault("X_TEST_ENV", "fallback"); got != "value" {
		t.Fatalf("unexpected env value: %s", got)
	}
	if got := envOrDefault("X_TEST_ENV_MISSING", "fallback"); got != "fallback" {
		t.Fatalf("unexpected fallback value: %s", got)
	}

	if mustParseDuration("D", "1s") != time.Second {
		t.Fatalf("unexpected duration parse")
	}
	if mustParsePositiveInt("N", "3") != 3 {
		t.Fatalf("unexpected int parse")
	}
	if !mustParseBool("B", "yes") {
		t.Fatalf("unexpected bool parse")
	}
}

func TestParseHelpersPanicOnInvalid(t *testing.T) {
	assertPanics(t, func() { mustParseDuration("D", "bad") })
	assertPanics(t, func() { mustParsePositiveInt("N", "0") })
	assertPanics(t, func() { mustParseBool("B", "unknown") })
}

func TestGeneratorWindow(t *testing.T) {
	from, to := generatorWindow(time.Date(2026, 5, 4, 18, 20, 0, 0, time.UTC), 30)
	if from.Format(time.DateOnly) != "2026-05-04" || to.Format(time.DateOnly) != "2026-06-03" {
		t.Fatalf("unexpected window: %s..%s", from.Format(time.DateOnly), to.Format(time.DateOnly))
	}
}

func TestRunScheduleGenerator(t *testing.T) {
	mock := &scheduleUsecaseMainMock{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	cfg := config{
		ScheduleGeneratorEnabled:    true,
		ScheduleGeneratorInterval:   5 * time.Millisecond,
		ScheduleGeneratorTimeout:    100 * time.Millisecond,
		ScheduleGeneratorWindowDays: 5,
	}

	done := make(chan struct{})
	go func() {
		runScheduleGenerator(ctx, logger, mock, cfg)
		close(done)
	}()

	time.Sleep(15 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("generator did not stop")
	}

	if mock.calls == 0 {
		t.Fatalf("expected generator calls")
	}
}

func TestLoadConfig(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":9999")
	t.Setenv("DATABASE_DSN", "postgres://a:b@localhost:5432/db?sslmode=disable")
	t.Setenv("SCHEDULE_GENERATOR_ENABLED", "false")
	t.Setenv("SCHEDULE_GENERATOR_INTERVAL", "2h")
	t.Setenv("SCHEDULE_GENERATOR_TIMEOUT", "3s")
	t.Setenv("SCHEDULE_GENERATOR_WINDOW_DAYS", "12")

	cfg := loadConfig()
	if cfg.HTTPAddr != ":9999" || cfg.ScheduleGeneratorEnabled {
		t.Fatalf("unexpected config %+v", cfg)
	}
	if cfg.ScheduleGeneratorInterval != 2*time.Hour || cfg.ScheduleGeneratorTimeout != 3*time.Second || cfg.ScheduleGeneratorWindowDays != 12 {
		t.Fatalf("unexpected generator config %+v", cfg)
	}
}

func TestLoadConfigPanicsOnInvalidGeneratorEnv(t *testing.T) {
	t.Setenv("SCHEDULE_GENERATOR_ENABLED", "not-bool")
	assertPanics(t, func() { loadConfig() })
}

func TestLoadConfigPanicsOnBlankDatabaseDSN(t *testing.T) {
	t.Setenv("DATABASE_DSN", "   ")
	assertPanics(t, func() { loadConfig() })
}

type scheduleUsecaseMainMock struct {
	calls int
}

func (m *scheduleUsecaseMainMock) CreateSchedule(context.Context, taskusecase.CreateScheduleInput) (*taskdomain.Schedule, error) {
	return nil, nil
}
func (m *scheduleUsecaseMainMock) UpdateSchedule(context.Context, int64, taskusecase.UpdateScheduleInput) (*taskdomain.Schedule, error) {
	return nil, nil
}
func (m *scheduleUsecaseMainMock) DeactivateSchedule(context.Context, int64) error { return nil }
func (m *scheduleUsecaseMainMock) GetScheduleByID(context.Context, int64) (*taskdomain.Schedule, error) {
	return nil, nil
}
func (m *scheduleUsecaseMainMock) ListSchedules(context.Context) ([]taskdomain.Schedule, error) {
	return nil, nil
}
func (m *scheduleUsecaseMainMock) ComputePlannedDates(context.Context, int64, time.Time, time.Time) ([]time.Time, error) {
	return nil, nil
}
func (m *scheduleUsecaseMainMock) GenerateTasks(context.Context, time.Time, time.Time) (int, error) {
	m.calls++
	return 1, nil
}

func assertPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic")
		}
	}()
	fn()
}

func TestMain(_ *testing.T) {
	_ = os.Getenv("NOOP")
}
