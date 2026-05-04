package task

import (
	"errors"
	"testing"
	"time"
)

func TestScheduleValidate_DailyIntervalMustBePositive(t *testing.T) {
	schedule := validBaseSchedule()
	schedule.Type = ScheduleTypeDaily
	schedule.Payload = SchedulePayload{
		Daily: &DailySchedule{Interval: 0},
	}

	err := schedule.Validate()
	if !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("expected ErrInvalidSchedule, got: %v", err)
	}
}

func TestScheduleValidate_MonthlyDayOutOfRange(t *testing.T) {
	schedule := validBaseSchedule()
	schedule.Type = ScheduleTypeMonthlyDay
	schedule.Payload = SchedulePayload{
		MonthlyDay: &MonthlyDaySchedule{Day: 31},
	}

	err := schedule.Validate()
	if !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("expected ErrInvalidSchedule, got: %v", err)
	}
}

func TestScheduleValidate_SpecificDatesRejectDuplicates(t *testing.T) {
	schedule := validBaseSchedule()
	schedule.Type = ScheduleTypeSpecificDates
	date := time.Date(2026, time.May, 10, 12, 0, 0, 0, time.UTC)
	schedule.Payload = SchedulePayload{
		SpecificDates: &SpecificDatesSchedule{
			Dates: []time.Time{date, date},
		},
	}

	err := schedule.Validate()
	if !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("expected ErrInvalidSchedule, got: %v", err)
	}
}

func TestScheduleValidate_OddEvenMode(t *testing.T) {
	schedule := validBaseSchedule()
	schedule.Type = ScheduleTypeOddEvenDays
	schedule.Payload = SchedulePayload{
		OddEvenDays: &OddEvenDaysSchedule{Mode: "bad"},
	}

	err := schedule.Validate()
	if !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("expected ErrInvalidSchedule, got: %v", err)
	}
}

func TestScheduleValidate_SpecificDatesMustNotBeEmpty(t *testing.T) {
	schedule := validBaseSchedule()
	schedule.Type = ScheduleTypeSpecificDates
	schedule.Payload = SchedulePayload{
		SpecificDates: &SpecificDatesSchedule{Dates: []time.Time{}},
	}

	err := schedule.Validate()
	if !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("expected ErrInvalidSchedule, got: %v", err)
	}
}

func TestScheduleValidate_PayloadMustMatchType(t *testing.T) {
	schedule := validBaseSchedule()
	schedule.Type = ScheduleTypeDaily
	schedule.Payload = SchedulePayload{
		MonthlyDay: &MonthlyDaySchedule{Day: 15},
	}

	err := schedule.Validate()
	if !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("expected ErrInvalidSchedule, got: %v", err)
	}
}

func TestEnsureOnlyPayloadForType(t *testing.T) {
	err := ensureOnlyPayloadForType(ScheduleTypeDaily, SchedulePayload{
		Daily: &DailySchedule{Interval: 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = ensureOnlyPayloadForType(ScheduleTypeDaily, SchedulePayload{
		Daily:      &DailySchedule{Interval: 1},
		MonthlyDay: &MonthlyDaySchedule{Day: 10},
	})
	if !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("expected ErrInvalidSchedule, got: %v", err)
	}

	err = ensureOnlyPayloadForType(ScheduleType("unknown"), SchedulePayload{
		Daily: &DailySchedule{Interval: 1},
	})
	if !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("expected ErrInvalidSchedule for unknown type, got: %v", err)
	}
}

func TestScheduleValidateCommonFields(t *testing.T) {
	schedule := validBaseSchedule()
	schedule.BaseTitle = ""
	schedule.Type = ScheduleTypeDaily
	schedule.Payload = SchedulePayload{Daily: &DailySchedule{Interval: 1}}
	if err := schedule.Validate(); !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("expected invalid base title")
	}

	schedule = validBaseSchedule()
	schedule.StatusTemplate = Status("bad")
	schedule.Type = ScheduleTypeDaily
	schedule.Payload = SchedulePayload{Daily: &DailySchedule{Interval: 1}}
	if err := schedule.Validate(); !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("expected invalid status template")
	}

	schedule = validBaseSchedule()
	schedule.StartDate = time.Time{}
	schedule.Type = ScheduleTypeDaily
	schedule.Payload = SchedulePayload{Daily: &DailySchedule{Interval: 1}}
	if err := schedule.Validate(); !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("expected invalid start date")
	}
}

func TestScheduleValidateSpecificDatesBeforeStart(t *testing.T) {
	schedule := validBaseSchedule()
	schedule.Type = ScheduleTypeSpecificDates
	schedule.Payload = SchedulePayload{
		SpecificDates: &SpecificDatesSchedule{
			Dates: []time.Time{
				time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	if err := schedule.Validate(); !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("expected invalid specific date before start")
	}
}

func TestScheduleValidateUnknownType(t *testing.T) {
	schedule := validBaseSchedule()
	schedule.Type = ScheduleType("unknown")
	schedule.Payload = SchedulePayload{
		Daily: &DailySchedule{Interval: 1},
	}
	if err := schedule.Validate(); !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("expected invalid unknown type")
	}
}

func TestScheduleValidate_EndDateBeforeStartDate(t *testing.T) {
	schedule := validBaseSchedule()
	end := schedule.StartDate.AddDate(0, 0, -1)
	schedule.Type = ScheduleTypeDaily
	schedule.Payload = SchedulePayload{
		Daily: &DailySchedule{Interval: 1},
	}
	schedule.EndDate = &end

	err := schedule.Validate()
	if !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("expected ErrInvalidSchedule, got: %v", err)
	}
}

func validBaseSchedule() Schedule {
	return Schedule{
		BaseTitle:      "Pay rent",
		StatusTemplate: StatusNew,
		StartDate:      time.Date(2026, time.May, 1, 10, 15, 0, 0, time.UTC),
		IsActive:       true,
	}
}
