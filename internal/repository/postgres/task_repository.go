package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, title, description, status, schedule_id, planned_for, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.CreatedAt, task.UpdatedAt)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, schedule_id, planned_for, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return found, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			updated_at = $4
		WHERE id = $5
		RETURNING id, title, description, status, schedule_id, planned_for, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.UpdatedAt, task.ID)
	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context, filter taskusecase.ListFilter) ([]taskdomain.Task, error) {
	query := `
		SELECT id, title, description, status, schedule_id, planned_for, created_at, updated_at
		FROM tasks
	`

	args := make([]any, 0, 3)
	conditions := make([]string, 0, 3)

	if filter.PlannedForFrom != nil {
		args = append(args, *filter.PlannedForFrom)
		conditions = append(conditions, fmt.Sprintf("planned_for >= $%d", len(args)))
	}
	if filter.PlannedForTo != nil {
		args = append(args, *filter.PlannedForTo)
		conditions = append(conditions, fmt.Sprintf("planned_for <= $%d", len(args)))
	}
	if filter.ScheduleID != nil {
		args = append(args, *filter.ScheduleID)
		conditions = append(conditions, fmt.Sprintf("schedule_id = $%d", len(args)))
	}

	if len(conditions) > 0 {
		query += "\nWHERE " + strings.Join(conditions, " AND ")
	}
	query += "\nORDER BY id DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task       taskdomain.Task
		status     string
		scheduleID sql.NullInt64
		plannedFor sql.NullTime
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&scheduleID,
		&plannedFor,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)
	if scheduleID.Valid {
		value := scheduleID.Int64
		task.ScheduleID = &value
	}
	if plannedFor.Valid {
		value := plannedFor.Time.UTC()
		task.PlannedFor = &value
	}

	return &task, nil
}
