package task

import (
	"fmt"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func computePlannedDates(schedule taskdomain.Schedule, from, to time.Time) ([]time.Time, error) {
	if err := schedule.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSchedule, err)
	}

	fromDate := normalizeDateUTC(from)
	toDate := normalizeDateUTC(to)
	if toDate.Before(fromDate) {
		return nil, fmt.Errorf("%w: from must be before or equal to to", ErrInvalidSchedule)
	}

	effectiveFrom, effectiveTo, hasIntersection := effectiveRange(schedule, fromDate, toDate)
	if !hasIntersection {
		return []time.Time{}, nil
	}

	switch schedule.Type {
	case taskdomain.ScheduleTypeDaily:
		return plannedDatesDaily(schedule, effectiveFrom, effectiveTo), nil
	case taskdomain.ScheduleTypeMonthlyDay:
		return plannedDatesMonthlyDay(schedule, effectiveFrom, effectiveTo), nil
	case taskdomain.ScheduleTypeSpecificDates:
		return plannedDatesSpecificDates(schedule, effectiveFrom, effectiveTo), nil
	case taskdomain.ScheduleTypeOddEvenDays:
		return plannedDatesOddEvenDays(schedule, effectiveFrom, effectiveTo), nil
	default:
		return nil, fmt.Errorf("%w: unknown schedule type", ErrInvalidSchedule)
	}
}

func effectiveRange(schedule taskdomain.Schedule, fromDate, toDate time.Time) (time.Time, time.Time, bool) {
	start := maxDate(normalizeDateUTC(schedule.StartDate), fromDate)
	end := toDate
	if schedule.EndDate != nil {
		end = minDate(end, normalizeDateUTC(*schedule.EndDate))
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, false
	}
	return start, end, true
}

func plannedDatesDaily(schedule taskdomain.Schedule, from, to time.Time) []time.Time {
	interval := schedule.Payload.Daily.Interval
	start := normalizeDateUTC(schedule.StartDate)

	daysSinceStart := int(from.Sub(start).Hours() / 24)
	offset := daysSinceStart % interval
	first := from
	if offset != 0 {
		first = from.AddDate(0, 0, interval-offset)
	}

	result := make([]time.Time, 0)
	for date := first; !date.After(to); date = date.AddDate(0, 0, interval) {
		result = append(result, date)
	}

	return result
}

func plannedDatesMonthlyDay(schedule taskdomain.Schedule, from, to time.Time) []time.Time {
	targetDay := schedule.Payload.MonthlyDay.Day
	current := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, time.UTC)
	limitMonth := time.Date(to.Year(), to.Month(), 1, 0, 0, 0, 0, time.UTC)

	result := make([]time.Time, 0)
	for !current.After(limitMonth) {
		daysInCurrentMonth := daysInMonth(current.Year(), current.Month())
		if targetDay <= daysInCurrentMonth {
			candidate := time.Date(current.Year(), current.Month(), targetDay, 0, 0, 0, 0, time.UTC)
			if !candidate.Before(from) && !candidate.After(to) {
				result = append(result, candidate)
			}
		}
		current = current.AddDate(0, 1, 0)
	}

	return result
}

func plannedDatesSpecificDates(schedule taskdomain.Schedule, from, to time.Time) []time.Time {
	result := make([]time.Time, 0)
	for _, date := range schedule.Payload.SpecificDates.Dates {
		normalized := normalizeDateUTC(date)
		if !normalized.Before(from) && !normalized.After(to) {
			result = append(result, normalized)
		}
	}

	return result
}

func plannedDatesOddEvenDays(schedule taskdomain.Schedule, from, to time.Time) []time.Time {
	result := make([]time.Time, 0)
	for date := from; !date.After(to); date = date.AddDate(0, 0, 1) {
		isOdd := date.Day()%2 == 1
		if schedule.Payload.OddEvenDays.Mode == taskdomain.OddEvenModeOdd && isOdd {
			result = append(result, date)
			continue
		}
		if schedule.Payload.OddEvenDays.Mode == taskdomain.OddEvenModeEven && !isOdd {
			result = append(result, date)
		}
	}
	return result
}

func normalizeDateUTC(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func minDate(left, right time.Time) time.Time {
	if left.Before(right) {
		return left
	}
	return right
}

func maxDate(left, right time.Time) time.Time {
	if left.After(right) {
		return left
	}
	return right
}

func daysInMonth(year int, month time.Month) int {
	// Day 0 of next month gives us the last day of current month.
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
