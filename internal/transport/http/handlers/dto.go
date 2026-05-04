package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
}

type taskDTO struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	ScheduleID  *int64            `json:"schedule_id,omitempty"`
	PlannedFor  *string           `json:"planned_for,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	var plannedFor *string
	if task.PlannedFor != nil {
		formatted := task.PlannedFor.UTC().Format(time.DateOnly)
		plannedFor = &formatted
	}

	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		ScheduleID:  task.ScheduleID,
		PlannedFor:  plannedFor,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}
