<p align="center">
  <h1 align="center">🗨️ Forum</h1>
  <p align="center">
    A full-featured web forum built from scratch in Go — no frameworks, just the standard library.
    <br/>
    <strong>Authentication · Posts · Comments · Likes · Categories · Filtering</strong>
  </p>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go 1.25"/>
  <img src="https://img.shields.io/badge/SQLite-3-003B57?style=for-the-badge&logo=sqlite&logoColor=white" alt="SQLite"/>
  <img src="https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker"/>
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License"/>
</p>

---

## 📖 Table of Contents

- [About](#-about)
- [Features](#-features)
- [Architecture](#-architecture)
- [Getting Started](#-getting-started)
  - [Prerequisites](#prerequisites)
  - [Run with Docker](#-run-with-docker)
  - [Run Locally (without Docker)](#-run-locally-without-docker)
- [Docker Commands Reference](#-docker-commands-reference)
- [Project Structure](#-project-structure)
- [Database Schema](#-database-schema)
- [API Routes](#-api-routes)
- [Team](#-team)

---

## 📝 About

**Forum** is a web-based discussion platform where users can register, log in, create posts, comment, and interact through likes and dislikes. The entire backend is written in **pure Go** using only the standard library (`net/http`) — no external web frameworks. Data is persisted in an embedded **SQLite** database, and templates are rendered server-side using Go's `html/template` package.

The project comes **Docker-ready** with a production-optimized Dockerfile for easy deployment.

---

## ✨ Features

| Feature | Description |
|---|---|
| 🔐 **User Authentication** | Secure registration & login with **bcrypt** password hashing and cookie-based sessions |
| 📝 **Create Posts** | Authenticated users can publish posts with a title, content, and one or more categories |
| 💬 **Comments** | Add comments to any post — drives the conversation forward |
| 👍👎 **Likes & Dislikes** | React to posts *and* individual comments with like/dislike votes |
| 📂 **Categories** | Organize posts under categories *(General, Tech, Gaming, Movies, Science)* |
| 🔍 **Filtering** | Filter the feed by: **All Posts**, **My Posts**, **Liked Posts**, or by **Category** |
| 🛡️ **Session Management** | Server-side sessions with secure token generation and automatic expiration (24h) |
| 🎨 **Responsive UI** | Clean, modern interface with custom CSS — no CSS frameworks |
| ⚠️ **Custom Error Pages** | Styled error pages for 400, 404, 405, 500, etc. |
| 🌱 **Seed Data** | Pre-populated database with sample users, posts, comments, and votes for instant demo |

---

## 🚀 Getting Started

### Prerequisites

- **Docker** (recommended) — [Install Docker](https://docs.docker.com/get-docker/)
- *or* **Go 1.25+** — [Install Go](https://go.dev/dl/)
- *and* **GCC** (required by the SQLite driver `go-sqlite3`)

---

### 🐳 Run with Docker

**1. Clone the repository**

```bash
git clone https://github.com/meedsr07/forum.git
cd forum
```

**2. Build the Docker image**

```bash
docker image build -f dockerfile -t forum .
```

**3. Run the container**

```bash
docker container run -p 8088:8088 --detach --name forum forum
```

**4. Open in your browser**

```
http://localhost:8088
```

---

### 🖥 Run Locally (without Docker)

```bash
# Clone the repository
git clone https://github.com/meedsr07/forum.git
cd forum

# Download dependencies
go mod download

# Build and run
go build -o forum && ./forum
```

The server will start at **http://localhost:8088**

---

## 📦 Docker Commands Reference

| Action | Command |
|---|---|
| **Build image** | `docker image build -f dockerfile -t forum .` |
| **Run container** | `docker container run -p 8088:8088 --detach --name forum forum` |
| **Stop container** | `docker container stop forum` |
| **Start stopped container** | `docker container start forum` |
| **Remove container** | `docker container rm forum` |
| **Remove image** | `docker image rm forum` |
| **View running containers** | `docker container ls` |
| **View all containers** | `docker container ls -a` |
| **View container logs** | `docker container logs forum` |
| **Follow logs in real-time** | `docker container logs -f forum` |
| **Enter container shell** | `docker container exec -it forum /bin/bash` |
| **List images** | `docker image ls` |
| **Rebuild (no cache)** | `docker image build --no-cache -f dockerfile -t forum .` |
| **Stop & remove all** | `docker container stop forum && docker container rm forum && docker image rm forum` |

---

## 📁 Project Structure

```
forum/
├── main.go                  # Entry point — registers routes & starts the server
├── go.mod                   # Go module definition
├── go.sum                   # Dependency checksums
├── dockerfile               # Docker build instructions
├── forum.db                 # SQLite database (auto-created)
│
├── database/
│   ├── db.go                # Database initialization & connection
│   ├── schema.sql           # Table definitions (DDL)
│   ├── seed.go              # Sample data seeder
│   ├── GetPsostdata.go      # Post query functions
│   ├── comment.go           # Comment query functions
│   └── likes.go             # Like/dislike query functions
│
├── handlers/
│   ├── register.go          # User registration (GET/POST)
│   ├── login.go             # User login & session creation
│   ├── logout.go            # Session destruction
│   ├── filtration.go        # Home page handler & post filtering
│   ├── createPost.go        # New post creation
│   ├── post.go              # Single post detail view
│   ├── comment.go           # Comment creation
│   ├── like.go              # Post & comment like handler
│   ├── dislike.go           # Post & comment dislike handler
│   ├── categories.go        # Category-based post filtering
│   ├── staticHandlers.go    # Static file serving
│   └── Errorhandler.go      # Custom error page renderer
│
├── models/
│   └── models.go            # Data structures (User, Post, Comment, Like, etc.)
│
├── templates/
│   ├── index.html           # Home page (post feed + sidebar)
│   ├── post.html            # Single post detail with comments
│   ├── login.html           # Login form
│   ├── register.html        # Registration form
│   ├── navbar.html          # Navigation bar partial
│   └── error.html           # Error page template
│
└── static/
    ├── Global.css            # Global reset & base styles
    ├── main.css              # Main layout
    ├── navbar.css            # Navigation bar styles
    ├── post.css              # Post card styles
    ├── postdetail.css        # Post detail page styles
    ├── auth.css              # Login & register form styles
    ├── categories.css        # Category sidebar styles
    ├── filter.css            # Filter bar styles
    └── errorpage.css         # Error page styles
```

---

## 🌐 API Routes

| Method | Route | Auth | Description |
|--------|-------|:----:|-------------|
| `GET` | `/` | ❌ | Home page — displays all posts with filters |
| `GET` | `/?filter=myposts` | ✅ | Filter: show only the logged-in user's posts |
| `GET` | `/?filter=liked` | ✅ | Filter: show only posts the user has liked |
| `GET` | `/?category={id}` | ❌ | Filter: show posts by category |
| `GET` | `/post/{id}` | ❌ | View a single post with its comments and votes |
| `POST` | `/Post/CreatePost` | ✅ | Create a new post |
| `POST` | `/comment/create` | ✅ | Add a comment to a post |
| `POST` | `/like/{id}` | ✅ | Like a post |
| `POST` | `/dislike/{id}` | ✅ | Dislike a post |
| `POST` | `/comment/like/{id}` | ✅ | Like a comment |
| `POST` | `/comment/dislike/{id}` | ✅ | Dislike a comment |
| `GET` | `/login` | ❌ | Login page |
| `POST` | `/login` | ❌ | Submit login credentials |
| `GET` | `/register` | ❌ | Registration page |
| `POST` | `/register` | ❌ | Submit registration form |
| `POST` | `/logout` | ✅ | Destroy session & log out |

---

## 👥 Team

This project was proudly built by:

| Member | GITEA |
|--------|--------|
| **msarar** | [@msarar](https://gitea.com/msarar) |
| **abouzerd** | [@abouzerd](https://gitea.com/abouzerd) |
| **melbouzi** | [@melbouzi](https://gitea.com/melbouzi) |
| **ouaaitalla** | [@ouaaitalla](https://gitea.com/ouaaitalla) |
| **mbenboua** | [@mbenboua](https://github.com/mbenboua) |

---

<p align="center">
  Made with ❤️ and Go
</p>
