package transporthttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	swaggerdocs "example.com/taskservice/internal/transport/http/docs"
	httphandlers "example.com/taskservice/internal/transport/http/handlers"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

func TestNewRouterRegistersRoutes(t *testing.T) {
	taskHandler := httphandlers.NewTaskHandler(&taskUsecaseNoop{})
	scheduleHandler := httphandlers.NewScheduleHandler(&scheduleUsecaseNoop{})
	docsHandler := swaggerdocs.NewHandler()

	router := NewRouter(taskHandler, scheduleHandler, docsHandler)

	cases := []struct {
		method string
		url    string
	}{
		{http.MethodGet, "/swagger/openapi.json"},
		{http.MethodGet, "/swagger/"},
		{http.MethodGet, "/swagger"},
		{http.MethodGet, "/api/v1/tasks"},
		{http.MethodPost, "/api/v1/tasks"},
		{http.MethodGet, "/api/v1/tasks/1"},
		{http.MethodGet, "/api/v1/schedules"},
		{http.MethodPost, "/api/v1/schedules"},
		{http.MethodGet, "/api/v1/schedules/1"},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.url, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code == http.StatusMethodNotAllowed {
			t.Fatalf("route not registered: %s %s -> %d", tc.method, tc.url, rec.Code)
		}
	}
}

type taskUsecaseNoop struct{}

func (t *taskUsecaseNoop) Create(ctx context.Context, input taskusecase.CreateInput) (*taskdomain.Task, error) {
	return nil, taskusecase.ErrInvalidInput
}
func (t *taskUsecaseNoop) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	return nil, taskdomain.ErrNotFound
}
func (t *taskUsecaseNoop) Update(ctx context.Context, id int64, input taskusecase.UpdateInput) (*taskdomain.Task, error) {
	return nil, taskusecase.ErrInvalidInput
}
func (t *taskUsecaseNoop) Delete(ctx context.Context, id int64) error { return taskdomain.ErrNotFound }
func (t *taskUsecaseNoop) List(ctx context.Context, filter taskusecase.ListFilter) ([]taskdomain.Task, error) {
	return []taskdomain.Task{}, nil
}

type scheduleUsecaseNoop struct{}

func (s *scheduleUsecaseNoop) CreateSchedule(ctx context.Context, input taskusecase.CreateScheduleInput) (*taskdomain.Schedule, error) {
	return nil, taskusecase.ErrInvalidSchedule
}
func (s *scheduleUsecaseNoop) UpdateSchedule(ctx context.Context, id int64, input taskusecase.UpdateScheduleInput) (*taskdomain.Schedule, error) {
	return nil, taskusecase.ErrInvalidSchedule
}
func (s *scheduleUsecaseNoop) DeactivateSchedule(ctx context.Context, id int64) error {
	return taskusecase.ErrScheduleNotFound
}
func (s *scheduleUsecaseNoop) GetScheduleByID(ctx context.Context, id int64) (*taskdomain.Schedule, error) {
	return nil, taskusecase.ErrScheduleNotFound
}
func (s *scheduleUsecaseNoop) ListSchedules(ctx context.Context) ([]taskdomain.Schedule, error) {
	return []taskdomain.Schedule{}, nil
}
func (s *scheduleUsecaseNoop) ComputePlannedDates(ctx context.Context, id int64, from, to time.Time) ([]time.Time, error) {
	return nil, nil
}
func (s *scheduleUsecaseNoop) GenerateTasks(ctx context.Context, from, to time.Time) (int, error) {
	return 0, nil
}
