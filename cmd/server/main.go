package main

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"running-tracker/internal/handlers"
	"running-tracker/internal/logger"
	"running-tracker/internal/metrics"
	"running-tracker/internal/storage"
)

func main() {
	log := logger.New()

	store := storage.NewMemoryStore()
	metrics.WorkoutsActive.Set(float64(store.Count()))
	log.Info("storage initialized", "workouts", store.Count())

	h := handlers.New(store, log)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(logger.HTTPLogger(log))

	r.Handle("/metrics", promhttp.Handler())
	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "openapi.yaml")
	})
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/openapi.yaml"),
	))

	r.Group(func(r chi.Router) {
		r.Use(metrics.Middleware)
		r.Get("/workouts", h.ListWorkouts)
		r.Post("/workouts", h.CreateWorkout)
		r.Get("/workouts/{id}", h.GetWorkout)
		r.Put("/workouts/{id}", h.UpdateWorkout)
		r.Delete("/workouts/{id}", h.DeleteWorkout)
		r.Get("/records", h.ListRecords)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Info("server starting",
		"port", port,
		"swagger", "http://localhost:"+port+"/swagger/index.html",
		"metrics", "http://localhost:"+port+"/metrics",
	)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Error("server stopped", "error", err.Error())
		os.Exit(1)
	}
}
