package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"running-tracker/internal/handlers"
	"running-tracker/internal/storage"
)

func main() {
	store := storage.NewMemoryStore()
	h := handlers.New(store)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/workouts", h.ListWorkouts)
	r.Post("/workouts", h.CreateWorkout)
	r.Get("/workouts/{id}", h.GetWorkout)
	r.Put("/workouts/{id}", h.UpdateWorkout)
	r.Delete("/workouts/{id}", h.DeleteWorkout)

	r.Get("/records", h.ListRecords)

	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "openapi.yaml")
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/openapi.yaml"),
	))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server:    http://localhost:%s", port)
	log.Printf("Swagger:   http://localhost:%s/swagger/index.html", port)
	log.Printf("OpenAPI:   http://localhost:%s/openapi.yaml", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
