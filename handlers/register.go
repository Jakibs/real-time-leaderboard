package handlers

import (
	"Leaderboard/models"
	"Leaderboard/storage"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func Registration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if user.Username == "" || user.PasswordHash == "" {
		http.Error(w, "Username or password is empty", http.StatusBadRequest)
		return
	}
	storage.Mu.Lock()
	defer storage.Mu.Unlock()

	for _, u := range storage.Users {
		if u.Username == user.Username {
			http.Error(w, "Username already exists", http.StatusBadRequest)
			return
		}
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	newId := len(storage.Users) + 1

	storage.Users = append(storage.Users, models.User{
		UserId:       newId,
		Username:     user.Username,
		PasswordHash: string(hashed),
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
