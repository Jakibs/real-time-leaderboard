package handlers

import (
	"Leaderboard/models"
	"Leaderboard/storage"
	"encoding/json"
	"net/http"
	"time"
)

type TopPlayersReport struct {
	Period     string                    `json:"period"`
	StartDate  string                    `json:"start_date"`
	EndDate    string                    `json:"end_date"`
	TopPlayers []models.LeaderboardEntry `json:"top_players"`
}

func GetTopPlayersReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "day"
	}

	var startDate time.Time
	endDate := time.Now()

	switch period {
	case "day":
		startDate = endDate.AddDate(0, 0, -1)
	case "week":
		startDate = endDate.AddDate(0, 0, -7)
	case "month":
		startDate = endDate.AddDate(0, -1, 0)
	case "year":
		startDate = endDate.AddDate(-1, 0, 0)
	default:
		http.Error(w, "Invalid period. Use: day, week, month, year", http.StatusBadRequest)
		return
	}

	query := `
        SELECT username, MAX(score) as max_score
        FROM leaderboard
        WHERE submitted_at BETWEEN $1 AND $2
        GROUP BY username
        ORDER BY max_score DESC
        LIMIT 10
    `

	rows, err := storage.DB.Query(query, startDate, endDate)
	if err != nil {
		http.Error(w, "Failed to get report", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var topPlayers []models.LeaderboardEntry
	rank := int64(1)
	for rows.Next() {
		var entry models.LeaderboardEntry
		var score float64
		if err := rows.Scan(&entry.Username, &score); err != nil {
			continue
		}
		entry.Score = score
		entry.Rank = rank
		topPlayers = append(topPlayers, entry)
		rank++
	}

	report := TopPlayersReport{
		Period:     period,
		StartDate:  startDate.Format("2006-01-02 15:04:05"),
		EndDate:    endDate.Format("2006-01-02 15:04:05"),
		TopPlayers: topPlayers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func GetUserStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}

	query := `
        SELECT 
            COUNT(*) as total_games,
            MAX(score) as best_score,
            AVG(score) as avg_score,
            MIN(submitted_at) as first_game,
            MAX(submitted_at) as last_game
        FROM leaderboard
        WHERE username = $1
    `

	var stats struct {
		TotalGames int       `json:"total_games"`
		BestScore  float64   `json:"best_score"`
		AvgScore   float64   `json:"avg_score"`
		FirstGame  time.Time `json:"first_game"`
		LastGame   time.Time `json:"last_game"`
	}

	err := storage.DB.QueryRow(query, username).Scan(
		&stats.TotalGames,
		&stats.BestScore,
		&stats.AvgScore,
		&stats.FirstGame,
		&stats.LastGame,
	)

	if err != nil {
		http.Error(w, "Failed to get user stats", http.StatusInternalServerError)
		return
	}

	recentQuery := `
        SELECT score, submitted_at
        FROM leaderboard
        WHERE username = $1
        ORDER BY submitted_at DESC
        LIMIT 10
    `

	rows, err := storage.DB.Query(recentQuery, username)
	if err != nil {
		http.Error(w, "Failed to get recent games", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var recentGames []map[string]interface{}
	for rows.Next() {
		var score float64
		var submittedAt time.Time
		if err := rows.Scan(&score, &submittedAt); err != nil {
			continue
		}
		recentGames = append(recentGames, map[string]interface{}{
			"score":        score,
			"submitted_at": submittedAt.Format("2006-01-02 15:04:05"),
		})
	}

	response := map[string]interface{}{
		"username":     username,
		"stats":        stats,
		"recent_games": recentGames,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
