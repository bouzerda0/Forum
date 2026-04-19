package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"forum/models"
)

// HomeHandler handles requests to the main forum page
func HomeHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// 1. Check if user is logged in (but don't force it!)
	userID, err := GetUserIDFromCookie(r, db)

	isLoggedIn := (err == nil)
	username := ""

	// 2. If logged in, get the real username from the database
	if isLoggedIn {
		err = db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
		if err != nil {
			log.Println("Error fetching username:", err)
			ErrorHandler(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	// Fetch Categories
	catRows, err := db.Query("SELECT id, name FROM categories")
	if err != nil {
		log.Println("Error fetching categories:", err)
		ErrorHandler(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer catRows.Close()

	var categories []models.Category
	for catRows.Next() {
		var c models.Category
		if err := catRows.Scan(&c.ID, &c.Name); err == nil {
			categories = append(categories, c)
		}
	}

	// Fetch Posts
	categoryFilter := r.URL.Query().Get("category")
	var postRows *sql.Rows

	if categoryFilter != "" {
		postRows, err = db.Query(`
			SELECT p.id, p.user_id, u.username, p.title, p.content, p.created_at
			FROM posts p
			JOIN users u ON p.user_id = u.id
			JOIN post_categories pc ON p.id = pc.post_id
			JOIN categories c ON pc.category_id = c.id
			WHERE c.name = ?
			ORDER BY p.created_at DESC
		`, categoryFilter)
	} else {
		postRows, err = db.Query(`
			SELECT p.id, p.user_id, u.username, p.title, p.content, p.created_at
			FROM posts p
			JOIN users u ON p.user_id = u.id
			ORDER BY p.created_at DESC
		`)
	}

	if err != nil {
		log.Println("Error fetching posts:", err)
		ErrorHandler(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer postRows.Close()

	var posts []models.Post
	for postRows.Next() {
		var p models.Post
		if err := postRows.Scan(&p.ID, &p.UserID, &p.Author, &p.Title, &p.Content, &p.CreatedAt); err != nil {
			log.Println("Error scanning post:", err)
			continue
		}

		// Fetch comments count
		db.QueryRow("SELECT COUNT(*) FROM comments WHERE post_id = ?", p.ID).Scan(&p.CommentsCount)
		// Fetch likes/dislikes
		db.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = ? AND value = 1", p.ID).Scan(&p.LikesCount)
		db.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = ? AND value = -1", p.ID).Scan(&p.DislikesCount)
		
		if isLoggedIn {
			db.QueryRow("SELECT COALESCE(value, 0) FROM likes WHERE post_id = ? AND user_id = ?", p.ID, userID).Scan(&p.UserLike)
		}

		posts = append(posts, p)
	}
	// Fetch Members (registered users)
	memberRows, err := db.Query("SELECT id, username, created_at FROM users ORDER BY created_at DESC LIMIT 10")
	if err != nil {
		log.Println("Error fetching members:", err)
	}
	type Member struct {
		ID        int
		Username  string
		CreatedAt string
		Initial   string
	}
	var members []Member
	if memberRows != nil {
		defer memberRows.Close()
		for memberRows.Next() {
			var m Member
			var createdAt interface{}
			if err := memberRows.Scan(&m.ID, &m.Username, &createdAt); err == nil {
				if m.Username != "" {
					m.Initial = string([]rune(m.Username)[0])
				}
				members = append(members, m)
			}
		}
	}

	// 3. Prepare the data for HTML
	data := map[string]interface{}{
		"IsLoggedIn": isLoggedIn,
		"Username":   username,
		"Categories": categories,
		"Posts":      posts,
		"Members":   members,
	}

	// 4. Prevent browser caching (So Logout works perfectly)
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// 5. Render the page
	err = tmpl.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		log.Println("Error rendering home page:", err)
		ErrorHandler(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
