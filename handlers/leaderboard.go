package handlers

import (
	"Leaderboard/storage"
	"encoding/json"
	"net/http"
)

func GetLeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	storage.Mu.Lock()
	defer storage.Mu.Unlock()

	topN := 10
	if len(storage.Leaderboard) < topN {
		topN = len(storage.Leaderboard)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(storage.Leaderboard[:topN])
}
