package task

import (
	"context"
	"fmt"
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func TestScheduleServiceGenerateTasks_IdempotentAcrossRuns(t *testing.T) {
	schedule := taskdomain.Schedule{
		ID:             7,
		BaseTitle:      "Daily reminder",
		StatusTemplate: taskdomain.StatusNew,
		Type:           taskdomain.ScheduleTypeDaily,
		Payload: taskdomain.SchedulePayload{
			Daily: &taskdomain.DailySchedule{Interval: 1},
		},
		StartDate: time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC),
		IsActive:  true,
	}

	scheduleRepo := &fakeScheduleRepository{
		activeSchedules: []taskdomain.Schedule{schedule},
	}
	materializeRepo := newDedupMaterializationRepository()

	service := NewScheduleService(scheduleRepo, materializeRepo)
	from := time.Date(2026, time.May, 5, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.May, 7, 0, 0, 0, 0, time.UTC)

	firstRunCreated, err := service.GenerateTasks(context.Background(), from, to)
	if err != nil {
		t.Fatalf("first run returned error: %v", err)
	}
	if firstRunCreated != 3 {
		t.Fatalf("unexpected created tasks on first run: got=%d want=%d", firstRunCreated, 3)
	}

	secondRunCreated, err := service.GenerateTasks(context.Background(), from, to)
	if err != nil {
		t.Fatalf("second run returned error: %v", err)
	}
	if secondRunCreated != 0 {
		t.Fatalf("second run should be idempotent: got=%d want=0", secondRunCreated)
	}
}

type fakeScheduleRepository struct {
	activeSchedules []taskdomain.Schedule
}

func (f *fakeScheduleRepository) Create(context.Context, *taskdomain.Schedule) (*taskdomain.Schedule, error) {
	return nil, fmt.Errorf("not implemented in test")
}

func (f *fakeScheduleRepository) GetByID(context.Context, int64) (*taskdomain.Schedule, error) {
	return nil, fmt.Errorf("not implemented in test")
}

func (f *fakeScheduleRepository) Update(context.Context, *taskdomain.Schedule) (*taskdomain.Schedule, error) {
	return nil, fmt.Errorf("not implemented in test")
}

func (f *fakeScheduleRepository) Deactivate(context.Context, int64, time.Time) error {
	return fmt.Errorf("not implemented in test")
}

func (f *fakeScheduleRepository) List(context.Context) ([]taskdomain.Schedule, error) {
	return nil, fmt.Errorf("not implemented in test")
}

func (f *fakeScheduleRepository) ListActiveInRange(context.Context, time.Time, time.Time) ([]taskdomain.Schedule, error) {
	result := make([]taskdomain.Schedule, len(f.activeSchedules))
	copy(result, f.activeSchedules)
	return result, nil
}

type dedupMaterializationRepository struct {
	seen map[string]struct{}
}

func newDedupMaterializationRepository() *dedupMaterializationRepository {
	return &dedupMaterializationRepository{
		seen: make(map[string]struct{}),
	}
}

func (r *dedupMaterializationRepository) CreateBatch(_ context.Context, tasks []MaterializedTaskInput) (int, error) {
	created := 0
	for _, task := range tasks {
		key := fmt.Sprintf("%d:%s", task.ScheduleID, task.PlannedFor.UTC().Format(time.DateOnly))
		if _, exists := r.seen[key]; exists {
			continue
		}
		r.seen[key] = struct{}{}
		created++
	}
	return created, nil
}
