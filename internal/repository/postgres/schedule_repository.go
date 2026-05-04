package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type ScheduleRepository struct {
	pool scheduleDB
}

func NewScheduleRepository(pool *pgxpool.Pool) *ScheduleRepository {
	return &ScheduleRepository{pool: pool}
}

func (r *ScheduleRepository) Create(ctx context.Context, schedule *taskdomain.Schedule) (*taskdomain.Schedule, error) {
	payload, err := encodeSchedulePayload(schedule.Type, schedule.Payload)
	if err != nil {
		return nil, err
	}

	const query = `
		INSERT INTO task_schedules (
			base_title, base_description, status_template, schedule_type, schedule_payload,
			start_date, end_date, is_active, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7, $8, $9, $10)
		RETURNING id, base_title, base_description, status_template, schedule_type, schedule_payload,
			start_date, end_date, is_active, created_at, updated_at
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		schedule.BaseTitle,
		schedule.BaseDescription,
		schedule.StatusTemplate,
		schedule.Type,
		payload,
		schedule.StartDate,
		schedule.EndDate,
		schedule.IsActive,
		schedule.CreatedAt,
		schedule.UpdatedAt,
	)

	return scanSchedule(row)
}

func (r *ScheduleRepository) GetByID(ctx context.Context, id int64) (*taskdomain.Schedule, error) {
	const query = `
		SELECT id, base_title, base_description, status_template, schedule_type, schedule_payload,
			start_date, end_date, is_active, created_at, updated_at
		FROM task_schedules
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	schedule, err := scanSchedule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrScheduleNotFound
		}
		return nil, err
	}

	return schedule, nil
}

func (r *ScheduleRepository) Update(ctx context.Context, schedule *taskdomain.Schedule) (*taskdomain.Schedule, error) {
	payload, err := encodeSchedulePayload(schedule.Type, schedule.Payload)
	if err != nil {
		return nil, err
	}

	const query = `
		UPDATE task_schedules
		SET base_title = $1,
			base_description = $2,
			status_template = $3,
			schedule_type = $4,
			schedule_payload = $5::jsonb,
			start_date = $6,
			end_date = $7,
			is_active = $8,
			updated_at = $9
		WHERE id = $10
		RETURNING id, base_title, base_description, status_template, schedule_type, schedule_payload,
			start_date, end_date, is_active, created_at, updated_at
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		schedule.BaseTitle,
		schedule.BaseDescription,
		schedule.StatusTemplate,
		schedule.Type,
		payload,
		schedule.StartDate,
		schedule.EndDate,
		schedule.IsActive,
		schedule.UpdatedAt,
		schedule.ID,
	)

	updated, err := scanSchedule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrScheduleNotFound
		}
		return nil, err
	}

	return updated, nil
}

func (r *ScheduleRepository) Deactivate(ctx context.Context, id int64, updatedAt time.Time) error {
	const query = `
		UPDATE task_schedules
		SET is_active = FALSE,
			updated_at = $1
		WHERE id = $2
	`

	result, err := r.pool.Exec(ctx, query, updatedAt, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return taskdomain.ErrScheduleNotFound
	}
	return nil
}

func (r *ScheduleRepository) List(ctx context.Context) ([]taskdomain.Schedule, error) {
	const query = `
		SELECT id, base_title, base_description, status_template, schedule_type, schedule_payload,
			start_date, end_date, is_active, created_at, updated_at
		FROM task_schedules
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schedules := make([]taskdomain.Schedule, 0)
	for rows.Next() {
		schedule, scanErr := scanSchedule(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		schedules = append(schedules, *schedule)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return schedules, nil
}

func (r *ScheduleRepository) ListActiveInRange(ctx context.Context, from, to time.Time) ([]taskdomain.Schedule, error) {
	const query = `
		SELECT id, base_title, base_description, status_template, schedule_type, schedule_payload,
			start_date, end_date, is_active, created_at, updated_at
		FROM task_schedules
		WHERE is_active = TRUE
		  AND start_date <= $2
		  AND (end_date IS NULL OR end_date >= $1)
		ORDER BY id ASC
	`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schedules := make([]taskdomain.Schedule, 0)
	for rows.Next() {
		schedule, scanErr := scanSchedule(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		schedules = append(schedules, *schedule)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return schedules, nil
}
