package main

import (
	"log"
	"os"

	"github.com/dinesh-88/elevare-auction/internal/bootstrap"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if err := bootstrap.RunMigrations(databaseURL); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("migrations applied")
}
