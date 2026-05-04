package task

import (
	"reflect"
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func TestComputePlannedDates_MonthlyDaySkipsShortMonths(t *testing.T) {
	schedule := taskdomain.Schedule{
		BaseTitle:      "Monthly 30",
		StatusTemplate: taskdomain.StatusNew,
		Type:           taskdomain.ScheduleTypeMonthlyDay,
		Payload: taskdomain.SchedulePayload{
			MonthlyDay: &taskdomain.MonthlyDaySchedule{Day: 30},
		},
		StartDate: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
		IsActive:  true,
	}

	got, err := computePlannedDates(
		schedule,
		time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []time.Time{
		time.Date(2026, time.March, 30, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC),
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected dates: got=%v want=%v", got, want)
	}
}

func TestComputePlannedDates_DailyRespectsInterval(t *testing.T) {
	schedule := taskdomain.Schedule{
		BaseTitle:      "Every 3 days",
		StatusTemplate: taskdomain.StatusNew,
		Type:           taskdomain.ScheduleTypeDaily,
		Payload: taskdomain.SchedulePayload{
			Daily: &taskdomain.DailySchedule{Interval: 3},
		},
		StartDate: time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC),
		IsActive:  true,
	}

	got, err := computePlannedDates(
		schedule,
		time.Date(2026, time.May, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.May, 10, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []time.Time{
		time.Date(2026, time.May, 4, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.May, 7, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.May, 10, 0, 0, 0, 0, time.UTC),
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected dates: got=%v want=%v", got, want)
	}
}

func TestComputePlannedDates_MonthlyDayHandlesLeapYear(t *testing.T) {
	schedule := taskdomain.Schedule{
		BaseTitle:      "Monthly 29",
		StatusTemplate: taskdomain.StatusNew,
		Type:           taskdomain.ScheduleTypeMonthlyDay,
		Payload: taskdomain.SchedulePayload{
			MonthlyDay: &taskdomain.MonthlyDaySchedule{Day: 29},
		},
		StartDate: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
		IsActive:  true,
	}

	got, err := computePlannedDates(
		schedule,
		time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.March, 31, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []time.Time{
		time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.March, 29, 0, 0, 0, 0, time.UTC),
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected dates: got=%v want=%v", got, want)
	}
}

func TestComputePlannedDates_MonthlyDaySkipsNonLeapFebruary(t *testing.T) {
	schedule := taskdomain.Schedule{
		BaseTitle:      "Monthly 29",
		StatusTemplate: taskdomain.StatusNew,
		Type:           taskdomain.ScheduleTypeMonthlyDay,
		Payload: taskdomain.SchedulePayload{
			MonthlyDay: &taskdomain.MonthlyDaySchedule{Day: 29},
		},
		StartDate: time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
		IsActive:  true,
	}

	got, err := computePlannedDates(
		schedule,
		time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, time.March, 31, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []time.Time{
		time.Date(2025, time.March, 29, 0, 0, 0, 0, time.UTC),
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected dates: got=%v want=%v", got, want)
	}
}
