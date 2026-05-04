package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

func TestScheduleHandlerCRUD(t *testing.T) {
	usecase := &scheduleUsecaseMock{}
	handler := NewScheduleHandler(usecase)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/schedules", bytes.NewBufferString(`{
		"base_title":"s","base_description":"d","status_template":"new","schedule_type":"daily",
		"schedule_payload":{"interval":1},"start_date":"2026-05-01","is_active":true
	}`))
	createRec := httptest.NewRecorder()
	handler.Create(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("unexpected create code: %d", createRec.Code)
	}

	listRec := httptest.NewRecorder()
	handler.List(listRec, httptest.NewRequest(http.MethodGet, "/api/v1/schedules", nil))
	if listRec.Code != http.StatusOK {
		t.Fatalf("unexpected list code: %d", listRec.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/schedules/1", nil)
	getReq = mux.SetURLVars(getReq, map[string]string{"id": "1"})
	getRec := httptest.NewRecorder()
	handler.GetByID(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("unexpected get code: %d", getRec.Code)
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/schedules/1", bytes.NewBufferString(`{
		"base_title":"u","base_description":"d","status_template":"new","schedule_type":"monthly_day",
		"schedule_payload":{"day":15},"start_date":"2026-05-01","is_active":true
	}`))
	updateReq = mux.SetURLVars(updateReq, map[string]string{"id": "1"})
	updateRec := httptest.NewRecorder()
	handler.Update(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("unexpected update code: %d", updateRec.Code)
	}

	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/schedules/1", nil)
	delReq = mux.SetURLVars(delReq, map[string]string{"id": "1"})
	delRec := httptest.NewRecorder()
	handler.Delete(delRec, delReq)
	if delRec.Code != http.StatusNoContent {
		t.Fatalf("unexpected delete code: %d", delRec.Code)
	}
}

func TestScheduleHandlerBadRequests(t *testing.T) {
	usecase := &scheduleUsecaseMock{createErr: taskusecase.ErrInvalidSchedule}
	handler := NewScheduleHandler(usecase)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/schedules", bytes.NewBufferString(`{"base_title":"x"}`))
	rec := httptest.NewRecorder()
	handler.Create(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/schedules/bad", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "bad"})
	rec = httptest.NewRecorder()
	handler.GetByID(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for invalid id: %d", rec.Code)
	}
}

func TestParseSchedulePayloadAndDTO(t *testing.T) {
	payload, err := parseSchedulePayload(taskdomain.ScheduleTypeSpecificDates, []byte(`{"dates":["2026-05-10"]}`))
	if err != nil || payload.SpecificDates == nil || len(payload.SpecificDates.Dates) != 1 {
		t.Fatalf("unexpected payload parse result: %+v err=%v", payload, err)
	}

	_, err = parseSchedulePayload(taskdomain.ScheduleTypeDaily, []byte(`{"interval":"x"}`))
	if err == nil {
		t.Fatalf("expected parse error for invalid payload")
	}

	if _, err := parseDateOnly("bad"); err == nil {
		t.Fatalf("expected invalid date parse error")
	}

	unknown := schedulePayloadToJSON(taskdomain.ScheduleType("x"), taskdomain.SchedulePayload{})
	if unknown == nil {
		t.Fatalf("expected fallback payload map")
	}

	dto := newScheduleDTO(&taskdomain.Schedule{
		ID:              1,
		BaseTitle:       "x",
		BaseDescription: "y",
		StatusTemplate:  taskdomain.StatusNew,
		Type:            taskdomain.ScheduleTypeDaily,
		Payload: taskdomain.SchedulePayload{
			Daily: &taskdomain.DailySchedule{Interval: 1},
		},
		StartDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	if dto.ScheduleType != taskdomain.ScheduleTypeDaily {
		t.Fatalf("unexpected dto conversion")
	}
}

type scheduleUsecaseMock struct {
	createErr error
	updateErr error
	deleteErr error
	getErr    error
	listErr   error
}

func (m *scheduleUsecaseMock) CreateSchedule(_ context.Context, input taskusecase.CreateScheduleInput) (*taskdomain.Schedule, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	return &taskdomain.Schedule{
		ID:              1,
		BaseTitle:       input.BaseTitle,
		BaseDescription: input.BaseDescription,
		StatusTemplate:  input.StatusTemplate,
		Type:            input.Type,
		Payload:         input.Payload,
		StartDate:       input.StartDate,
		EndDate:         input.EndDate,
		IsActive:        input.IsActive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func (m *scheduleUsecaseMock) UpdateSchedule(_ context.Context, id int64, input taskusecase.UpdateScheduleInput) (*taskdomain.Schedule, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	return &taskdomain.Schedule{
		ID:              id,
		BaseTitle:       input.BaseTitle,
		BaseDescription: input.BaseDescription,
		StatusTemplate:  input.StatusTemplate,
		Type:            input.Type,
		Payload:         input.Payload,
		StartDate:       input.StartDate,
		EndDate:         input.EndDate,
		IsActive:        input.IsActive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func (m *scheduleUsecaseMock) DeactivateSchedule(_ context.Context, _ int64) error {
	return m.deleteErr
}

func (m *scheduleUsecaseMock) GetScheduleByID(_ context.Context, id int64) (*taskdomain.Schedule, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	return &taskdomain.Schedule{
		ID:              id,
		BaseTitle:       "x",
		BaseDescription: "y",
		StatusTemplate:  taskdomain.StatusNew,
		Type:            taskdomain.ScheduleTypeDaily,
		Payload: taskdomain.SchedulePayload{
			Daily: &taskdomain.DailySchedule{Interval: 1},
		},
		StartDate: now,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (m *scheduleUsecaseMock) ListSchedules(_ context.Context) ([]taskdomain.Schedule, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	return []taskdomain.Schedule{{
		ID:              1,
		BaseTitle:       "x",
		BaseDescription: "y",
		StatusTemplate:  taskdomain.StatusNew,
		Type:            taskdomain.ScheduleTypeDaily,
		Payload: taskdomain.SchedulePayload{
			Daily: &taskdomain.DailySchedule{Interval: 1},
		},
		StartDate: now,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}}, nil
}

func (m *scheduleUsecaseMock) ComputePlannedDates(_ context.Context, _ int64, _, _ time.Time) ([]time.Time, error) {
	return nil, nil
}

func (m *scheduleUsecaseMock) GenerateTasks(_ context.Context, _, _ time.Time) (int, error) {
	return 0, nil
}

func TestScheduleHandlerErrorMappings(t *testing.T) {
	usecase := &scheduleUsecaseMock{
		updateErr: taskusecase.ErrInvalidSchedule,
		deleteErr: taskusecase.ErrScheduleNotFound,
		getErr:    taskusecase.ErrScheduleNotFound,
	}
	handler := NewScheduleHandler(usecase)

	getReq := mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/api/v1/schedules/1", nil), map[string]string{"id": "1"})
	getRec := httptest.NewRecorder()
	handler.GetByID(getRec, getReq)
	if getRec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", getRec.Code)
	}

	updateReq := mux.SetURLVars(httptest.NewRequest(http.MethodPut, "/api/v1/schedules/1", bytes.NewBufferString(`{
		"base_title":"u","base_description":"d","status_template":"new","schedule_type":"monthly_day",
		"schedule_payload":{"day":15},"start_date":"2026-05-01","is_active":true
	}`)), map[string]string{"id": "1"})
	updateRec := httptest.NewRecorder()
	handler.Update(updateRec, updateReq)
	if updateRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", updateRec.Code)
	}

	deleteReq := mux.SetURLVars(httptest.NewRequest(http.MethodDelete, "/api/v1/schedules/1", nil), map[string]string{"id": "1"})
	deleteRec := httptest.NewRecorder()
	handler.Delete(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", deleteRec.Code)
	}
}

func TestScheduleHandlerListInternalError(t *testing.T) {
	handler := NewScheduleHandler(&scheduleUsecaseMock{listErr: errors.New("boom")})
	rec := httptest.NewRecorder()
	handler.List(rec, httptest.NewRequest(http.MethodGet, "/api/v1/schedules", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestScheduleHandlerUpdateBadIDAndJSON(t *testing.T) {
	handler := NewScheduleHandler(&scheduleUsecaseMock{})

	updateReq := mux.SetURLVars(httptest.NewRequest(http.MethodPut, "/api/v1/schedules/x", bytes.NewBufferString(`{}`)), map[string]string{"id": "x"})
	updateRec := httptest.NewRecorder()
	handler.Update(updateRec, updateReq)
	if updateRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad id, got %d", updateRec.Code)
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/schedules", bytes.NewBufferString(`{"base_title":"x","unknown":1}`))
	createRec := httptest.NewRecorder()
	handler.Create(createRec, createReq)
	if createRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad create json, got %d", createRec.Code)
	}
}
