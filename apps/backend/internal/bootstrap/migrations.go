package bootstrap

import (
	"errors"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(databaseURL string) error {
	if strings.TrimSpace(databaseURL) == "" {
		return errors.New("DATABASE_URL is required")
	}

	migrationsPath := resolveMigrationsPath()
	dsn := databaseURL
	noLock := strings.TrimSpace(os.Getenv("MIGRATE_NO_LOCK"))
	if noLock == "" || strings.EqualFold(noLock, "true") || noLock == "1" {
		dsn = withQueryParam(databaseURL, "x-no-lock", "true")
	}

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
