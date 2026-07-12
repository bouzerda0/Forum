package models

import "time"

// HabitatPost represents a roommate search post
type HabitatPost struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Username       string    `json:"username"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Location       string    `json:"location"`
	AvailableSpots int       `json:"available_spots"`
	FilledSpots     int       `json:"filled_spots"`
	RentCost       float64   `json:"rent_cost"`
	LifestyleTags  []string  `json:"lifestyle_tags"`
	Rules          string    `json:"rules"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// HabitatRequest represents a roommate request
type HabitatRequest struct {
	ID         string     `json:"id"`
	PostID     string     `json:"post_id"`
	UserID     string     `json:"user_id"`
	Username   string     `json:"username"`
	Message    string     `json:"message"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	ReviewedAt *time.Time `json:"reviewed_at"`
}
