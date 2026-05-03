package models

import "time"

type WorkoutInput struct {
	Name         string  `json:"name,omitempty"`
	Description  string  `json:"description,omitempty"`
	DistanceKm   float64 `json:"distance_km"`
	DurationMin  float64 `json:"duration_min"`
	AvgHeartRate int     `json:"avg_heart_rate,omitempty"`
	Calories     int     `json:"calories,omitempty"`
	Date         string  `json:"date,omitempty"`
}

type Workout struct {
	ID           int       `json:"id"`
	Name         string    `json:"name,omitempty"`
	Description  string    `json:"description,omitempty"`
	DistanceKm   float64   `json:"distance_km"`
	DurationMin  float64   `json:"duration_min"`
	PaceMinPerKm float64   `json:"pace_min_per_km"`
	AvgHeartRate int       `json:"avg_heart_rate,omitempty"`
	Calories     int       `json:"calories,omitempty"`
	Date         string    `json:"date,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
