package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"

	taskusecase "example.com/taskservice/internal/usecase/task"
)

func (r *Repository) CreateBatch(ctx context.Context, tasks []taskusecase.MaterializedTaskInput) (int, error) {
	if len(tasks) == 0 {
		return 0, nil
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	const query = `
		INSERT INTO tasks (
			title, description, status, created_at, updated_at, schedule_id, planned_for
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (schedule_id, planned_for)
		WHERE schedule_id IS NOT NULL AND planned_for IS NOT NULL
		DO NOTHING
	`

	created := 0
	for i := range tasks {
		result, execErr := tx.Exec(
			ctx,
			query,
			tasks[i].Title,
			tasks[i].Description,
			tasks[i].Status,
			tasks[i].CreatedAt,
			tasks[i].UpdatedAt,
			tasks[i].ScheduleID,
			tasks[i].PlannedFor,
		)
		if execErr != nil {
			return 0, execErr
		}
		created += int(result.RowsAffected())
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return created, nil
}
