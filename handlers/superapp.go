package handlers

import (
	"database/sql"
	"log"
	"net/http"
)

// SuperAppHomeHandler displays the new central navigation hub
func SuperAppHomeHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Check if user is logged in
	userID, err := GetUserIDFromCookie(r, db)
	isLoggedIn := err == nil
	username := ""

	if isLoggedIn {
		db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	}

	// Get quick stats
	var userCount, postCount, transitCount int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&postCount)
	db.QueryRow("SELECT COUNT(*) FROM nexus_transit_posts WHERE status = 'open'").Scan(&transitCount)

	data := map[string]interface{}{
		"IsLoggedIn": isLoggedIn,
		"Username":   username,
		"Stats": map[string]int{
			"Users":        userCount,
			"Posts":       postCount,
			"TransitPosts": transitCount,
		},
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	err = tmpl.ExecuteTemplate(w, "superapp_home.html", data)
	if err != nil {
		log.Println("Error rendering superapp home page:", err)
		ErrorHandler(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
