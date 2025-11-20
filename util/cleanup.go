package util

import (
	"context"
	"log"
	"time"
	"backgammon/repository"
)

func CleanupStaleLobbyPresence() {
	db := repository.GetDB()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	log.Println("Started stale lobby presence cleanup job (runs every 30s)")
	for range ticker.C {
		count, err := db.CleanupStaleLobbyPresence(context.Background(), 60*time.Second)
		if err != nil {
			log.Printf("Failed to cleanup stale lobby presence: %v", err)
		} else if count > 0 {
			log.Printf("Removed %d stale lobby presence records", count)
		}
	}
}

func CleanupExpiredInvitations() {
	db := repository.GetDB()
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	log.Println("Started expired invitation cleanup job (runs every 60s)")
	for range ticker.C {
		count, err := db.CleanupExpiredInvitations(context.Background(), 5*time.Minute)
		if err != nil {
			log.Printf("Failed to cleanup expired invitations: %v", err)
		} else if count > 0 {
			log.Printf("Marked %d invitations as expired", count)
		}
	}
}