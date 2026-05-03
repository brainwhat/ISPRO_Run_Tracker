package models

import "time"

type StandardDistance struct {
	Meters int
	Name   string
}

var StandardDistances = []StandardDistance{
	{100, "100м"},
	{200, "200м"},
	{400, "400м"},
	{800, "800м"},
	{1000, "1К"},
	{1500, "1500м"},
	{1609, "1 миля"},
	{3000, "3К"},
	{5000, "5К"},
	{10000, "10К"},
	{21097, "Полумарафон"},
	{42195, "Марафон"},
}

type Record struct {
	DistanceM    int       `json:"distance_m"`
	DistanceName string    `json:"distance_name"`
	TimeMin      float64   `json:"time_min"`
	PaceMinPerKm float64   `json:"pace_min_per_km"`
	WorkoutID    int       `json:"workout_id"`
	WorkoutDate  string    `json:"workout_date,omitempty"`
	SetAt        time.Time `json:"set_at"`
}

type NewRecordNotification struct {
	DistanceM      int     `json:"distance_m"`
	DistanceName   string  `json:"distance_name"`
	TimeMin        float64 `json:"time_min"`
	PaceMinPerKm   float64 `json:"pace_min_per_km"`
	PreviousRecord *Record `json:"previous_record,omitempty"`
	Message        string  `json:"message"`
}

type CreateWorkoutResponse struct {
	Workout    Workout                 `json:"workout"`
	NewRecords []NewRecordNotification `json:"new_records"`
}
