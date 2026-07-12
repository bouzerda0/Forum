package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"forum/models"
)

// NexusTransitHomeHandler displays all transit posts
func NexusTransitHomeHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	userID, err := GetUserIDFromCookie(r, db)
	isLoggedIn := err == nil
	username := ""

	if isLoggedIn {
		db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	}

	// Get filter params
	direction := r.URL.Query().Get("direction")
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "open"
	}

	// Build query
	var posts []models.TransitPost
	query := `
		SELECT t.id, t.user_id, u.username, t.direction, t.departure_time,
			   t.departure_location, t.total_seats, t.available_seats, t.status, t.notes, t.created_at
		FROM nexus_transit_posts t
		JOIN users u ON t.user_id = u.id
		WHERE t.status != 'cancelled'
	`
	args := []interface{}{}

	if direction != "" {
		query += " AND t.direction = ?"
		args = append(args, direction)
	}

	if status != "all" {
		query += " AND t.status = ?"
		args = append(args, status)
	}

	query += " ORDER BY t.departure_time ASC"

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Println("Error fetching transit posts:", err)
		ErrorHandler(w, "Failed to load rides", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var p models.TransitPost
		err := rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Direction, &p.DepartureTime,
			&p.DepartureLocation, &p.TotalSeats, &p.AvailableSeats, &p.Status, &p.Notes, &p.CreatedAt)
		if err != nil {
			log.Println("Error scanning transit post:", err)
			continue
		}
		posts = append(posts, p)
	}

	data := map[string]interface{}{
		"IsLoggedIn": isLoggedIn,
		"Username":   username,
		"Posts":      posts,
		"Direction":  direction,
		"Status":     status,
	}

	tmpl.ExecuteTemplate(w, "nexus_transit_home.html", data)
}

// NexusTransitCreateHandler handles creating a new ride
func NexusTransitCreateHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	userID, err := GetUserIDFromCookie(r, db)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		data := map[string]interface{}{
			"Error": "",
		}
		tmpl.ExecuteTemplate(w, "nexus_transit_create.html", data)
		return
	}

	if r.Method == http.MethodPost {
		direction := r.FormValue("direction")
		departureTimeStr := r.FormValue("departure_time")
		departureLocation := r.FormValue("departure_location")
		notes := r.FormValue("notes")

		// Parse departure time
		departureTime, err := time.Parse("2006-01-02T15:04", departureTimeStr)
		if err != nil {
			data := map[string]interface{}{
				"Error": "Invalid departure time format",
			}
			tmpl.ExecuteTemplate(w, "nexus_transit_create.html", data)
			return
		}

		// Validate direction
		if direction != "aller" && direction != "retour" {
			data := map[string]interface{}{
				"Error": "Invalid direction",
			}
			tmpl.ExecuteTemplate(w, "nexus_transit_create.html", data)
			return
		}

		// Insert into database
		_, err = db.Exec(`
			INSERT INTO nexus_transit_posts
			(user_id, direction, departure_time, departure_location, total_seats, available_seats, notes)
			VALUES (?, ?, ?, ?, 4, 4, ?)
		`, userID, direction, departureTime, departureLocation, notes)

		if err != nil {
			log.Println("Error creating transit post:", err)
			data := map[string]interface{}{
				"Error": "Failed to create ride",
			}
			tmpl.ExecuteTemplate(w, "nexus_transit_create.html", data)
			return
		}

		http.Redirect(w, r, "/nexus-transit", http.StatusSeeOther)
	}
}

// NexusTransitDetailHandler shows a single ride with passengers
func NexusTransitDetailHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	postID := r.URL.Query().Get("id")
	if postID == "" {
		ErrorHandler(w, "Ride not found", http.StatusNotFound)
		return
	}

	userID, err := GetUserIDFromCookie(r, db)
	isLoggedIn := err == nil
	username := ""

	if isLoggedIn {
		db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	}

	// Get the post
	var p models.TransitPost
	err = db.QueryRow(`
		SELECT t.id, t.user_id, u.username, t.direction, t.departure_time,
			   t.departure_location, t.total_seats, t.available_seats, t.status, t.notes, t.created_at
		FROM nexus_transit_posts t
		JOIN users u ON t.user_id = u.id
		WHERE t.id = ?
	`, postID).Scan(&p.ID, &p.UserID, &p.Username, &p.Direction, &p.DepartureTime,
		&p.DepartureLocation, &p.TotalSeats, &p.AvailableSeats, &p.Status, &p.Notes, &p.CreatedAt)

	if err != nil {
		ErrorHandler(w, "Ride not found", http.StatusNotFound)
		return
	}

	// Get passengers
	rows, err := db.Query(`
		SELECT b.id, b.post_id, b.user_id, u.username, b.booking_order, b.status, b.created_at
		FROM nexus_transit_bookings b
		JOIN users u ON b.user_id = u.id
		WHERE b.post_id = ? AND b.status = 'confirmed'
		ORDER BY b.booking_order ASC
	`, postID)

	var passengers []models.TransitBooking
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var b models.TransitBooking
			rows.Scan(&b.ID, &b.PostID, &b.UserID, &b.Username, &b.BookingOrder, &b.Status, &b.CreatedAt)
			passengers = append(passengers, b)
		}
	}

	// Check if current user has booked
	hasBooked := false
	if isLoggedIn {
		var count int
		db.QueryRow("SELECT COUNT(*) FROM nexus_transit_bookings WHERE post_id = ? AND user_id = ? AND status = 'confirmed'", postID, userID).Scan(&count)
		hasBooked = count > 0
	}

	data := map[string]interface{}{
		"IsLoggedIn":  isLoggedIn,
		"Username":    username,
		"Post":        p,
		"Passengers":  passengers,
		"HasBooked":   hasBooked,
		"IsOwner":     isLoggedIn && p.UserID == fmt.Sprintf("%d", userID),
	}

	tmpl.ExecuteTemplate(w, "nexus_transit_detail.html", data)
}

// NexusTransitBookHandler handles booking a seat
func NexusTransitBookHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/nexus-transit", http.StatusSeeOther)
		return
	}

	userID, err := GetUserIDFromCookie(r, db)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	postID := r.FormValue("post_id")

	// Check if user already booked this ride
	var existingCount int
	db.QueryRow("SELECT COUNT(*) FROM nexus_transit_bookings WHERE post_id = ? AND user_id = ? AND status = 'confirmed'", postID, userID).Scan(&existingCount)
	if existingCount > 0 {
		http.Redirect(w, r, "/nexus-transit/detail?id="+postID, http.StatusSeeOther)
		return
	}

	// Get current booking count
	var currentCount int
	db.QueryRow("SELECT COUNT(*) FROM nexus_transit_bookings WHERE post_id = ? AND status = 'confirmed'", postID).Scan(&currentCount)

	if currentCount >= 4 {
		// Ride is full - should not happen if UI is correct
		http.Redirect(w, r, "/nexus-transit/detail?id="+postID, http.StatusSeeOther)
		return
	}

	// Calculate booking order (1-4)
	bookingOrder := currentCount + 1

	// Insert booking
	_, err = db.Exec(`
		INSERT INTO nexus_transit_bookings (post_id, user_id, booking_order)
		VALUES (?, ?, ?)
	`, postID, userID, bookingOrder)

	if err != nil {
		log.Println("Error booking ride:", err)
		ErrorHandler(w, "Failed to book ride", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/nexus-transit/detail?id="+postID, http.StatusSeeOther)
}

// NexusTransitCancelBookingHandler handles cancelling a booking
func NexusTransitCancelBookingHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/nexus-transit", http.StatusSeeOther)
		return
	}

	userID, err := GetUserIDFromCookie(r, db)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	postID := r.FormValue("post_id")

	// Update booking status to cancelled
	_, err = db.Exec("UPDATE nexus_transit_bookings SET status = 'cancelled' WHERE post_id = ? AND user_id = ? AND status = 'confirmed'", postID, userID)

	if err != nil {
		log.Println("Error cancelling booking:", err)
		ErrorHandler(w, "Failed to cancel booking", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/nexus-transit/detail?id="+postID, http.StatusSeeOther)
}

// API endpoint to get available seats count
func NexusTransitAPIStatsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	postID := r.URL.Query().Get("post_id")

	var availableSeats int
	var status string
	err := db.QueryRow("SELECT available_seats, status FROM nexus_transit_posts WHERE id = ?", postID).Scan(&availableSeats, &status)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Ride not found",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"available_seats": availableSeats,
		"status": status,
	})
}
