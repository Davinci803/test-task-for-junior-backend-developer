package task

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type ScheduleService struct {
	scheduleRepo    ScheduleRepository
	materializeRepo TaskMaterializationRepository
	now             func() time.Time
}

func NewScheduleService(scheduleRepo ScheduleRepository, materializeRepo TaskMaterializationRepository) *ScheduleService {
	return &ScheduleService{
		scheduleRepo:    scheduleRepo,
		materializeRepo: materializeRepo,
		now:             func() time.Time { return time.Now().UTC() },
	}
}

func (s *ScheduleService) CreateSchedule(ctx context.Context, input CreateScheduleInput) (*taskdomain.Schedule, error) {
	schedule, err := buildScheduleFromCreateInput(input)
	if err != nil {
		return nil, err
	}

	now := s.now()
	schedule.CreatedAt = now
	schedule.UpdatedAt = now

	created, err := s.scheduleRepo.Create(ctx, schedule)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *ScheduleService) UpdateSchedule(ctx context.Context, id int64, input UpdateScheduleInput) (*taskdomain.Schedule, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidSchedule)
	}

	schedule, err := buildScheduleFromUpdateInput(input)
	if err != nil {
		return nil, err
	}
	schedule.ID = id
	schedule.UpdatedAt = s.now()

	updated, err := s.scheduleRepo.Update(ctx, schedule)
	if err != nil {
		if errors.Is(err, taskdomain.ErrScheduleNotFound) {
			return nil, ErrScheduleNotFound
		}
		return nil, err
	}

	return updated, nil
}

func (s *ScheduleService) DeactivateSchedule(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidSchedule)
	}

	if err := s.scheduleRepo.Deactivate(ctx, id, s.now()); err != nil {
		if errors.Is(err, taskdomain.ErrScheduleNotFound) {
			return ErrScheduleNotFound
		}
		return err
	}

	return nil
}

func (s *ScheduleService) GetScheduleByID(ctx context.Context, id int64) (*taskdomain.Schedule, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidSchedule)
	}

	schedule, err := s.scheduleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, taskdomain.ErrScheduleNotFound) {
			return nil, ErrScheduleNotFound
		}
		return nil, err
	}

	return schedule, nil
}

func (s *ScheduleService) ListSchedules(ctx context.Context) ([]taskdomain.Schedule, error) {
	return s.scheduleRepo.List(ctx)
}

func (s *ScheduleService) ComputePlannedDates(ctx context.Context, id int64, from, to time.Time) ([]time.Time, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidSchedule)
	}

	schedule, err := s.scheduleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, taskdomain.ErrScheduleNotFound) {
			return nil, ErrScheduleNotFound
		}
		return nil, err
	}

	return computePlannedDates(*schedule, from, to)
}

func (s *ScheduleService) GenerateTasks(ctx context.Context, from, to time.Time) (int, error) {
	if s.materializeRepo == nil {
		return 0, errors.New("task materialization repository is not configured")
	}

	fromDate := normalizeDateUTC(from)
	toDate := normalizeDateUTC(to)
	if toDate.Before(fromDate) {
		return 0, fmt.Errorf("%w: invalid date range", ErrInvalidSchedule)
	}

	schedules, err := s.scheduleRepo.ListActiveInRange(ctx, fromDate, toDate)
	if err != nil {
		return 0, err
	}

	batch := make([]MaterializedTaskInput, 0)
	now := s.now()
	for i := range schedules {
		dates, err := computePlannedDates(schedules[i], fromDate, toDate)
		if err != nil {
			return 0, err
		}

		for _, plannedFor := range dates {
			batch = append(batch, MaterializedTaskInput{
				ScheduleID:  schedules[i].ID,
				PlannedFor:  plannedFor,
				Title:       schedules[i].BaseTitle,
				Description: schedules[i].BaseDescription,
				Status:      schedules[i].StatusTemplate,
				CreatedAt:   now,
				UpdatedAt:   now,
			})
		}
	}

	if len(batch) == 0 {
		return 0, nil
	}

	created, err := s.materializeRepo.CreateBatch(ctx, batch)
	if err != nil {
		return 0, err
	}

	return created, nil
}

func buildScheduleFromCreateInput(input CreateScheduleInput) (*taskdomain.Schedule, error) {
	model := &taskdomain.Schedule{
		BaseTitle:       strings.TrimSpace(input.BaseTitle),
		BaseDescription: strings.TrimSpace(input.BaseDescription),
		StatusTemplate:  input.StatusTemplate,
		Type:            input.Type,
		Payload:         input.Payload,
		StartDate:       input.StartDate,
		EndDate:         input.EndDate,
		IsActive:        true,
	}

	if err := model.Validate(); err != nil {
		if errors.Is(err, taskdomain.ErrInvalidSchedule) {
			return nil, fmt.Errorf("%w: %v", ErrInvalidSchedule, err)
		}
		return nil, err
	}

	return model, nil
}

func buildScheduleFromUpdateInput(input UpdateScheduleInput) (*taskdomain.Schedule, error) {
	model := &taskdomain.Schedule{
		BaseTitle:       strings.TrimSpace(input.BaseTitle),
		BaseDescription: strings.TrimSpace(input.BaseDescription),
		StatusTemplate:  input.StatusTemplate,
		Type:            input.Type,
		Payload:         input.Payload,
		StartDate:       input.StartDate,
		EndDate:         input.EndDate,
		IsActive:        input.IsActive,
	}

	if err := model.Validate(); err != nil {
		if errors.Is(err, taskdomain.ErrInvalidSchedule) {
			return nil, fmt.Errorf("%w: %v", ErrInvalidSchedule, err)
		}
		return nil, err
	}

	return model, nil
}
