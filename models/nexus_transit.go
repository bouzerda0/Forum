package models

import "time"

// TransitPost represents a carpooling/ride-sharing offer
type TransitPost struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	Username         string    `json:"username"`
	Direction        string    `json:"direction"` // 'aller' or 'retour'
	DepartureTime    time.Time `json:"departure_time"`
	DepartureLocation string  `json:"departure_location"`
	TotalSeats       int       `json:"total_seats"` // Always 4
	AvailableSeats   int       `json:"available_seats"`
	Status           string    `json:"status"` // 'open', 'full', 'cancelled'
	Notes            string    `json:"notes"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TransitBooking represents a passenger booking on a ride
type TransitBooking struct {
	ID           string    `json:"id"`
	PostID       string    `json:"post_id"`
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	BookingOrder int       `json:"booking_order"` // 1-4 for FIFO
	Status       string    `json:"status"` // 'confirmed', 'cancelled'
	CreatedAt    time.Time `json:"created_at"`
}

// TransitPostWithBookings includes passengers for a specific ride
type TransitPostWithBookings struct {
	TransitPost
	Passengers []TransitBooking `json:"passengers"`
}
