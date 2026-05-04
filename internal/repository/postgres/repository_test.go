package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	pgxmock "github.com/pashagolub/pgxmock/v4"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

func TestTaskRepositoryCRUDAndList(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("mock init error: %v", err)
	}
	defer mock.Close()

	repo := &Repository{pool: mock}
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery("INSERT INTO tasks").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "title", "description", "status", "schedule_id", "planned_for", "created_at", "updated_at"}).
				AddRow(int64(1), "t", "d", "new", nil, nil, now, now),
		)
	created, err := repo.Create(context.Background(), &taskdomain.Task{
		Title: "t", Description: "d", Status: taskdomain.StatusNew, CreatedAt: now, UpdatedAt: now,
	})
	if err != nil || created.ID != 1 {
		t.Fatalf("unexpected create result: %+v err=%v", created, err)
	}

	mock.ExpectQuery("SELECT id, title, description, status, schedule_id, planned_for, created_at, updated_at").
		WithArgs(int64(1)).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "title", "description", "status", "schedule_id", "planned_for", "created_at", "updated_at"}).
				AddRow(int64(1), "t", "d", "new", int64(7), now, now, now),
		)
	got, err := repo.GetByID(context.Background(), 1)
	if err != nil || got.ScheduleID == nil || *got.ScheduleID != 7 {
		t.Fatalf("unexpected get result: %+v err=%v", got, err)
	}

	mock.ExpectQuery("UPDATE tasks").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), int64(1)).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "title", "description", "status", "schedule_id", "planned_for", "created_at", "updated_at"}).
				AddRow(int64(1), "tu", "du", "done", nil, nil, now, now),
		)
	_, err = repo.Update(context.Background(), &taskdomain.Task{
		ID: 1, Title: "tu", Description: "du", Status: taskdomain.StatusDone, UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}

	mock.ExpectExec("DELETE FROM tasks").
		WithArgs(int64(1)).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))
	if err := repo.Delete(context.Background(), 1); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	mock.ExpectQuery("SELECT id, title, description, status, schedule_id, planned_for, created_at, updated_at").
		WithArgs().
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "title", "description", "status", "schedule_id", "planned_for", "created_at", "updated_at"}).
				AddRow(int64(1), "t", "d", "new", nil, nil, now, now),
		)
	items, err := repo.List(context.Background(), taskusecase.ListFilter{})
	if err != nil || len(items) != 1 {
		t.Fatalf("unexpected list result: len=%d err=%v", len(items), err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryConstructors(t *testing.T) {
	if New(nil) == nil {
		t.Fatalf("expected repository instance")
	}
	if NewScheduleRepository(nil) == nil {
		t.Fatalf("expected schedule repository instance")
	}
}

func TestTaskRepositoryNotFoundMapping(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	repo := &Repository{pool: mock}

	mock.ExpectQuery("SELECT id, title, description, status, schedule_id, planned_for, created_at, updated_at").
		WithArgs(int64(99)).
		WillReturnError(errors.New("no rows in result set"))
	_, err := repo.GetByID(context.Background(), 99)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestTaskRepositoryUpdateDeleteAndListBranches(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	repo := &Repository{pool: mock}
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery("UPDATE tasks").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), int64(1)).
		WillReturnError(pgx.ErrNoRows)
	_, err := repo.Update(context.Background(), &taskdomain.Task{
		ID: 1, Title: "u", Description: "d", Status: taskdomain.StatusDone, UpdatedAt: now,
	})
	if !errors.Is(err, taskdomain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	mock.ExpectExec("DELETE FROM tasks").
		WithArgs(int64(2)).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))
	err = repo.Delete(context.Background(), 2)
	if !errors.Is(err, taskdomain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound on delete, got %v", err)
	}

	sid := int64(5)
	mock.ExpectQuery("SELECT id, title, description, status, schedule_id, planned_for, created_at, updated_at").
		WithArgs(now, now, sid).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "title", "description", "status", "schedule_id", "planned_for", "created_at", "updated_at"}).
				AddRow(int64(1), "t", "d", "new", sid, now, now, now),
		)
	_, err = repo.List(context.Background(), taskusecase.ListFilter{
		PlannedForFrom: &now,
		PlannedForTo:   &now,
		ScheduleID:     &sid,
	})
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}
}

func TestScheduleRepositoryAndCodec(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("mock init error: %v", err)
	}
	defer mock.Close()
	repo := &ScheduleRepository{pool: mock}
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	payload, err := encodeSchedulePayload(taskdomain.ScheduleTypeDaily, taskdomain.SchedulePayload{
		Daily: &taskdomain.DailySchedule{Interval: 2},
	})
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	if _, err := decodeSchedulePayload(taskdomain.ScheduleTypeDaily, payload); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if _, err := encodeSchedulePayload(taskdomain.ScheduleTypeMonthlyDay, taskdomain.SchedulePayload{
		MonthlyDay: &taskdomain.MonthlyDaySchedule{Day: 10},
	}); err != nil {
		t.Fatalf("encode monthly error: %v", err)
	}
	if _, err := encodeSchedulePayload(taskdomain.ScheduleTypeSpecificDates, taskdomain.SchedulePayload{
		SpecificDates: &taskdomain.SpecificDatesSchedule{Dates: []time.Time{now}},
	}); err != nil {
		t.Fatalf("encode specific dates error: %v", err)
	}
	if _, err := encodeSchedulePayload(taskdomain.ScheduleTypeOddEvenDays, taskdomain.SchedulePayload{
		OddEvenDays: &taskdomain.OddEvenDaysSchedule{Mode: taskdomain.OddEvenModeOdd},
	}); err != nil {
		t.Fatalf("encode odd/even error: %v", err)
	}
	if _, err := decodeSchedulePayload(taskdomain.ScheduleTypeMonthlyDay, []byte(`{"day":10}`)); err != nil {
		t.Fatalf("decode monthly error: %v", err)
	}
	if _, err := decodeSchedulePayload(taskdomain.ScheduleTypeSpecificDates, []byte(`{"dates":["2026-05-01"]}`)); err != nil {
		t.Fatalf("decode specific dates error: %v", err)
	}
	if _, err := decodeSchedulePayload(taskdomain.ScheduleTypeOddEvenDays, []byte(`{"mode":"odd"}`)); err != nil {
		t.Fatalf("decode odd/even error: %v", err)
	}
	if _, err := encodeSchedulePayload(taskdomain.ScheduleType("bad"), taskdomain.SchedulePayload{}); err == nil {
		t.Fatalf("expected encode error for unknown type")
	}
	if _, err := decodeSchedulePayload(taskdomain.ScheduleType("bad"), []byte(`{}`)); err == nil {
		t.Fatalf("expected decode error for unknown type")
	}

	mock.ExpectQuery("INSERT INTO task_schedules").
		WithArgs(
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
		).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "base_title", "base_description", "status_template", "schedule_type", "schedule_payload", "start_date", "end_date", "is_active", "created_at", "updated_at"}).
				AddRow(int64(1), "b", "d", "new", "daily", `{"interval":2}`, now, nil, true, now, now),
		)
	created, err := repo.Create(context.Background(), &taskdomain.Schedule{
		BaseTitle:       "b",
		BaseDescription: "d",
		StatusTemplate:  taskdomain.StatusNew,
		Type:            taskdomain.ScheduleTypeDaily,
		Payload: taskdomain.SchedulePayload{
			Daily: &taskdomain.DailySchedule{Interval: 2},
		},
		StartDate: now, IsActive: true, CreatedAt: now, UpdatedAt: now,
	})
	if err != nil || created.ID != 1 {
		t.Fatalf("unexpected create schedule result: %+v err=%v", created, err)
	}

	mock.ExpectExec("UPDATE task_schedules").
		WithArgs(pgxmock.AnyArg(), int64(1)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if err := repo.Deactivate(context.Background(), 1, now); err != nil {
		t.Fatalf("unexpected deactivate error: %v", err)
	}

	mock.ExpectQuery("SELECT id, base_title, base_description, status_template, schedule_type, schedule_payload").
		WithArgs().
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "base_title", "base_description", "status_template", "schedule_type", "schedule_payload", "start_date", "end_date", "is_active", "created_at", "updated_at"}).
				AddRow(int64(1), "b", "d", "new", "daily", `{"interval":2}`, now, nil, true, now, now),
		)
	items, err := repo.List(context.Background())
	if err != nil || len(items) != 1 {
		t.Fatalf("unexpected list schedules result: len=%d err=%v", len(items), err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestScheduleRepositoryGetUpdateAndActiveInRange(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("mock init error: %v", err)
	}
	defer mock.Close()
	repo := &ScheduleRepository{pool: mock}
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery("SELECT id, base_title, base_description, status_template, schedule_type, schedule_payload").
		WithArgs(int64(10)).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "base_title", "base_description", "status_template", "schedule_type", "schedule_payload", "start_date", "end_date", "is_active", "created_at", "updated_at"}).
				AddRow(int64(10), "b", "d", "new", "daily", `{"interval":1}`, now, nil, true, now, now),
		)
	item, err := repo.GetByID(context.Background(), 10)
	if err != nil || item.ID != 10 {
		t.Fatalf("unexpected get result: %+v err=%v", item, err)
	}

	mock.ExpectQuery("UPDATE task_schedules").
		WithArgs(
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), int64(10),
		).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "base_title", "base_description", "status_template", "schedule_type", "schedule_payload", "start_date", "end_date", "is_active", "created_at", "updated_at"}).
				AddRow(int64(10), "u", "d", "new", "daily", `{"interval":1}`, now, nil, true, now, now),
		)
	_, err = repo.Update(context.Background(), &taskdomain.Schedule{
		ID:              10,
		BaseTitle:       "u",
		BaseDescription: "d",
		StatusTemplate:  taskdomain.StatusNew,
		Type:            taskdomain.ScheduleTypeDaily,
		Payload: taskdomain.SchedulePayload{
			Daily: &taskdomain.DailySchedule{Interval: 1},
		},
		StartDate: now, IsActive: true, UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}

	mock.ExpectQuery("FROM task_schedules").
		WithArgs(now, now).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "base_title", "base_description", "status_template", "schedule_type", "schedule_payload", "start_date", "end_date", "is_active", "created_at", "updated_at"}).
				AddRow(int64(10), "u", "d", "new", "daily", `{"interval":1}`, now, nil, true, now, now),
		)
	items, err := repo.ListActiveInRange(context.Background(), now, now)
	if err != nil || len(items) != 1 {
		t.Fatalf("unexpected active list result: len=%d err=%v", len(items), err)
	}
}

func TestScheduleRepositoryNotFoundAndErrors(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	repo := &ScheduleRepository{pool: mock}
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery("SELECT id, base_title, base_description, status_template, schedule_type, schedule_payload").
		WithArgs(int64(1)).
		WillReturnError(pgx.ErrNoRows)
	_, err := repo.GetByID(context.Background(), 1)
	if !errors.Is(err, taskdomain.ErrScheduleNotFound) {
		t.Fatalf("expected ErrScheduleNotFound, got %v", err)
	}

	mock.ExpectQuery("UPDATE task_schedules").
		WithArgs(
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), int64(1),
		).
		WillReturnError(pgx.ErrNoRows)
	_, err = repo.Update(context.Background(), &taskdomain.Schedule{
		ID:              1,
		BaseTitle:       "u",
		BaseDescription: "d",
		StatusTemplate:  taskdomain.StatusNew,
		Type:            taskdomain.ScheduleTypeDaily,
		Payload: taskdomain.SchedulePayload{
			Daily: &taskdomain.DailySchedule{Interval: 1},
		},
		StartDate: now, UpdatedAt: now,
	})
	if !errors.Is(err, taskdomain.ErrScheduleNotFound) {
		t.Fatalf("expected ErrScheduleNotFound on update, got %v", err)
	}

	mock.ExpectExec("UPDATE task_schedules").
		WithArgs(pgxmock.AnyArg(), int64(1)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))
	err = repo.Deactivate(context.Background(), 1, now)
	if !errors.Is(err, taskdomain.ErrScheduleNotFound) {
		t.Fatalf("expected ErrScheduleNotFound on deactivate, got %v", err)
	}
}

func TestTaskMaterializationBatch(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("mock init error: %v", err)
	}
	defer mock.Close()
	repo := &Repository{pool: mock}
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO tasks").
		WithArgs(
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	created, err := repo.CreateBatch(context.Background(), []taskusecase.MaterializedTaskInput{
		{
			ScheduleID:  1,
			PlannedFor:  now,
			Title:       "x",
			Description: "y",
			Status:      taskdomain.StatusNew,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	})
	if err != nil || created != 1 {
		t.Fatalf("unexpected create batch result: created=%d err=%v", created, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTaskMaterializationBatchBranches(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	repo := &Repository{pool: mock}
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	created, err := repo.CreateBatch(context.Background(), nil)
	if err != nil || created != 0 {
		t.Fatalf("expected empty batch fast path")
	}

	mock.ExpectBegin().WillReturnError(errors.New("begin failed"))
	_, err = repo.CreateBatch(context.Background(), []taskusecase.MaterializedTaskInput{{
		ScheduleID: 1, PlannedFor: now, Title: "x", Status: taskdomain.StatusNew, CreatedAt: now, UpdatedAt: now,
	}})
	if err == nil {
		t.Fatalf("expected begin error")
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO tasks").WillReturnError(errors.New("exec failed"))
	mock.ExpectRollback()
	_, err = repo.CreateBatch(context.Background(), []taskusecase.MaterializedTaskInput{{
		ScheduleID: 1, PlannedFor: now, Title: "x", Status: taskdomain.StatusNew, CreatedAt: now, UpdatedAt: now,
	}})
	if err == nil {
		t.Fatalf("expected exec error")
	}
}
