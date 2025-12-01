package handlers

import (
	"Leaderboard/models"
	"Leaderboard/storage"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
)

func SubmitScoreHandlerDB(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {

		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	claims := &models.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return models.JwtKey, nil
	})

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req struct {
		Score int `json:"score"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.Score < 0 {
		http.Error(w, "score cannot be negative", http.StatusBadRequest)
		return
	}

	query := "INSERT INTO leaderboard (username, score) VALUES ($1, $2)"
	_, err = storage.DB.Exec(query, claims.Username, req.Score)
	if err != nil {
		http.Error(w, "Failed to submit score", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "score submitted successfully"})
}

func GetLeaderboardHandlerDB(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	query := `
        SELECT username, MAX(score) as max_score 
        FROM leaderboard 
        GROUP BY username 
        ORDER BY max_score DESC 
        LIMIT 10
    `
	rows, err := storage.DB.Query(query)
	if err != nil {
		http.Error(w, "Failed to get leaderboard", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var leaderboard []models.LeaderboardEntry
	for rows.Next() {
		var entry models.LeaderboardEntry
		if err := rows.Scan(&entry.Username, &entry.Score); err != nil {
			continue
		}
		leaderboard = append(leaderboard, entry)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leaderboard)
}
