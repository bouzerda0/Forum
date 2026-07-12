package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-z]+( [a-z0-9]+)?$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

var tmpl *template.Template

func init() {
	tmpl = template.Must(template.New("").Funcs(template.FuncMap{
		"mkSlice": func(args ...interface{}) []interface{} {
			return args
		},
		"mod": func(i, j int) int {
			return i % j
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"len": func(slice []interface{}) int {
			return len(slice)
		},
		"le": func(a, b int) bool {
			return a <= b
		},
		"lt": func(a, b int) bool {
			return a < b
		},
		"gt": func(a, b int) bool {
			return a > b
		},
	}).ParseGlob("templates/*.html"))
}
func RegisterHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// 1. If GET request: Show the register page
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		tmpl.ExecuteTemplate(w, "register.html", nil)
		return
	}
	

	// 2. If POST request: Save the new user
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			tmpl.ExecuteTemplate(w, "register.html", map[string]interface{}{
				"Error": "Error reading form",
			})
			return
		}
		///////Email & Username check\\\\\\
		username := r.FormValue("username")
		email := r.FormValue("email")

		if len(username) < 4 || len(username) > 20 {
			w.WriteHeader(http.StatusBadRequest)
			tmpl.ExecuteTemplate(w, "register.html", map[string]interface{}{
				"Error": "Username length must be between 4 and 20 characters.",
			})
			return
		}

		if !usernameRegex.MatchString(username) {
			w.WriteHeader(http.StatusBadRequest)
			tmpl.ExecuteTemplate(w, "register.html", map[string]interface{}{
				"Error": "Username must contain only letters and numbers.",
			})
			return
		}

		if !emailRegex.MatchString(email) {
			w.WriteHeader(http.StatusBadRequest)
			tmpl.ExecuteTemplate(w, "register.html", map[string]interface{}{
				"Error": "Please enter a valid email address.",
			})
			return
		}

		var existingID int
		err = db.QueryRow("SELECT id FROM users WHERE email = ? OR username = ?", email, username).Scan(&existingID)

		if err != sql.ErrNoRows {
			w.WriteHeader(http.StatusConflict)
			tmpl.ExecuteTemplate(w, "register.html", map[string]interface{}{
				"Error": "Email or Username already exists",
			})
			return
		}
		///////Password check\\\\
		password := r.FormValue("password")
		password = strings.TrimSpace(password)
		// 3. Hash the password before saving it
		if len(password) > 72 {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request

			tmpl.ExecuteTemplate(w, "register.html", map[string]interface{}{
				"Error": "Password is too long! (72 characters max)",
			})

			return
		}
		if len(password) < 8 {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request

			tmpl.ExecuteTemplate(w, "register.html", map[string]interface{}{
				"Error": "Password is too short! (8 characters minimum)",
			})

			return
		}

		if !strings.ContainsAny(password, "0123456789") {
			w.WriteHeader(http.StatusBadRequest)
			tmpl.ExecuteTemplate(w, "register.html", map[string]interface{}{
				"Error": "Password must contain at least one number",
			})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("Bcrypt hash error:", err)
			tmpl.ExecuteTemplate(w, "register.html", map[string]interface{}{
				"Error": "Error creating account, try again later",
			})
			return
		}

		_, err = db.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)", username, email, string(hashedPassword))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("Database insert error:", err)
			tmpl.ExecuteTemplate(w, "register.html", map[string]interface{}{
				"Error": "Could not create account",
			})
			return
		}

		// Success! Redirect to login page
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// 3. Reject anything else
	ErrorHandler(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}
