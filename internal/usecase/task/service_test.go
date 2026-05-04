package task

import (
	"context"
	"errors"
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func TestServiceCreate(t *testing.T) {
	repo := &taskRepoStub{}
	svc := NewService(repo)
	svc.now = func() time.Time { return time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC) }

	created, err := svc.Create(context.Background(), CreateInput{
		Title:       " Test ",
		Description: " Desc ",
		Status:      "",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if created.Title != "Test" || created.Description != "Desc" {
		t.Fatalf("unexpected trimming result: %+v", created)
	}
	if created.Status != taskdomain.StatusNew {
		t.Fatalf("unexpected default status: %s", created.Status)
	}
}

func TestServiceGetByIDValidation(t *testing.T) {
	repo := &taskRepoStub{}
	svc := NewService(repo)

	_, err := svc.GetByID(context.Background(), 0)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestServiceUpdateAndDeleteValidation(t *testing.T) {
	repo := &taskRepoStub{}
	svc := NewService(repo)

	_, err := svc.Update(context.Background(), -1, UpdateInput{Title: "x", Status: taskdomain.StatusNew})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput from update, got: %v", err)
	}

	err = svc.Delete(context.Background(), 0)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput from delete, got: %v", err)
	}
}

func TestServiceUpdateAndListSuccess(t *testing.T) {
	repo := &taskRepoStub{}
	svc := NewService(repo)
	svc.now = func() time.Time { return time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC) }

	updated, err := svc.Update(context.Background(), 1, UpdateInput{
		Title:       " Updated ",
		Description: " Desc ",
		Status:      taskdomain.StatusDone,
	})
	if err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}
	if updated.Title != "Updated" || updated.Status != taskdomain.StatusDone {
		t.Fatalf("unexpected update result: %+v", updated)
	}

	items, err := svc.List(context.Background(), ListFilter{})
	if err != nil || len(items) == 0 {
		t.Fatalf("unexpected list result: len=%d err=%v", len(items), err)
	}
}

func TestValidateCreateAndUpdateInput(t *testing.T) {
	_, err := validateCreateInput(CreateInput{Title: "", Status: taskdomain.StatusNew})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got: %v", err)
	}

	_, err = validateCreateInput(CreateInput{Title: "x", Status: taskdomain.Status("bad")})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for bad status, got: %v", err)
	}

	_, err = validateUpdateInput(UpdateInput{Title: "", Status: taskdomain.StatusDone})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for empty title, got: %v", err)
	}

	_, err = validateUpdateInput(UpdateInput{Title: "x", Status: taskdomain.Status("bad")})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for bad update status, got: %v", err)
	}
}

func TestServiceListFilterValidation(t *testing.T) {
	repo := &taskRepoStub{}
	svc := NewService(repo)

	scheduleID := int64(-10)
	_, err := svc.List(context.Background(), ListFilter{ScheduleID: &scheduleID})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got: %v", err)
	}

	from := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	_, err = svc.List(context.Background(), ListFilter{PlannedForFrom: &from, PlannedForTo: &to})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for invalid range, got: %v", err)
	}
}

type taskRepoStub struct{}

func (t *taskRepoStub) Create(_ context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	task.ID = 1
	return task, nil
}

func (t *taskRepoStub) GetByID(_ context.Context, id int64) (*taskdomain.Task, error) {
	return &taskdomain.Task{ID: id, Title: "x", Status: taskdomain.StatusNew}, nil
}

func (t *taskRepoStub) Update(_ context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	return task, nil
}

func (t *taskRepoStub) Delete(_ context.Context, _ int64) error {
	return nil
}

func (t *taskRepoStub) List(_ context.Context, _ ListFilter) ([]taskdomain.Task, error) {
	return []taskdomain.Task{{ID: 1, Title: "x", Status: taskdomain.StatusNew}}, nil
}
