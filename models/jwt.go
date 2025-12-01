package models

import (
	"github.com/golang-jwt/jwt/v5"
)

var JwtKey = []byte("mykey")

type Claims struct {
	UserID   int    `json:"user"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}
