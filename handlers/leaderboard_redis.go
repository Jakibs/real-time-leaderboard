package handlers

import (
	"Leaderboard/models"
	"Leaderboard/storage"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strings"
)

func SubmitScoreRedis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization required", http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &models.Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return models.JwtKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var req models.ScoreSubmission
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Score < 0 {
		http.Error(w, "Score cannot be negative", http.StatusBadRequest)
		return
	}

	if req.GameID == "" {
		req.GameID = "global"
	}

	leaderboardKey := fmt.Sprintf("leaderboard:%s", req.GameID)

	err = storage.RedisClient.ZAdd(storage.RedisCtx, leaderboardKey, redis.Z{
		Score:  float64(req.Score),
		Member: claims.Username,
	}).Err()

	if err != nil {
		http.Error(w, "Failed to submit score", http.StatusInternalServerError)
		return
	}

	userDbKey := fmt.Sprintf("user:%s:games", claims.Username)
	storage.RedisClient.SAdd(storage.RedisCtx, userDbKey, req.GameID)

	pgQuery := "INSERT INTO leaderboard (username, score) VALUES ($1, $2)"
	storage.DB.Exec(pgQuery, claims.Username, req.Score)

	rank, _ := storage.RedisClient.ZRevRank(storage.RedisCtx, leaderboardKey, claims.Username).Result()

	results, _ := storage.RedisClient.ZRevRangeWithScores(storage.RedisCtx, leaderboardKey, 0, 9).Result()
	var leaderboard []models.LeaderboardEntry
	for i, result := range results {
		entry := models.LeaderboardEntry{
			Username: result.Member.(string),
			Score:    result.Score,
			Rank:     int64(i + 1),
		}
		leaderboard = append(leaderboard, entry)
	}

	BroadcastLeaderboardUpdate(req.GameID, map[string]interface{}{
		"type":        "leaderboard_update",
		"game_id":     req.GameID,
		"leaderboard": leaderboard,
	})
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Score submitted successfully",
		"rank":    rank + 1,
		"score":   req.Score,
	})
}

func GetLeaderboardRedis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	gameID := r.URL.Query().Get("game_id")
	if gameID == "" {
		gameID = "global"
	}

	leaderboardKey := fmt.Sprintf("leaderboard:%s", gameID)

	results, err := storage.RedisClient.ZRevRangeWithScores(storage.RedisCtx, leaderboardKey, 0, 9).Result()
	if err != nil {
		http.Error(w, "Failed to get leaderboard", http.StatusInternalServerError)
		return
	}

	var leaderboard []models.LeaderboardEntry
	for i, result := range results {
		entry := models.LeaderboardEntry{
			Username: result.Member.(string),
			Score:    result.Score,
			Rank:     int64(i + 1),
		}
		leaderboard = append(leaderboard, entry)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leaderboard)
}

func GetUserRank(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization required", http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &models.Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return models.JwtKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	gameID := r.URL.Query().Get("game_id")
	if gameID == "" {
		gameID = "global"
	}

	leaderboardKey := fmt.Sprintf("leaderboard:%s", gameID)

	rank, err := storage.RedisClient.ZRevRank(storage.RedisCtx, leaderboardKey, claims.Username).Result()
	if err != nil {
		http.Error(w, "User not found in leaderboard", http.StatusNotFound)
		return
	}

	score, err := storage.RedisClient.ZScore(storage.RedisCtx, leaderboardKey, claims.Username).Result()
	if err != nil {
		http.Error(w, "Failed to get score", http.StatusInternalServerError)
		return
	}

	total, _ := storage.RedisClient.ZCard(storage.RedisCtx, leaderboardKey).Result()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"username":      claims.Username,
		"rank":          rank + 1,
		"score":         score,
		"total_players": total,
	})
}
