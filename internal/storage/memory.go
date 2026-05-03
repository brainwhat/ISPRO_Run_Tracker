package storage

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"running-tracker/internal/models"
)

var ErrNotFound = errors.New("workout not found")

type MemoryStore struct {
	mu       sync.RWMutex
	workouts []models.Workout
	records  map[int]*models.Record // keyed by distance in meters
	nextID   int
}

func NewMemoryStore() *MemoryStore {
	s := &MemoryStore{
		nextID:  1,
		records: make(map[int]*models.Record),
	}
	s.seed()
	return s
}

func (s *MemoryStore) seed() {
	samples := []models.WorkoutInput{
		{Name: "Easy 5K", Description: "Лёгкий восстановительный бег", DistanceKm: 5.0, DurationMin: 30.0, AvgHeartRate: 130, Calories: 270, Date: "2026-01-08"},
		{Name: "Tempo 10K", Description: "Темповый бег в соревновательном темпе", DistanceKm: 10.0, DurationMin: 50.0, AvgHeartRate: 162, Calories: 550, Date: "2026-01-15"},
		{Name: "Длинная 15K", Description: "Длинная воскресная пробежка", DistanceKm: 15.0, DurationMin: 90.0, AvgHeartRate: 148, Calories: 820, Date: "2026-01-19"},
		{Name: "Полумарафон", Description: "Тестовый полумарафон", DistanceKm: 21.1, DurationMin: 115.0, AvgHeartRate: 158, Calories: 1150, Date: "2026-02-02"},
		{Name: "Восстановление 3K", Description: "Очень лёгкий бег после соревнований", DistanceKm: 3.0, DurationMin: 20.0, AvgHeartRate: 118, Calories: 155, Date: "2026-02-04"},
		{Name: "Интервалы 8K", Description: "400м × 10 с отдыхом", DistanceKm: 8.0, DurationMin: 38.0, AvgHeartRate: 174, Calories: 445, Date: "2026-02-11"},
		{Name: "Утренняя 7K", Description: "Ранний бег перед работой", DistanceKm: 7.0, DurationMin: 40.0, AvgHeartRate: 145, Calories: 375, Date: "2026-02-17"},
		{Name: "Марафон", Description: "Городской марафон", DistanceKm: 42.195, DurationMin: 255.0, AvgHeartRate: 162, Calories: 2400, Date: "2026-03-02"},
		{Name: "Прогулочная 4K", Description: "Разминка перед силовой", DistanceKm: 4.0, DurationMin: 26.0, AvgHeartRate: 125, Calories: 210, Date: "2026-03-10"},
		{Name: "Горный 12K", Description: "Бег с набором высоты", DistanceKm: 12.0, DurationMin: 80.0, AvgHeartRate: 165, Calories: 720, Date: "2026-03-22"},
	}
	for _, inp := range samples {
		s.insert(inp)
	}
}

func (s *MemoryStore) insert(inp models.WorkoutInput) models.Workout {
	w := models.Workout{
		ID:           s.nextID,
		Name:         inp.Name,
		Description:  inp.Description,
		DistanceKm:   inp.DistanceKm,
		DurationMin:  inp.DurationMin,
		PaceMinPerKm: inp.DurationMin / inp.DistanceKm,
		AvgHeartRate: inp.AvgHeartRate,
		Calories:     inp.Calories,
		Date:         inp.Date,
		CreatedAt:    time.Now(),
	}
	s.nextID++
	s.workouts = append(s.workouts, w)
	return w
}

// checkRecords checks all standard distances against the workout and updates personal records.
// Must be called while holding s.mu (write lock).
func (s *MemoryStore) checkRecords(w models.Workout) []models.NewRecordNotification {
	distanceM := w.DistanceKm * 1000
	pace := w.DurationMin / w.DistanceKm // min/km

	var notifications []models.NewRecordNotification

	for _, std := range models.StandardDistances {
		targetM := float64(std.Meters)

		// Workout must be at least as long as the target, but no more than 10% longer.
		if distanceM < targetM || distanceM > targetM*1.1 {
			continue
		}

		targetKm := targetM / 1000.0
		projectedMin := round2(pace * targetKm)
		projectedPace := round2(pace)

		current, exists := s.records[std.Meters]
		if exists && projectedMin >= current.TimeMin {
			continue
		}

		var prev *models.Record
		if exists {
			cp := *current
			prev = &cp
		}

		s.records[std.Meters] = &models.Record{
			DistanceM:    std.Meters,
			DistanceName: std.Name,
			TimeMin:      projectedMin,
			PaceMinPerKm: projectedPace,
			WorkoutID:    w.ID,
			WorkoutDate:  w.Date,
			SetAt:        time.Now(),
		}

		var msg string
		if prev == nil {
			msg = fmt.Sprintf("Первый рекорд на %s: %s (темп %s/км)!", std.Name, formatDuration(projectedMin), formatDuration(projectedPace))
		} else {
			msg = fmt.Sprintf("Новый рекорд на %s: %s → %s (темп %s/км)!", std.Name, formatDuration(prev.TimeMin), formatDuration(projectedMin), formatDuration(projectedPace))
		}

		notifications = append(notifications, models.NewRecordNotification{
			DistanceM:      std.Meters,
			DistanceName:   std.Name,
			TimeMin:        projectedMin,
			PaceMinPerKm:   projectedPace,
			PreviousRecord: prev,
			Message:        msg,
		})
	}

	return notifications
}

// formatDuration formats minutes as M:SS or H:MM:SS.
func formatDuration(minutes float64) string {
	totalSec := int(math.Round(minutes * 60))
	h := totalSec / 3600
	m := (totalSec % 3600) / 60
	sec := totalSec % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, sec)
	}
	return fmt.Sprintf("%d:%02d", m, sec)
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func (s *MemoryStore) List(limit, offset int) []models.Workout {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if offset >= len(s.workouts) {
		return []models.Workout{}
	}
	end := offset + limit
	if end > len(s.workouts) {
		end = len(s.workouts)
	}
	return s.workouts[offset:end]
}

func (s *MemoryStore) Get(id int) (models.Workout, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, w := range s.workouts {
		if w.ID == id {
			return w, nil
		}
	}
	return models.Workout{}, ErrNotFound
}

func (s *MemoryStore) Create(inp models.WorkoutInput) (models.Workout, []models.NewRecordNotification) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w := s.insert(inp)
	notifs := s.checkRecords(w)
	if notifs == nil {
		notifs = []models.NewRecordNotification{}
	}
	return w, notifs
}

func (s *MemoryStore) Update(id int, inp models.WorkoutInput) (models.Workout, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, w := range s.workouts {
		if w.ID == id {
			s.workouts[i].Name = inp.Name
			s.workouts[i].Description = inp.Description
			s.workouts[i].DistanceKm = inp.DistanceKm
			s.workouts[i].DurationMin = inp.DurationMin
			s.workouts[i].PaceMinPerKm = inp.DurationMin / inp.DistanceKm
			s.workouts[i].AvgHeartRate = inp.AvgHeartRate
			s.workouts[i].Calories = inp.Calories
			s.workouts[i].Date = inp.Date
			return s.workouts[i], nil
		}
	}
	return models.Workout{}, ErrNotFound
}

func (s *MemoryStore) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, w := range s.workouts {
		if w.ID == id {
			s.workouts = append(s.workouts[:i], s.workouts[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (s *MemoryStore) ListRecords() []models.Record {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]models.Record, 0, len(s.records))
	// Return in standard distance order
	for _, std := range models.StandardDistances {
		if r, ok := s.records[std.Meters]; ok {
			result = append(result, *r)
		}
	}
	return result
}
