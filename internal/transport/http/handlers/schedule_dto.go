package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

type scheduleMutationDTO struct {
	BaseTitle       string                  `json:"base_title"`
	BaseDescription string                  `json:"base_description"`
	StatusTemplate  taskdomain.Status       `json:"status_template"`
	ScheduleType    taskdomain.ScheduleType `json:"schedule_type"`
	SchedulePayload json.RawMessage         `json:"schedule_payload"`
	StartDate       string                  `json:"start_date"`
	EndDate         *string                 `json:"end_date"`
	IsActive        *bool                   `json:"is_active,omitempty"`
}

type scheduleDTO struct {
	ID              int64                   `json:"id"`
	BaseTitle       string                  `json:"base_title"`
	BaseDescription string                  `json:"base_description"`
	StatusTemplate  taskdomain.Status       `json:"status_template"`
	ScheduleType    taskdomain.ScheduleType `json:"schedule_type"`
	SchedulePayload any                     `json:"schedule_payload"`
	StartDate       string                  `json:"start_date"`
	EndDate         *string                 `json:"end_date,omitempty"`
	IsActive        bool                    `json:"is_active"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
}

func (dto scheduleMutationDTO) toCreateInput() (taskusecase.CreateScheduleInput, error) {
	payload, err := parseSchedulePayload(dto.ScheduleType, dto.SchedulePayload)
	if err != nil {
		return taskusecase.CreateScheduleInput{}, err
	}

	startDate, err := parseDateOnly(dto.StartDate)
	if err != nil {
		return taskusecase.CreateScheduleInput{}, fmt.Errorf("invalid start_date: %w", err)
	}

	var endDate *time.Time
	if dto.EndDate != nil {
		parsed, parseErr := parseDateOnly(*dto.EndDate)
		if parseErr != nil {
			return taskusecase.CreateScheduleInput{}, fmt.Errorf("invalid end_date: %w", parseErr)
		}
		endDate = &parsed
	}

	isActive := true
	if dto.IsActive != nil {
		isActive = *dto.IsActive
	}

	return taskusecase.CreateScheduleInput{
		BaseTitle:       dto.BaseTitle,
		BaseDescription: dto.BaseDescription,
		StatusTemplate:  dto.StatusTemplate,
		Type:            dto.ScheduleType,
		Payload:         payload,
		StartDate:       startDate,
		EndDate:         endDate,
		IsActive:        isActive,
	}, nil
}

func (dto scheduleMutationDTO) toUpdateInput() (taskusecase.UpdateScheduleInput, error) {
	payload, err := parseSchedulePayload(dto.ScheduleType, dto.SchedulePayload)
	if err != nil {
		return taskusecase.UpdateScheduleInput{}, err
	}

	startDate, err := parseDateOnly(dto.StartDate)
	if err != nil {
		return taskusecase.UpdateScheduleInput{}, fmt.Errorf("invalid start_date: %w", err)
	}

	var endDate *time.Time
	if dto.EndDate != nil {
		parsed, parseErr := parseDateOnly(*dto.EndDate)
		if parseErr != nil {
			return taskusecase.UpdateScheduleInput{}, fmt.Errorf("invalid end_date: %w", parseErr)
		}
		endDate = &parsed
	}

	isActive := true
	if dto.IsActive != nil {
		isActive = *dto.IsActive
	}

	return taskusecase.UpdateScheduleInput{
		BaseTitle:       dto.BaseTitle,
		BaseDescription: dto.BaseDescription,
		StatusTemplate:  dto.StatusTemplate,
		Type:            dto.ScheduleType,
		Payload:         payload,
		StartDate:       startDate,
		EndDate:         endDate,
		IsActive:        isActive,
	}, nil
}

func newScheduleDTO(schedule *taskdomain.Schedule) scheduleDTO {
	var endDate *string
	if schedule.EndDate != nil {
		formatted := schedule.EndDate.UTC().Format(time.DateOnly)
		endDate = &formatted
	}

	return scheduleDTO{
		ID:              schedule.ID,
		BaseTitle:       schedule.BaseTitle,
		BaseDescription: schedule.BaseDescription,
		StatusTemplate:  schedule.StatusTemplate,
		ScheduleType:    schedule.Type,
		SchedulePayload: schedulePayloadToJSON(schedule.Type, schedule.Payload),
		StartDate:       schedule.StartDate.UTC().Format(time.DateOnly),
		EndDate:         endDate,
		IsActive:        schedule.IsActive,
		CreatedAt:       schedule.CreatedAt,
		UpdatedAt:       schedule.UpdatedAt,
	}
}

func parseSchedulePayload(scheduleType taskdomain.ScheduleType, raw json.RawMessage) (taskdomain.SchedulePayload, error) {
	if len(raw) == 0 {
		return taskdomain.SchedulePayload{}, errors.New("schedule_payload is required")
	}

	switch scheduleType {
	case taskdomain.ScheduleTypeDaily:
		var payload struct {
			Interval int `json:"interval"`
		}
		if err := strictUnmarshal(raw, &payload); err != nil {
			return taskdomain.SchedulePayload{}, fmt.Errorf("invalid daily payload: %w", err)
		}
		return taskdomain.SchedulePayload{
			Daily: &taskdomain.DailySchedule{Interval: payload.Interval},
		}, nil
	case taskdomain.ScheduleTypeMonthlyDay:
		var payload struct {
			Day int `json:"day"`
		}
		if err := strictUnmarshal(raw, &payload); err != nil {
			return taskdomain.SchedulePayload{}, fmt.Errorf("invalid monthly_day payload: %w", err)
		}
		return taskdomain.SchedulePayload{
			MonthlyDay: &taskdomain.MonthlyDaySchedule{Day: payload.Day},
		}, nil
	case taskdomain.ScheduleTypeSpecificDates:
		var payload struct {
			Dates []string `json:"dates"`
		}
		if err := strictUnmarshal(raw, &payload); err != nil {
			return taskdomain.SchedulePayload{}, fmt.Errorf("invalid specific_dates payload: %w", err)
		}
		dates := make([]time.Time, 0, len(payload.Dates))
		for _, rawDate := range payload.Dates {
			parsed, parseErr := parseDateOnly(rawDate)
			if parseErr != nil {
				return taskdomain.SchedulePayload{}, fmt.Errorf("invalid specific_dates date %q: %w", rawDate, parseErr)
			}
			dates = append(dates, parsed)
		}
		return taskdomain.SchedulePayload{
			SpecificDates: &taskdomain.SpecificDatesSchedule{Dates: dates},
		}, nil
	case taskdomain.ScheduleTypeOddEvenDays:
		var payload struct {
			Mode taskdomain.OddEvenMode `json:"mode"`
		}
		if err := strictUnmarshal(raw, &payload); err != nil {
			return taskdomain.SchedulePayload{}, fmt.Errorf("invalid odd_even_days payload: %w", err)
		}
		return taskdomain.SchedulePayload{
			OddEvenDays: &taskdomain.OddEvenDaysSchedule{Mode: payload.Mode},
		}, nil
	default:
		return taskdomain.SchedulePayload{}, errors.New("unsupported schedule_type")
	}
}

func schedulePayloadToJSON(scheduleType taskdomain.ScheduleType, payload taskdomain.SchedulePayload) any {
	switch scheduleType {
	case taskdomain.ScheduleTypeDaily:
		if payload.Daily == nil {
			return map[string]any{}
		}
		return map[string]any{"interval": payload.Daily.Interval}
	case taskdomain.ScheduleTypeMonthlyDay:
		if payload.MonthlyDay == nil {
			return map[string]any{}
		}
		return map[string]any{"day": payload.MonthlyDay.Day}
	case taskdomain.ScheduleTypeSpecificDates:
		if payload.SpecificDates == nil {
			return map[string]any{}
		}
		dates := make([]string, 0, len(payload.SpecificDates.Dates))
		for _, d := range payload.SpecificDates.Dates {
			dates = append(dates, d.UTC().Format(time.DateOnly))
		}
		return map[string]any{"dates": dates}
	case taskdomain.ScheduleTypeOddEvenDays:
		if payload.OddEvenDays == nil {
			return map[string]any{}
		}
		return map[string]any{"mode": payload.OddEvenDays.Mode}
	default:
		return map[string]any{}
	}
}

func parseDateOnly(value string) (time.Time, error) {
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return time.Time{}, err
	}
	return parsed.UTC(), nil
}

func strictUnmarshal(raw []byte, dst any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}

	if decoder.More() {
		return errors.New("invalid trailing data")
	}

	return nil
}
