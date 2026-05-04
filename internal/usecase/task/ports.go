package task

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
}

type CreateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
}

type UpdateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
}

type ScheduleRepository interface {
	Create(ctx context.Context, schedule *taskdomain.Schedule) (*taskdomain.Schedule, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Schedule, error)
	Update(ctx context.Context, schedule *taskdomain.Schedule) (*taskdomain.Schedule, error)
	Deactivate(ctx context.Context, id int64, updatedAt time.Time) error
	List(ctx context.Context) ([]taskdomain.Schedule, error)
	ListActiveInRange(ctx context.Context, from, to time.Time) ([]taskdomain.Schedule, error)
}

type MaterializedTaskInput struct {
	ScheduleID  int64
	PlannedFor  time.Time
	Title       string
	Description string
	Status      taskdomain.Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TaskMaterializationRepository interface {
	CreateBatch(ctx context.Context, tasks []MaterializedTaskInput) (created int, err error)
}

type ScheduleUsecase interface {
	CreateSchedule(ctx context.Context, input CreateScheduleInput) (*taskdomain.Schedule, error)
	UpdateSchedule(ctx context.Context, id int64, input UpdateScheduleInput) (*taskdomain.Schedule, error)
	DeactivateSchedule(ctx context.Context, id int64) error
	GetScheduleByID(ctx context.Context, id int64) (*taskdomain.Schedule, error)
	ListSchedules(ctx context.Context) ([]taskdomain.Schedule, error)
	ComputePlannedDates(ctx context.Context, id int64, from, to time.Time) ([]time.Time, error)
	GenerateTasks(ctx context.Context, from, to time.Time) (int, error)
}

type CreateScheduleInput struct {
	BaseTitle       string
	BaseDescription string
	StatusTemplate  taskdomain.Status
	Type            taskdomain.ScheduleType
	Payload         taskdomain.SchedulePayload
	StartDate       time.Time
	EndDate         *time.Time
	IsActive        bool
}

type UpdateScheduleInput struct {
	BaseTitle       string
	BaseDescription string
	StatusTemplate  taskdomain.Status
	Type            taskdomain.ScheduleType
	Payload         taskdomain.SchedulePayload
	StartDate       time.Time
	EndDate         *time.Time
	IsActive        bool
}
