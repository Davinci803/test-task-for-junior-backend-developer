package handlers

import (
	"encoding/json"
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func TestScheduleMutationDTOConversions(t *testing.T) {
	base := scheduleMutationDTO{
		BaseTitle:       "x",
		BaseDescription: "y",
		StatusTemplate:  taskdomain.StatusNew,
		ScheduleType:    taskdomain.ScheduleTypeDaily,
		SchedulePayload: json.RawMessage(`{"interval":2}`),
		StartDate:       "2026-05-01",
		IsActive:        ptrBool(true),
	}

	if _, err := base.toCreateInput(); err != nil {
		t.Fatalf("unexpected create input error: %v", err)
	}
	if _, err := base.toUpdateInput(); err != nil {
		t.Fatalf("unexpected update input error: %v", err)
	}

	bad := base
	bad.StartDate = "bad"
	if _, err := bad.toCreateInput(); err == nil {
		t.Fatalf("expected bad start_date error")
	}
}

func TestSchedulePayloadToJSONBranches(t *testing.T) {
	daily := schedulePayloadToJSON(taskdomain.ScheduleTypeDaily, taskdomain.SchedulePayload{
		Daily: &taskdomain.DailySchedule{Interval: 1},
	})
	if daily == nil {
		t.Fatalf("expected daily payload json")
	}
	monthly := schedulePayloadToJSON(taskdomain.ScheduleTypeMonthlyDay, taskdomain.SchedulePayload{
		MonthlyDay: &taskdomain.MonthlyDaySchedule{Day: 10},
	})
	if monthly == nil {
		t.Fatalf("expected monthly payload json")
	}
	specific := schedulePayloadToJSON(taskdomain.ScheduleTypeSpecificDates, taskdomain.SchedulePayload{
		SpecificDates: &taskdomain.SpecificDatesSchedule{Dates: []time.Time{time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)}},
	})
	if specific == nil {
		t.Fatalf("expected specific payload json")
	}
	oddEven := schedulePayloadToJSON(taskdomain.ScheduleTypeOddEvenDays, taskdomain.SchedulePayload{
		OddEvenDays: &taskdomain.OddEvenDaysSchedule{Mode: taskdomain.OddEvenModeEven},
	})
	if oddEven == nil {
		t.Fatalf("expected odd/even payload json")
	}
}

func TestParseSchedulePayloadBranches(t *testing.T) {
	cases := []struct {
		typ taskdomain.ScheduleType
		raw string
	}{
		{taskdomain.ScheduleTypeDaily, `{"interval":1}`},
		{taskdomain.ScheduleTypeMonthlyDay, `{"day":10}`},
		{taskdomain.ScheduleTypeSpecificDates, `{"dates":["2026-05-10"]}`},
		{taskdomain.ScheduleTypeOddEvenDays, `{"mode":"odd"}`},
	}

	for _, tc := range cases {
		if _, err := parseSchedulePayload(tc.typ, json.RawMessage(tc.raw)); err != nil {
			t.Fatalf("unexpected parse error for %s: %v", tc.typ, err)
		}
	}

	if _, err := parseSchedulePayload(taskdomain.ScheduleType("x"), json.RawMessage(`{}`)); err == nil {
		t.Fatalf("expected unsupported schedule type error")
	}
}

func ptrBool(v bool) *bool { return &v }
