package postgres

import (
	"encoding/json"
	"fmt"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func encodeSchedulePayload(scheduleType taskdomain.ScheduleType, payload taskdomain.SchedulePayload) ([]byte, error) {
	switch scheduleType {
	case taskdomain.ScheduleTypeDaily:
		if payload.Daily == nil {
			return nil, fmt.Errorf("%w: daily payload is required", taskdomain.ErrInvalidSchedule)
		}
		return json.Marshal(map[string]int{
			"interval": payload.Daily.Interval,
		})
	case taskdomain.ScheduleTypeMonthlyDay:
		if payload.MonthlyDay == nil {
			return nil, fmt.Errorf("%w: monthly_day payload is required", taskdomain.ErrInvalidSchedule)
		}
		return json.Marshal(map[string]int{
			"day": payload.MonthlyDay.Day,
		})
	case taskdomain.ScheduleTypeSpecificDates:
		if payload.SpecificDates == nil {
			return nil, fmt.Errorf("%w: specific_dates payload is required", taskdomain.ErrInvalidSchedule)
		}
		dates := make([]string, 0, len(payload.SpecificDates.Dates))
		for _, date := range payload.SpecificDates.Dates {
			dates = append(dates, date.UTC().Format(time.DateOnly))
		}
		return json.Marshal(map[string][]string{
			"dates": dates,
		})
	case taskdomain.ScheduleTypeOddEvenDays:
		if payload.OddEvenDays == nil {
			return nil, fmt.Errorf("%w: odd_even_days payload is required", taskdomain.ErrInvalidSchedule)
		}
		return json.Marshal(map[string]taskdomain.OddEvenMode{
			"mode": payload.OddEvenDays.Mode,
		})
	default:
		return nil, fmt.Errorf("%w: unknown schedule type", taskdomain.ErrInvalidSchedule)
	}
}

func decodeSchedulePayload(scheduleType taskdomain.ScheduleType, raw []byte) (taskdomain.SchedulePayload, error) {
	switch scheduleType {
	case taskdomain.ScheduleTypeDaily:
		var payload struct {
			Interval int `json:"interval"`
		}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return taskdomain.SchedulePayload{}, err
		}
		return taskdomain.SchedulePayload{
			Daily: &taskdomain.DailySchedule{Interval: payload.Interval},
		}, nil
	case taskdomain.ScheduleTypeMonthlyDay:
		var payload struct {
			Day int `json:"day"`
		}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return taskdomain.SchedulePayload{}, err
		}
		return taskdomain.SchedulePayload{
			MonthlyDay: &taskdomain.MonthlyDaySchedule{Day: payload.Day},
		}, nil
	case taskdomain.ScheduleTypeSpecificDates:
		var payload struct {
			Dates []string `json:"dates"`
		}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return taskdomain.SchedulePayload{}, err
		}
		dates := make([]time.Time, 0, len(payload.Dates))
		for _, date := range payload.Dates {
			parsed, err := time.Parse(time.DateOnly, date)
			if err != nil {
				return taskdomain.SchedulePayload{}, err
			}
			dates = append(dates, parsed.UTC())
		}
		return taskdomain.SchedulePayload{
			SpecificDates: &taskdomain.SpecificDatesSchedule{Dates: dates},
		}, nil
	case taskdomain.ScheduleTypeOddEvenDays:
		var payload struct {
			Mode taskdomain.OddEvenMode `json:"mode"`
		}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return taskdomain.SchedulePayload{}, err
		}
		return taskdomain.SchedulePayload{
			OddEvenDays: &taskdomain.OddEvenDaysSchedule{Mode: payload.Mode},
		}, nil
	default:
		return taskdomain.SchedulePayload{}, fmt.Errorf("%w: unknown schedule type", taskdomain.ErrInvalidSchedule)
	}
}
