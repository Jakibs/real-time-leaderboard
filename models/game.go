package models

type Game struct {
	GameID   string `json:"game_id"`
	GameName string `json:"game_name"`
}
type ScoreSubmission struct {
	GameID string `json:"game_id"`
	Score  int    `json:"score"`
}

type LeaderboardEntry struct {
	Username string  `json:"username"`
	Score    float64 `json:"score"`
	Rank     int64   `json:"rank"`
}
