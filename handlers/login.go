package handlers

import (
	"Leaderboard/models"
	"Leaderboard/storage"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}
	var req models.User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Login error", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.PasswordHash == "" {
		http.Error(w, "Username or password is empty", http.StatusBadRequest)
		return
	}
	storage.Mu.Lock()
	defer storage.Mu.Unlock()

	var userfound *models.User
	for i, u := range storage.Users {
		if u.Username == req.Username {
			userfound = &storage.Users[i]
			break
		}
	}
	if userfound == nil {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(userfound.PasswordHash), []byte(req.PasswordHash)); err != nil {
		http.Error(w, "Invalid password", http.StatusBadRequest)
		return
	}
	expTime := (time.Now().Add(24 * time.Hour))

	claims := &models.Claims{
		UserID:   userfound.UserId,
		Username: userfound.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(models.JwtKey)
	if err != nil {
		http.Error(w, "Could not create token", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}
