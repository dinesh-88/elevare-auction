package bootstrap

import (
	"errors"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(databaseURL string) error {
	if strings.TrimSpace(databaseURL) == "" {
		return errors.New("DATABASE_URL is required")
	}

	migrationsPath := resolveMigrationsPath()
	dsn := migrateDatabaseURL(databaseURL)

	m, err := migrate.New("file://"+migrationsPath, dsn)
	if err != nil {
		return err
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
		return err
	}
	return nil
}

func resolveMigrationsPath() string {
	if configured := strings.TrimSpace(os.Getenv("MIGRATIONS_PATH")); configured != "" {
		if abs, err := filepath.Abs(configured); err == nil {
			return abs
		}
		return configured
	}

	candidates := []string{
		"/app/migrations",
		"migrations",
		"./migrations",
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			if abs, absErr := filepath.Abs(candidate); absErr == nil {
				return abs
			}
			return candidate
		}
	}
	return "/app/migrations"
}

func migrateDatabaseURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	switch strings.ToLower(parsed.Scheme) {
	case "postgres", "postgresql":
		parsed.Scheme = "pgx5"
	}
	return parsed.String()
}
