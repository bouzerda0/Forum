package models

import "time"

type Post struct {
	ID            int
	UserID        int
	Author        string
	Title         string
	Content       string
	CreatedAt     time.Time
	CommentsCount int
	LikesCount    int
	DislikesCount int
	UserLike      int // 1 for like, -1 for dislike, 0 for none
	Categories    []string
}

type Comment struct {
	ID            int
	PostID        int
	UserID        int
	Author        string
	Content       string
	CreatedAt     time.Time
	LikesCount    int
	DislikesCount int
	UserLike      int
}

type Category struct {
	ID   int
	Name string
}
