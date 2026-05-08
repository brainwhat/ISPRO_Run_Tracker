package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"running-tracker/internal/metrics"
	"running-tracker/internal/models"
	"running-tracker/internal/storage"
)

var tracer = otel.Tracer("running-tracker/handlers")

type Handler struct {
	store *storage.MemoryStore
	log   *slog.Logger
}

func New(store *storage.MemoryStore, log *slog.Logger) *Handler {
	return &Handler{store: store, log: log}
}

func (h *Handler) reqLog(r *http.Request) *slog.Logger {
	return h.log.With(slog.String("request_id", middleware.GetReqID(r.Context())))
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
	items := h.store.List(limit, offset)
	h.reqLog(r).Debug("list workouts", "limit", limit, "offset", offset, "returned", len(items))
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) GetWorkout(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.reqLog(r).Warn("invalid workout id", "raw", chi.URLParam(r, "id"))
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	wk, err := h.store.Get(id)
	if errors.Is(err, storage.ErrNotFound) {
		h.reqLog(r).Warn("workout not found", "id", id)
		writeError(w, http.StatusNotFound, "workout not found")
		return
	}
	writeJSON(w, http.StatusOK, wk)
}

func (h *Handler) CreateWorkout(w http.ResponseWriter, r *http.Request) {
	log := h.reqLog(r)

	var inp models.WorkoutInput
	if err := json.NewDecoder(r.Body).Decode(&inp); err != nil {
		log.Warn("invalid request body", "error", err.Error())
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if inp.DistanceKm <= 0 || inp.DurationMin <= 0 {
		log.Warn("invalid workout input", "distance_km", inp.DistanceKm, "duration_min", inp.DurationMin)
		writeError(w, http.StatusBadRequest, "distance_km and duration_min are required and must be positive")
		return
	}

	ctx, storeSpan := tracer.Start(r.Context(), "storage.create", trace.WithAttributes(
		attribute.String("workout.name", inp.Name),
		attribute.Float64("workout.distance_km", inp.DistanceKm),
		attribute.Float64("workout.duration_min", inp.DurationMin),
	))
	wk, newRecords := h.store.Create(inp)
	storeSpan.SetAttributes(attribute.Int("workout.id", wk.ID))
	storeSpan.End()

	metrics.WorkoutsCreatedTotal.Inc()
	metrics.WorkoutsActive.Inc()
	metrics.WorkoutDistanceKm.Observe(inp.DistanceKm)

	_, recSpan := tracer.Start(ctx, "records.notify", trace.WithAttributes(
		attribute.Int("records.count", len(newRecords)),
	))
	for _, rec := range newRecords {
		metrics.PersonalRecordsBrokenTotal.WithLabelValues(rec.DistanceName).Inc()
		log.Info("personal record broken",
			"distance", rec.DistanceName,
			"time_min", rec.TimeMin,
			"workout_id", wk.ID,
		)
	}
	recSpan.End()

	log.Info("workout created",
		"id", wk.ID,
		"distance_km", wk.DistanceKm,
		"duration_min", wk.DurationMin,
		"new_records", len(newRecords),
	)

	writeJSON(w, http.StatusCreated, models.CreateWorkoutResponse{
		Workout:    wk,
		NewRecords: newRecords,
	})
}

func (h *Handler) UpdateWorkout(w http.ResponseWriter, r *http.Request) {
	log := h.reqLog(r)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Warn("invalid workout id", "raw", chi.URLParam(r, "id"))
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var inp models.WorkoutInput
	if err := json.NewDecoder(r.Body).Decode(&inp); err != nil {
		log.Warn("invalid request body", "error", err.Error())
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if inp.DistanceKm <= 0 || inp.DurationMin <= 0 {
		log.Warn("invalid workout input", "distance_km", inp.DistanceKm, "duration_min", inp.DurationMin)
		writeError(w, http.StatusBadRequest, "distance_km and duration_min are required and must be positive")
		return
	}
	wk, err := h.store.Update(id, inp)
	if errors.Is(err, storage.ErrNotFound) {
		log.Warn("workout not found", "id", id)
		writeError(w, http.StatusNotFound, "workout not found")
		return
	}
	log.Info("workout updated", "id", id, "distance_km", wk.DistanceKm, "duration_min", wk.DurationMin)
	writeJSON(w, http.StatusOK, wk)
}

func (h *Handler) DeleteWorkout(w http.ResponseWriter, r *http.Request) {
	log := h.reqLog(r)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Warn("invalid workout id", "raw", chi.URLParam(r, "id"))
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.store.Delete(id); errors.Is(err, storage.ErrNotFound) {
		log.Warn("workout not found", "id", id)
		writeError(w, http.StatusNotFound, "workout not found")
		return
	}
	metrics.WorkoutsDeletedTotal.Inc()
	metrics.WorkoutsActive.Dec()
	log.Info("workout deleted", "id", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListRecords(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.store.ListRecords())
}
