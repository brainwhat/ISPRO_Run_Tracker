package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"running-tracker/internal/handlers"
	"running-tracker/internal/logger"
	"running-tracker/internal/metrics"
	"running-tracker/internal/storage"
	"running-tracker/internal/telemetry"
)

func tracingMiddleware(next http.Handler) http.Handler {
	tracer := otel.Tracer("running-tracker/http")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), r.Method+" "+r.URL.Path, trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("url.path", r.URL.Path),
		))
		defer span.End()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r.WithContext(ctx))

		path := chi.RouteContext(r.Context()).RoutePattern()
		if path == "" {
			path = r.URL.Path
		}
		status := ww.Status()
		span.SetAttributes(
			attribute.String("http.route", path),
			attribute.Int("http.status_code", status),
		)
		if status >= 500 {
			span.SetStatus(codes.Error, "internal server error")
		}
	})
}

func main() {
	log := logger.New()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	otelEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelEndpoint == "" {
		otelEndpoint = "127.0.0.1:14317"
	}
	shutdownTracing, err := telemetry.InitTracing(ctx, "running-tracker", otelEndpoint)
	if err != nil {
		log.Error("failed to init tracing", "error", err.Error())
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTracing(shutdownCtx); err != nil {
			log.Warn("failed to shutdown tracing", "error", err.Error())
		}
	}()

	store := storage.NewMemoryStore()
	metrics.WorkoutsActive.Set(float64(store.Count()))
	log.Info("storage initialized", "workouts", store.Count())

	h := handlers.New(store, log)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(tracingMiddleware)
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
