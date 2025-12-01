package storage

import (
	"Leaderboard/models"
	"sync"
)

var Leaderboard []models.LeaderboardEntry
var Users []models.User
var Mu sync.Mutex
