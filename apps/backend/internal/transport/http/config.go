package httptransport

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr                string
	DatabaseURL         string
	JWTSecret           string
	PlatformAdminToken  string
	SitePreviewToken    string
	StorageProvider     string
	UploadsDir          string
	S3Bucket            string
	S3Region            string
	S3Endpoint          string
	S3AccessKeyID       string
	S3SecretAccessKey   string
	S3PublicBaseURL     string
	S3UsePathStyle      bool
	JWTTTL              time.Duration
	AccessTokenCookie   string
	CSRFCookie          string
	CookieSecure        bool
	TenantAppScheme     string
	TenantAppBaseDomain string
	TenantAppPort       string
}

func LoadConfig() Config {
	ttl := getEnvDuration("JWT_TTL", 30*time.Minute)
	return Config{
		Addr:                resolveListenAddr(),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/elevare-auction?sslmode=disable"),
		JWTSecret:           getEnv("JWT_SECRET", "dev-secret-change-me"),
		PlatformAdminToken:  getEnv("PLATFORM_ADMIN_TOKEN", "dev-platform-admin-token"),
		SitePreviewToken:    getEnv("SITE_PREVIEW_TOKEN", "dev-site-preview-token"),
		StorageProvider:     getEnv("STORAGE_PROVIDER", "local"),
		UploadsDir:          getEnv("UPLOADS_DIR", "/tmp/elevare-uploads"),
		S3Bucket:            getEnv("S3_BUCKET", ""),
		S3Region:            getEnv("S3_REGION", "us-east-1"),
		S3Endpoint:          getEnv("S3_ENDPOINT", ""),
		S3AccessKeyID:       getEnv("S3_ACCESS_KEY_ID", ""),
		S3SecretAccessKey:   getEnv("S3_SECRET_ACCESS_KEY", ""),
		S3PublicBaseURL:     getEnv("S3_PUBLIC_BASE_URL", ""),
		S3UsePathStyle:      getEnvBool("S3_USE_PATH_STYLE", true),
		JWTTTL:              ttl,
		AccessTokenCookie:   getEnv("ACCESS_TOKEN_COOKIE_NAME", "access_token"),
		CSRFCookie:          getEnv("CSRF_COOKIE_NAME", "csrf_token"),
		CookieSecure:        getEnvBool("COOKIE_SECURE", false),
		TenantAppScheme:     getEnv("TENANT_APP_SCHEME", "https"),
		TenantAppBaseDomain: getEnv("TENANT_APP_BASE_DOMAIN", "elevareauctions.com"),
		TenantAppPort:       strings.TrimSpace(os.Getenv("TENANT_APP_PORT")),
	}
}

func resolveListenAddr() string {
	raw := strings.TrimSpace(os.Getenv("APP_ADDR"))
	if raw == "" {
		raw = strings.TrimSpace(os.Getenv("PORT"))
	}
	if raw == "" {
		return ":8080"
	}
	if strings.HasPrefix(raw, ":") {
		return raw
	}
	if _, err := strconv.Atoi(raw); err == nil {
		return ":" + raw
	}
	return raw
}

func getEnv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(k string, fallback bool) bool {
	v := os.Getenv(k)
	if v == "" {
		return fallback
	}
	switch v {
	case "1", "true", "TRUE", "True", "yes", "YES", "on", "ON":
		return true
	case "0", "false", "FALSE", "False", "no", "NO", "off", "OFF":
		return false
	default:
		return fallback
	}
}

func getEnvDuration(k string, fallback time.Duration) time.Duration {
	v := os.Getenv(k)
	if v == "" {
		return fallback
	}
	if parsed, err := time.ParseDuration(v); err == nil && parsed > 0 {
		return parsed
	}
	if seconds, err := strconv.Atoi(v); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	return fallback
}
