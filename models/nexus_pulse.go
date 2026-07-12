package models

import "time"

// PulseFoodPost represents a food order coordination
type PulseFoodPost struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"user_id"`
	Username           string    `json:"username"`
	RestaurantName     string    `json:"restaurant_name"`
	RestaurantLink     string    `json:"restaurant_link"`
	Description        string    `json:"description"`
	DeadlineTime       time.Time `json:"deadline_time"`
	MaxParticipants    int       `json:"max_participants"` // Always 5
	CurrentParticipants int      `json:"current_participants"`
	Status             string    `json:"status"`
	DeliveryFeeSplit   bool      `json:"delivery_fee_split"`
	Notes              string    `json:"notes"`
	CreatedAt          time.Time `json:"created_at"`
}

// PulseFoodJoin represents a user joining a food order
type PulseFoodJoin struct {
	ID            string    `json:"id"`
	PostID        string    `json:"post_id"`
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	OrderDetails  string    `json:"order_details"`
	EstimatedCost float64   `json:"estimated_cost"`
	CreatedAt     time.Time `json:"created_at"`
}

// PulseEvent represents a campus event
type PulseEvent struct {
	ID               string     `json:"id"`
	UserID           string    `json:"user_id"`
	Username         string    `json:"username"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	EventType        string    `json:"event_type"`
	EventDate        time.Time `json:"event_date"`
	Location         string    `json:"location"`
	MaxAttendees     int       `json:"max_attendees"`
	CurrentAttendees int       `json:"current_attendees"`
	Status           string    `json:"status"`
	ImageURL         string    `json:"image_url"`
	CreatedAt        time.Time `json:"created_at"`
}

// PulseEventAttendee represents attendance status for an event
type PulseEventAttendee struct {
	ID        string    `json:"id"`
	EventID   string    `json:"event_id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
