package task

import (
	"fmt"
	"sort"
	"time"
)

type ScheduleType string

const (
	ScheduleTypeDaily         ScheduleType = "daily"
	ScheduleTypeMonthlyDay    ScheduleType = "monthly_day"
	ScheduleTypeSpecificDates ScheduleType = "specific_dates"
	ScheduleTypeOddEvenDays   ScheduleType = "odd_even_days"
)

type OddEvenMode string

const (
	OddEvenModeOdd  OddEvenMode = "odd"
	OddEvenModeEven OddEvenMode = "even"
)

type DailySchedule struct {
	Interval int `json:"interval"`
}

type MonthlyDaySchedule struct {
	Day int `json:"day"`
}

type SpecificDatesSchedule struct {
	Dates []time.Time `json:"dates"`
}

type OddEvenDaysSchedule struct {
	Mode OddEvenMode `json:"mode"`
}

type SchedulePayload struct {
	Daily         *DailySchedule         `json:"daily,omitempty"`
	MonthlyDay    *MonthlyDaySchedule    `json:"monthly_day,omitempty"`
	SpecificDates *SpecificDatesSchedule `json:"specific_dates,omitempty"`
	OddEvenDays   *OddEvenDaysSchedule   `json:"odd_even_days,omitempty"`
}

type Schedule struct {
	ID              int64           `json:"id"`
	BaseTitle       string          `json:"base_title"`
	BaseDescription string          `json:"base_description"`
	StatusTemplate  Status          `json:"status_template"`
	Type            ScheduleType    `json:"schedule_type"`
	Payload         SchedulePayload `json:"schedule_payload"`
	StartDate       time.Time       `json:"start_date"`
	EndDate         *time.Time      `json:"end_date,omitempty"`
	IsActive        bool            `json:"is_active"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

func (s *Schedule) Validate() error {
	if s.BaseTitle == "" {
		return fmt.Errorf("%w: base_title is required", ErrInvalidSchedule)
	}

	if !s.StatusTemplate.Valid() {
		return fmt.Errorf("%w: invalid status_template", ErrInvalidSchedule)
	}

	if s.StartDate.IsZero() {
		return fmt.Errorf("%w: start_date is required", ErrInvalidSchedule)
	}

	startDate := normalizeDate(s.StartDate)
	if s.EndDate != nil {
		endDate := normalizeDate(*s.EndDate)
		if endDate.Before(startDate) {
			return fmt.Errorf("%w: end_date must be greater than or equal to start_date", ErrInvalidSchedule)
		}
	}

	switch s.Type {
	case ScheduleTypeDaily:
		if s.Payload.Daily == nil {
			return fmt.Errorf("%w: daily payload is required", ErrInvalidSchedule)
		}
		if s.Payload.Daily.Interval <= 0 {
			return fmt.Errorf("%w: daily interval must be positive", ErrInvalidSchedule)
		}
	case ScheduleTypeMonthlyDay:
		if s.Payload.MonthlyDay == nil {
			return fmt.Errorf("%w: monthly_day payload is required", ErrInvalidSchedule)
		}
		if s.Payload.MonthlyDay.Day < 1 || s.Payload.MonthlyDay.Day > 30 {
			return fmt.Errorf("%w: monthly_day.day must be in range 1..30", ErrInvalidSchedule)
		}
	case ScheduleTypeSpecificDates:
		if s.Payload.SpecificDates == nil {
			return fmt.Errorf("%w: specific_dates payload is required", ErrInvalidSchedule)
		}
		if len(s.Payload.SpecificDates.Dates) == 0 {
			return fmt.Errorf("%w: specific_dates must contain at least one date", ErrInvalidSchedule)
		}

		seen := make(map[string]struct{}, len(s.Payload.SpecificDates.Dates))
		for _, date := range s.Payload.SpecificDates.Dates {
			normalized := normalizeDate(date)
			if normalized.Before(startDate) {
				return fmt.Errorf("%w: specific_dates cannot contain dates before start_date", ErrInvalidSchedule)
			}
			key := normalized.Format(time.DateOnly)
			if _, ok := seen[key]; ok {
				return fmt.Errorf("%w: specific_dates cannot contain duplicates", ErrInvalidSchedule)
			}
			seen[key] = struct{}{}
		}

		sort.Slice(s.Payload.SpecificDates.Dates, func(i, j int) bool {
			return normalizeDate(s.Payload.SpecificDates.Dates[i]).Before(normalizeDate(s.Payload.SpecificDates.Dates[j]))
		})
	case ScheduleTypeOddEvenDays:
		if s.Payload.OddEvenDays == nil {
			return fmt.Errorf("%w: odd_even_days payload is required", ErrInvalidSchedule)
		}
		if s.Payload.OddEvenDays.Mode != OddEvenModeOdd && s.Payload.OddEvenDays.Mode != OddEvenModeEven {
			return fmt.Errorf("%w: odd_even_days.mode must be odd or even", ErrInvalidSchedule)
		}
	default:
		return fmt.Errorf("%w: unknown schedule_type", ErrInvalidSchedule)
	}

	if err := ensureOnlyPayloadForType(s.Type, s.Payload); err != nil {
		return err
	}

	s.StartDate = startDate
	if s.EndDate != nil {
		date := normalizeDate(*s.EndDate)
		s.EndDate = &date
	}
	return nil
}

func ensureOnlyPayloadForType(scheduleType ScheduleType, payload SchedulePayload) error {
	payloadCount := 0
	if payload.Daily != nil {
		payloadCount++
	}
	if payload.MonthlyDay != nil {
		payloadCount++
	}
	if payload.SpecificDates != nil {
		payloadCount++
	}
	if payload.OddEvenDays != nil {
		payloadCount++
	}

	if payloadCount != 1 {
		return fmt.Errorf("%w: exactly one payload type must be set", ErrInvalidSchedule)
	}

	switch scheduleType {
	case ScheduleTypeDaily:
		if payload.Daily == nil {
			return fmt.Errorf("%w: payload does not match schedule_type", ErrInvalidSchedule)
		}
	case ScheduleTypeMonthlyDay:
		if payload.MonthlyDay == nil {
			return fmt.Errorf("%w: payload does not match schedule_type", ErrInvalidSchedule)
		}
	case ScheduleTypeSpecificDates:
		if payload.SpecificDates == nil {
			return fmt.Errorf("%w: payload does not match schedule_type", ErrInvalidSchedule)
		}
	case ScheduleTypeOddEvenDays:
		if payload.OddEvenDays == nil {
			return fmt.Errorf("%w: payload does not match schedule_type", ErrInvalidSchedule)
		}
	default:
		return fmt.Errorf("%w: unknown schedule_type", ErrInvalidSchedule)
	}

	return nil
}

func normalizeDate(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}
