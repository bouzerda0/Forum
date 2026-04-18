package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"forum/models"
)

func PostDetailHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	postID := r.URL.Query().Get("id")
	if postID == "" {
		ErrorHandler(w, "Post Not Found", http.StatusNotFound)
		return
	}

	userID, _ := GetUserIDFromCookie(r, db)

	// Fetch Post
	var p models.Post
	err := db.QueryRow(`
		SELECT p.id, p.user_id, u.username, p.title, p.content, p.created_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = ?
	`, postID).Scan(&p.ID, &p.UserID, &p.Author, &p.Title, &p.Content, &p.CreatedAt)

	if err != nil {
		ErrorHandler(w, "Post Not Found", http.StatusNotFound)
		return
	}

	// Stats
	db.QueryRow("SELECT COUNT(*) FROM comments WHERE post_id = ?", p.ID).Scan(&p.CommentsCount)
	db.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = ? AND value = 1", p.ID).Scan(&p.LikesCount)
	db.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = ? AND value = -1", p.ID).Scan(&p.DislikesCount)
	if userID != 0 {
		db.QueryRow("SELECT COALESCE(value, 0) FROM likes WHERE post_id = ? AND user_id = ?", p.ID, userID).Scan(&p.UserLike)
	}

	// Fetch Categories
	rows, _ := db.Query(`
		SELECT c.name FROM categories c
		JOIN post_categories pc ON c.id = pc.category_id
		WHERE pc.post_id = ?
	`, p.ID)
	for rows.Next() {
		var name string
		rows.Scan(&name)
		p.Categories = append(p.Categories, name)
	}
	rows.Close()

	// Fetch Comments
	commentRows, err := db.Query(`
		SELECT c.id, c.user_id, u.username, c.content, c.created_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ?
		ORDER BY c.created_at ASC
	`, p.ID)

	var comments []models.Comment
	if err == nil {
		for commentRows.Next() {
			var c models.Comment
			commentRows.Scan(&c.ID, &c.UserID, &c.Author, &c.Content, &c.CreatedAt)
			
			// Comment Stats
			db.QueryRow("SELECT COUNT(*) FROM likes WHERE comment_id = ? AND value = 1", c.ID).Scan(&c.LikesCount)
			db.QueryRow("SELECT COUNT(*) FROM likes WHERE comment_id = ? AND value = -1", c.ID).Scan(&c.DislikesCount)
			if userID != 0 {
				db.QueryRow("SELECT COALESCE(value, 0) FROM likes WHERE comment_id = ? AND user_id = ?", c.ID, userID).Scan(&c.UserLike)
			}
			comments = append(comments, c)
		}
		commentRows.Close()
	}

	data := map[string]interface{}{
		"IsLoggedIn": userID != 0,
		"Post":       p,
		"Comments":   comments,
	}
	tmpl.ExecuteTemplate(w, "post.html", data)
}

func CreatePostHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	userID, err := GetUserIDFromCookie(r, db)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	rows, err := db.Query("SELECT id, name FROM categories")
	if err != nil {
		log.Println("Error fetching categories:", err)
		ErrorHandler(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err == nil {
			categories = append(categories, c)
		}
	}

	if r.Method == http.MethodGet {
		data := map[string]interface{}{
			"Categories": categories,
			"Error":      "",
		}
		tmpl.ExecuteTemplate(w, "create_post.html", data)
		return
	}

	if r.Method == http.MethodPost {
		title := strings.TrimSpace(r.FormValue("title"))
		content := strings.TrimSpace(r.FormValue("content"))
		r.ParseForm()
		categoryIDs := r.Form["categories"]

		if title == "" || content == "" || len(categoryIDs) == 0 {
			data := map[string]interface{}{
				"Categories": categories,
				"Error":      "Title, content, and at least one category are required.",
			}
			tmpl.ExecuteTemplate(w, "create_post.html", data)
			return
		}

		res, err := db.Exec("INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?)", userID, title, content)
		if err != nil {
			log.Println("Error inserting post:", err)
			ErrorHandler(w, "Error saving post", http.StatusInternalServerError)
			return
		}

		postID, _ := res.LastInsertId()
		for _, catIDStr := range categoryIDs {
			db.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", postID, catIDStr)
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
}

func LikeHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	userID, err := GetUserIDFromCookie(r, db)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	targetType := r.FormValue("type")
	targetID := r.FormValue("id")
	valueStr := r.FormValue("value")

	var value int
	if valueStr == "1" {
		value = 1
	} else if valueStr == "-1" {
		value = -1
	} else {
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}

	var existingValue int
	var query string
	var args []interface{}

	if targetType == "post" {
		err = db.QueryRow("SELECT value FROM likes WHERE user_id = ? AND post_id = ?", userID, targetID).Scan(&existingValue)
		if err == sql.ErrNoRows {
			query = "INSERT INTO likes (user_id, post_id, value) VALUES (?, ?, ?)"
			args = []interface{}{userID, targetID, value}
		} else {
			if existingValue == value {
				query = "DELETE FROM likes WHERE user_id = ? AND post_id = ?"
				args = []interface{}{userID, targetID}
			} else {
				query = "UPDATE likes SET value = ? WHERE user_id = ? AND post_id = ?"
				args = []interface{}{value, userID, targetID}
			}
		}
	} else {
		err = db.QueryRow("SELECT value FROM likes WHERE user_id = ? AND comment_id = ?", userID, targetID).Scan(&existingValue)
		if err == sql.ErrNoRows {
			query = "INSERT INTO likes (user_id, comment_id, value) VALUES (?, ?, ?)"
			args = []interface{}{userID, targetID, value}
		} else {
			if existingValue == value {
				query = "DELETE FROM likes WHERE user_id = ? AND comment_id = ?"
				args = []interface{}{userID, targetID}
			} else {
				query = "UPDATE likes SET value = ? WHERE user_id = ? AND comment_id = ?"
				args = []interface{}{value, userID, targetID}
			}
		}
	}

	db.Exec(query, args...)
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

func CommentHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	userID, err := GetUserIDFromCookie(r, db)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	postID := r.FormValue("post_id")
	content := strings.TrimSpace(r.FormValue("content"))

	if content != "" {
		db.Exec("INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)", postID, userID, content)
	}

	http.Redirect(w, r, "/post?id="+postID, http.StatusSeeOther)
}