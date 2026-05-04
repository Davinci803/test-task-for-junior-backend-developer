package task

import (
	"context"
	"errors"
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func TestScheduleServiceCreateAndList(t *testing.T) {
	repo := &scheduleRepoStub{}
	svc := NewScheduleService(repo, &dedupMaterializationRepository{seen: map[string]struct{}{}})

	created, err := svc.CreateSchedule(context.Background(), CreateScheduleInput{
		BaseTitle:       "Rent",
		BaseDescription: "Monthly",
		StatusTemplate:  taskdomain.StatusNew,
		Type:            taskdomain.ScheduleTypeMonthlyDay,
		Payload: taskdomain.SchedulePayload{
			MonthlyDay: &taskdomain.MonthlyDaySchedule{Day: 15},
		},
		StartDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		IsActive:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected created ID")
	}

	items, err := svc.ListSchedules(context.Background())
	if err != nil || len(items) == 0 {
		t.Fatalf("unexpected list result: len=%d err=%v", len(items), err)
	}
}

func TestScheduleServiceErrorsAndCompute(t *testing.T) {
	repo := &scheduleRepoStub{
		getByIDErr: taskdomain.ErrScheduleNotFound,
	}
	svc := NewScheduleService(repo, &dedupMaterializationRepository{seen: map[string]struct{}{}})

	_, err := svc.GetScheduleByID(context.Background(), 10)
	if !errors.Is(err, ErrScheduleNotFound) {
		t.Fatalf("expected ErrScheduleNotFound, got: %v", err)
	}

	_, err = svc.UpdateSchedule(context.Background(), 0, UpdateScheduleInput{})
	if !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("expected ErrInvalidSchedule, got: %v", err)
	}

	repo.getByIDErr = nil
	repo.getByIDSchedule = &taskdomain.Schedule{
		ID:             11,
		BaseTitle:      "Odd",
		StatusTemplate: taskdomain.StatusNew,
		Type:           taskdomain.ScheduleTypeOddEvenDays,
		Payload: taskdomain.SchedulePayload{
			OddEvenDays: &taskdomain.OddEvenDaysSchedule{Mode: taskdomain.OddEvenModeOdd},
		},
		StartDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		IsActive:  true,
	}

	dates, err := svc.ComputePlannedDates(context.Background(), 11, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC))
	if err != nil || len(dates) == 0 {
		t.Fatalf("expected dates, got len=%d err=%v", len(dates), err)
	}

	_, err = svc.ComputePlannedDates(context.Background(), 0, time.Now(), time.Now())
	if !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("expected ErrInvalidSchedule for invalid id, got %v", err)
	}
}

func TestScheduleServiceUpdateDeactivateAndBuilders(t *testing.T) {
	repo := &scheduleRepoStub{}
	svc := NewScheduleService(repo, &dedupMaterializationRepository{seen: map[string]struct{}{}})

	updated, err := svc.UpdateSchedule(context.Background(), 1, UpdateScheduleInput{
		BaseTitle:       "Updated",
		BaseDescription: "Desc",
		StatusTemplate:  taskdomain.StatusInProgress,
		Type:            taskdomain.ScheduleTypeDaily,
		Payload: taskdomain.SchedulePayload{
			Daily: &taskdomain.DailySchedule{Interval: 1},
		},
		StartDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		IsActive:  true,
	})
	if err != nil || updated.ID != 1 {
		t.Fatalf("unexpected update result: %+v err=%v", updated, err)
	}

	if err := svc.DeactivateSchedule(context.Background(), 1); err != nil {
		t.Fatalf("unexpected deactivate error: %v", err)
	}

	repo.deactivateErr = taskdomain.ErrScheduleNotFound
	if err := svc.DeactivateSchedule(context.Background(), 1); !errors.Is(err, ErrScheduleNotFound) {
		t.Fatalf("expected ErrScheduleNotFound, got: %v", err)
	}

	_, err = buildScheduleFromCreateInput(CreateScheduleInput{
		BaseTitle:      "X",
		StatusTemplate: taskdomain.Status("bad"),
		Type:           taskdomain.ScheduleTypeDaily,
		Payload: taskdomain.SchedulePayload{
			Daily: &taskdomain.DailySchedule{Interval: 1},
		},
		StartDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		IsActive:  true,
	})
	if !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("expected ErrInvalidSchedule from create builder, got %v", err)
	}

	_, err = buildScheduleFromUpdateInput(UpdateScheduleInput{
		BaseTitle:      "X",
		StatusTemplate: taskdomain.StatusNew,
		Type:           taskdomain.ScheduleTypeDaily,
		Payload:        taskdomain.SchedulePayload{},
		StartDate:      time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		IsActive:       true,
	})
	if !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("expected ErrInvalidSchedule from update builder, got %v", err)
	}
}

type scheduleRepoStub struct {
	data            []taskdomain.Schedule
	getByIDSchedule *taskdomain.Schedule
	getByIDErr      error
	deactivateErr   error
}

func (s *scheduleRepoStub) Create(_ context.Context, schedule *taskdomain.Schedule) (*taskdomain.Schedule, error) {
	schedule.ID = int64(len(s.data) + 1)
	s.data = append(s.data, *schedule)
	return schedule, nil
}

func (s *scheduleRepoStub) GetByID(_ context.Context, id int64) (*taskdomain.Schedule, error) {
	if s.getByIDErr != nil {
		return nil, s.getByIDErr
	}
	if s.getByIDSchedule != nil {
		return s.getByIDSchedule, nil
	}
	for i := range s.data {
		if s.data[i].ID == id {
			return &s.data[i], nil
		}
	}
	return nil, taskdomain.ErrScheduleNotFound
}

func (s *scheduleRepoStub) Update(_ context.Context, schedule *taskdomain.Schedule) (*taskdomain.Schedule, error) {
	return schedule, nil
}

func (s *scheduleRepoStub) Deactivate(_ context.Context, _ int64, _ time.Time) error {
	return s.deactivateErr
}

func (s *scheduleRepoStub) List(_ context.Context) ([]taskdomain.Schedule, error) {
	result := make([]taskdomain.Schedule, len(s.data))
	copy(result, s.data)
	return result, nil
}

func (s *scheduleRepoStub) ListActiveInRange(_ context.Context, _, _ time.Time) ([]taskdomain.Schedule, error) {
	result := make([]taskdomain.Schedule, len(s.data))
	copy(result, s.data)
	return result, nil
}
