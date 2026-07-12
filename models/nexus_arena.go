package models

import "time"

// ArenaMatch represents a sports match (ONE per day limit)
type ArenaMatch struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	SportType     string    `json:"sport_type"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	MatchDate     time.Time `json:"match_date"`
	MatchTime     string    `json:"match_time"`
	Location      string    `json:"location"`
	MaxPlayers    int       `json:"max_players"`
	CurrentPlayers int      `json:"current_players"`
	SkillLevel    string    `json:"skill_level"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

// ArenaParticipant represents a player in a match
type ArenaParticipant struct {
	ID        string    `json:"id"`
	MatchID   string    `json:"match_id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
