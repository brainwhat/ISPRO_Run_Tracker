package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"running-tracker/internal/models"
	"running-tracker/internal/storage"
)

type Handler struct {
	store *storage.MemoryStore
}

func New(store *storage.MemoryStore) *Handler {
	return &Handler{store: store}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (h *Handler) ListWorkouts(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	writeJSON(w, http.StatusOK, h.store.List(limit, offset))
}

func (h *Handler) GetWorkout(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	wk, err := h.store.Get(id)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, "workout not found")
		return
	}
	writeJSON(w, http.StatusOK, wk)
}

func (h *Handler) CreateWorkout(w http.ResponseWriter, r *http.Request) {
	var inp models.WorkoutInput
	if err := json.NewDecoder(r.Body).Decode(&inp); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if inp.DistanceKm <= 0 || inp.DurationMin <= 0 {
		writeError(w, http.StatusBadRequest, "distance_km and duration_min are required and must be positive")
		return
	}
	wk, newRecords := h.store.Create(inp)
	writeJSON(w, http.StatusCreated, models.CreateWorkoutResponse{
		Workout:    wk,
		NewRecords: newRecords,
	})
}

func (h *Handler) UpdateWorkout(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var inp models.WorkoutInput
	if err := json.NewDecoder(r.Body).Decode(&inp); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if inp.DistanceKm <= 0 || inp.DurationMin <= 0 {
		writeError(w, http.StatusBadRequest, "distance_km and duration_min are required and must be positive")
		return
	}
	wk, err := h.store.Update(id, inp)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, "workout not found")
		return
	}
	writeJSON(w, http.StatusOK, wk)
}

func (h *Handler) DeleteWorkout(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.store.Delete(id); errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, "workout not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListRecords(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.store.ListRecords())
}
