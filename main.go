package main

import (
	"Leaderboard/handlers"
	"Leaderboard/storage"
	"fmt"
	"log"
	"net/http"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	if err := storage.InitDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer storage.CloseDB()

	if err := storage.InitRedis(); err != nil {
		log.Fatal("Failed to initialize Redis:", err)
	}
	defer storage.CloseRedis()

	go handlers.GlobalHub.Run()

	mux := http.NewServeMux()

	mux.HandleFunc("/score", handlers.SubmitScoreRedis)
	mux.HandleFunc("/leaderboard", handlers.GetLeaderboardRedis)
	mux.HandleFunc("/rank", handlers.GetUserRank)
	mux.HandleFunc("/report", handlers.GetTopPlayersReport)
	mux.HandleFunc("/stats", handlers.GetUserStats)
	mux.HandleFunc("/register", handlers.RegistrationDB)
	mux.HandleFunc("/login", handlers.LoginDB)
	mux.HandleFunc("/ws", handlers.ServeWs)

	handler := enableCORS(mux)

	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		fmt.Println("Server error:", err)
	}
}
