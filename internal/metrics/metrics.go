package metrics

import "github.com/prometheus/client_golang/prometheus"

// HTTP metrics
var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests by method, path and status code",
		},
		[]string{"method", "path", "status_code"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 10},
		},
		[]string{"method", "path"},
	)

	HTTPRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)
)

// Продуктовые метрики
var (
	WorkoutsCreatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "workouts_created_total",
			Help: "Total number of workouts ever created",
		},
	)

	WorkoutsDeletedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "workouts_deleted_total",
			Help: "Total number of workouts ever deleted",
		},
	)

	WorkoutsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "workouts_active_total",
			Help: "Current number of workouts in the store",
		},
	)

	// Показывает самые популярные дистанции
	WorkoutDistanceKm = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "workout_distance_km",
			Help:    "Distribution of workout distances in kilometres",
			Buckets: []float64{1, 3, 5, 10, 15, 21.1, 42.2, 100},
		},
	)

	// Сколько всего раз были побиты рекорды
	PersonalRecordsBrokenTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "personal_records_broken_total",
			Help: "Number of personal records broken, labelled by distance name",
		},
		[]string{"distance_name"},
	)
)

func init() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		HTTPRequestsInFlight,
		WorkoutsCreatedTotal,
		WorkoutsDeletedTotal,
		WorkoutsActive,
		WorkoutDistanceKm,
		PersonalRecordsBrokenTotal,
	)
}
