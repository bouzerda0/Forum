package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"forum/models"
)

// AllPostsHandler handles requests to the dedicated all posts page
func AllPostsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// 1. Check if user is logged in
	userID, err := GetUserIDFromCookie(r, db)
	isLoggedIn := (err == nil)
	username := ""

	// 2. If logged in, get the real username
	if isLoggedIn {
		err = db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
		if err != nil {
			log.Println("Error fetching username:", err)
			ErrorHandler(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	// 3. Fetch all posts
	postRows, err := db.Query(`
		SELECT p.id, p.user_id, u.username, p.title, p.content, p.created_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		ORDER BY p.created_at DESC
	`)
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

		// Fetch Categories for this post
		catRowsForPost, _ := db.Query(`
			SELECT c.name FROM categories c
			JOIN post_categories pc ON c.id = pc.category_id
			WHERE pc.post_id = ?
		`, p.ID)
		if catRowsForPost != nil {
			for catRowsForPost.Next() {
				var name string
				catRowsForPost.Scan(&name)
				p.Categories = append(p.Categories, name)
			}
			catRowsForPost.Close()
		}

		posts = append(posts, p)
	}

	// 4. Fetch Categories for styling or tags if needed (optional, doing it just in case)
	catRows, err := db.Query("SELECT id, name FROM categories")
	var categories []models.Category
	if err == nil {
		defer catRows.Close()
		for catRows.Next() {
			var c models.Category
			if err := catRows.Scan(&c.ID, &c.Name); err == nil {
				categories = append(categories, c)
			}
		}
	}

	// 5. Prepare the data for HTML
	data := map[string]interface{}{
		"IsLoggedIn": isLoggedIn,
		"Username":   username,
		"Posts":      posts,
		"Categories": categories,
	}

	// 6. Prevent browser caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// 7. Render the page
	err = tmpl.ExecuteTemplate(w, "all_posts.html", data)
	if err != nil {
		log.Println("Error rendering all posts page:", err)
		ErrorHandler(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
