package main

import (
	"errors"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	noLock := strings.TrimSpace(os.Getenv("MIGRATE_NO_LOCK"))
	if noLock == "" || strings.EqualFold(noLock, "true") || noLock == "1" {
		databaseURL = withQueryParam(databaseURL, "x-no-lock", "true")
	}
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "/app/migrations"
	}
	m, err := migrate.New("file://"+migrationsPath, databaseURL)
	if err != nil {
		log.Fatalf("failed to initialize migrator: %v", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Printf("migration source close warning: %v", srcErr)
		}
		if dbErr != nil {
			log.Printf("migration db close warning: %v", dbErr)
		}
	}()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("migrations applied")
}

func withQueryParam(rawURL, key, value string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	query := parsed.Query()
	if query.Get(key) == "" {
		query.Set(key, value)
		parsed.RawQuery = query.Encode()
	}
	return parsed.String()
}
