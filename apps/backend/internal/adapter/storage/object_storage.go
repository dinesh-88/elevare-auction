package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type LocalObjectStorage struct {
	BaseDir       string
	BaseURLPrefix string
}

func (s *LocalObjectStorage) Put(_ context.Context, key string, _ string, body io.Reader, _ int64) (string, error) {
	safeKey := path.Clean("/" + key)
	safeKey = strings.TrimPrefix(safeKey, "/")
	if safeKey == "" {
		return "", fmt.Errorf("invalid key")
	}

	diskPath := filepath.Join(s.BaseDir, filepath.FromSlash(safeKey))
	if err := os.MkdirAll(filepath.Dir(diskPath), 0o755); err != nil {
		return "", err
	}

	dst, err := os.Create(diskPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, body); err != nil {
		return "", err
	}

	basePrefix := strings.TrimRight(s.BaseURLPrefix, "/")
	return basePrefix + "/" + safeKey, nil
}

type S3ObjectStorage struct {
	Client        *s3.Client
	Bucket        string
	PublicBaseURL string
	Region        string
}

type S3Config struct {
	Region          string
	Bucket          string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	PublicBaseURL   string
	UsePathStyle    bool
}

func NewS3ObjectStorage(ctx context.Context, cfg S3Config) (*S3ObjectStorage, error) {
	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint != "" && !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}
	if endpoint != "" {
		if _, err := url.ParseRequestURI(endpoint); err != nil {
			return nil, fmt.Errorf("invalid S3 endpoint %q: %w", cfg.Endpoint, err)
		}
	}

	loadOpts := make([]func(*config.LoadOptions) error, 0, 2)
	loadOpts = append(loadOpts, config.WithRegion(cfg.Region))
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		loadOpts = append(loadOpts, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")))
	}
	awsCfg, err := config.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
		o.UsePathStyle = cfg.UsePathStyle
	})

	return &S3ObjectStorage{
		Client:        client,
		Bucket:        cfg.Bucket,
		PublicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
		Region:        cfg.Region,
	}, nil
}

func (s *S3ObjectStorage) Put(ctx context.Context, key string, contentType string, body io.Reader, contentLength int64) (string, error) {
	safeKey := path.Clean("/" + key)
	safeKey = strings.TrimPrefix(safeKey, "/")
	if safeKey == "" {
		return "", fmt.Errorf("invalid key")
	}

	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.Bucket),
		Key:           aws.String(safeKey),
		Body:          body,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(contentLength),
	})
	if err != nil {
		return "", err
	}

	if s.PublicBaseURL != "" {
		return s.PublicBaseURL + "/" + safeKey, nil
	}

	escapedKey := strings.ReplaceAll(url.PathEscape(safeKey), "%2F", "/")
	if s.Region == "" {
		return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.Bucket, escapedKey), nil
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.Bucket, s.Region, escapedKey), nil
}
