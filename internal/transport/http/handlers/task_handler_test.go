package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

func TestTaskHandlerCreateAndList(t *testing.T) {
	usecase := &taskUsecaseMock{}
	handler := NewTaskHandler(usecase)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(`{"title":"a","description":"b","status":"new"}`))
	createRec := httptest.NewRecorder()
	handler.Create(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("unexpected create status: %d", createRec.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/tasks?schedule_id=1", nil)
	listRec := httptest.NewRecorder()
	handler.List(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("unexpected list status: %d", listRec.Code)
	}
}

func TestTaskHandlerGetUpdateDeleteErrors(t *testing.T) {
	usecase := &taskUsecaseMock{
		getErr:    taskdomain.ErrNotFound,
		updateErr: taskusecase.ErrInvalidInput,
		deleteErr: taskdomain.ErrNotFound,
	}
	handler := NewTaskHandler(usecase)

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/10", nil)
	getReq = mux.SetURLVars(getReq, map[string]string{"id": "10"})
	getRec := httptest.NewRecorder()
	handler.GetByID(getRec, getReq)
	if getRec.Code != http.StatusNotFound {
		t.Fatalf("unexpected get status: %d", getRec.Code)
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/tasks/10", bytes.NewBufferString(`{"title":"x","description":"y","status":"done"}`))
	updateReq = mux.SetURLVars(updateReq, map[string]string{"id": "10"})
	updateRec := httptest.NewRecorder()
	handler.Update(updateRec, updateReq)
	if updateRec.Code != http.StatusBadRequest {
		t.Fatalf("unexpected update status: %d", updateRec.Code)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/10", nil)
	deleteReq = mux.SetURLVars(deleteReq, map[string]string{"id": "10"})
	deleteRec := httptest.NewRecorder()
	handler.Delete(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusNotFound {
		t.Fatalf("unexpected delete status: %d", deleteRec.Code)
	}
}

func TestTaskHandlerInvalidQueryAndJSON(t *testing.T) {
	usecase := &taskUsecaseMock{}
	handler := NewTaskHandler(usecase)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks?planned_for_from=bad", nil)
	rec := httptest.NewRecorder()
	handler.List(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request for invalid filter, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(`{"title":"a","unknown":1}`))
	rec = httptest.NewRecorder()
	handler.Create(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request for unknown field, got %d", rec.Code)
	}
}

func TestParseTaskListFilter(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks?planned_for_from=2026-05-01&planned_for_to=2026-05-31&schedule_id=9", nil)
	filter, err := parseTaskListFilter(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filter.ScheduleID == nil || *filter.ScheduleID != 9 {
		t.Fatalf("unexpected schedule_id")
	}
}

type taskUsecaseMock struct {
	getErr    error
	updateErr error
	deleteErr error
	listErr   error
}

func (m *taskUsecaseMock) Create(_ context.Context, input taskusecase.CreateInput) (*taskdomain.Task, error) {
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	return &taskdomain.Task{
		ID:          1,
		Title:       input.Title,
		Description: input.Description,
		Status:      input.Status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (m *taskUsecaseMock) GetByID(_ context.Context, _ int64) (*taskdomain.Task, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	return &taskdomain.Task{ID: 1, Title: "x", Status: taskdomain.StatusNew, CreatedAt: now, UpdatedAt: now}, nil
}

func (m *taskUsecaseMock) Update(_ context.Context, id int64, input taskusecase.UpdateInput) (*taskdomain.Task, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	return &taskdomain.Task{ID: id, Title: input.Title, Description: input.Description, Status: input.Status, CreatedAt: now, UpdatedAt: now}, nil
}

func (m *taskUsecaseMock) Delete(_ context.Context, _ int64) error {
	return m.deleteErr
}

func (m *taskUsecaseMock) List(_ context.Context, _ taskusecase.ListFilter) ([]taskdomain.Task, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	sid := int64(1)
	planned := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	return []taskdomain.Task{
		{ID: 1, Title: "x", Status: taskdomain.StatusNew, ScheduleID: &sid, PlannedFor: &planned, CreatedAt: now, UpdatedAt: now},
	}, nil
}

func TestTaskHandlerBadPathIDAndInternalList(t *testing.T) {
	usecase := &taskUsecaseMock{listErr: errors.New("boom")}
	handler := NewTaskHandler(usecase)

	req := mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/api/v1/tasks/x", nil), map[string]string{"id": "x"})
	rec := httptest.NewRecorder()
	handler.GetByID(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad id, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	handler.List(rec, httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for internal list error, got %d", rec.Code)
	}

	updateReq := mux.SetURLVars(httptest.NewRequest(http.MethodPut, "/api/v1/tasks/x", bytes.NewBufferString(`{"title":"x","description":"d","status":"new"}`)), map[string]string{"id": "x"})
	updateRec := httptest.NewRecorder()
	handler.Update(updateRec, updateReq)
	if updateRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for update bad id, got %d", updateRec.Code)
	}

	deleteReq := mux.SetURLVars(httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/x", nil), map[string]string{"id": "x"})
	deleteRec := httptest.NewRecorder()
	handler.Delete(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for delete bad id, got %d", deleteRec.Code)
	}
}

func TestWriteJSONHelpers(t *testing.T) {
	rec := httptest.NewRecorder()
	writeError(rec, http.StatusBadRequest, errors.New("bad"))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unexpected code")
	}
	var payload map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
}

func TestWriteUsecaseErrorMapping(t *testing.T) {
	cases := []struct {
		err  error
		code int
	}{
		{taskdomain.ErrNotFound, http.StatusNotFound},
		{taskdomain.ErrScheduleNotFound, http.StatusNotFound},
		{taskusecase.ErrInvalidInput, http.StatusBadRequest},
		{taskusecase.ErrInvalidSchedule, http.StatusBadRequest},
		{taskusecase.ErrScheduleNotFound, http.StatusNotFound},
		{errors.New("internal"), http.StatusInternalServerError},
	}

	for _, tc := range cases {
		rec := httptest.NewRecorder()
		writeUsecaseError(rec, tc.err)
		if rec.Code != tc.code {
			t.Fatalf("unexpected status for %v: got=%d want=%d", tc.err, rec.Code, tc.code)
		}
	}
}
