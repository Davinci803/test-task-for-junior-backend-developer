package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	taskusecase "example.com/taskservice/internal/usecase/task"
)

type ScheduleHandler struct {
	usecase taskusecase.ScheduleUsecase
}

func NewScheduleHandler(usecase taskusecase.ScheduleUsecase) *ScheduleHandler {
	return &ScheduleHandler{usecase: usecase}
}

func (h *ScheduleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req scheduleMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input, err := req.toCreateInput()
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	created, err := h.usecase.CreateSchedule(r.Context(), input)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newScheduleDTO(created))
}

func (h *ScheduleHandler) List(w http.ResponseWriter, r *http.Request) {
	schedules, err := h.usecase.ListSchedules(r.Context())
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	response := make([]scheduleDTO, 0, len(schedules))
	for i := range schedules {
		response = append(response, newScheduleDTO(&schedules[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *ScheduleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getScheduleIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	schedule, err := h.usecase.GetScheduleByID(r.Context(), id)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newScheduleDTO(schedule))
}

func (h *ScheduleHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getScheduleIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req scheduleMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input, err := req.toUpdateInput()
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updated, err := h.usecase.UpdateSchedule(r.Context(), id, input)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newScheduleDTO(updated))
}

func (h *ScheduleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getScheduleIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.DeactivateSchedule(r.Context(), id); err != nil {
		writeUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getScheduleIDFromRequest(r *http.Request) (int64, error) {
	rawID := mux.Vars(r)["id"]
	if rawID == "" {
		return 0, errors.New("missing schedule id")
	}

	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid schedule id")
	}

	return id, nil
}
