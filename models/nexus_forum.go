package models

import "time"

// ForumMeme represents a meme post (NO ANONYMITY)
type ForumMeme struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	Title         string    `json:"title"`
	ImageURL      string    `json:"image_url"`
	Description   string    `json:"description"`
	LikesCount    int       `json:"likes_count"`
	CommentsCount int       `json:"comments_count"`
	UserLiked     bool      `json:"user_liked"`
	CreatedAt     time.Time `json:"created_at"`
}

// ForumMemeComment represents a comment on a meme
type ForumMemeComment struct {
	ID        string    `json:"id"`
	MemeID    string    `json:"meme_id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ForumPoll represents a community poll
type ForumPoll struct {
	ID                 string     `json:"id"`
	UserID             string     `json:"user_id"`
	Username           string     `json:"username"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	AllowMultiple      bool       `json:"allow_multiple"`
	ClosesAt           *time.Time `json:"closes_at"`
	Status             string     `json:"status"`
	TotalVotes         int        `json:"total_votes"`
	Options            []PollOption `json:"options"`
	UserVoted          bool       `json:"user_voted"`
	CreatedAt          time.Time  `json:"created_at"`
}

// PollOption represents a single poll choice
type PollOption struct {
	ID          string `json:"id"`
	PollID      string `json:"poll_id"`
	OptionText  string `json:"option_text"`
	VoteCount   int    `json:"vote_count"`
	Percentage  float64 `json:"percentage"`
	UserVoted bool   `json:"user_voted"`
}
