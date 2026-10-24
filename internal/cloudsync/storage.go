package cloudsync

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/oche/diagram/internal/config"
)

// Storage is the abstract blob store for diagram JSON and preview images.
type Storage interface {
	// Kind returns "s3" or "db".
	Kind() string
	// PutJSON writes a JSON blob. Returns storage key.
	PutJSON(ctx context.Context, userID, projectID string, body []byte) (string, error)
	// PutPreview writes a PNG preview. Returns storage key.
	PutPreview(ctx context.Context, userID, projectID string, body []byte) (string, error)
	// GetJSON reads a stored JSON blob by key.
	GetJSON(ctx context.Context, key string) ([]byte, error)
	// GetPreview reads a stored PNG preview by key.
	GetPreview(ctx context.Context, key string) ([]byte, error)
	// PublicPreviewURL returns a directly-reachable URL when (and only when) a
	// CDN-facing base URL is configured. Empty string means the frontend should
	// fetch via the backend proxy — use this for private / Tailscale-scoped S3.
	PublicPreviewURL(key string) string
	// Delete removes both JSON + preview for a project.
	Delete(ctx context.Context, userID, projectID string, jsonKey, previewKey string) error
}

// NewStorage picks the right backend based on config. Returns the DB-backed
// storage when S3 isn't configured (dev / beta without infra).
func NewStorage(cfg config.StorageConfig) (Storage, error) {
	if cfg.Bucket == "" {
		return NewDBStorage(), nil
	}
	return NewS3Storage(cfg)
}

// --- S3 storage (works with AWS S3 and any S3-compatible server like Garage/MinIO) ---

type s3Storage struct {
	client        *s3.Client
	presign       *s3.PresignClient
	bucket        string
	publicBaseURL string
}

func NewS3Storage(cfg config.StorageConfig) (Storage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("loading aws config for s3: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		if cfg.ForcePathStyle {
			o.UsePathStyle = true
		}
	})
	return &s3Storage{
		client:        client,
		presign:       s3.NewPresignClient(client),
		bucket:        cfg.Bucket,
		publicBaseURL: cfg.PublicBaseURL,
	}, nil
}

func (s *s3Storage) Kind() string { return "s3" }

func jsonKey(userID, projectID string) string {
	return "users/" + userID + "/" + projectID + ".json"
}
func previewKey(userID, projectID string) string {
	return "users/" + userID + "/" + projectID + ".png"
}

func (s *s3Storage) PutJSON(ctx context.Context, userID, projectID string, body []byte) (string, error) {
	key := jsonKey(userID, projectID)
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return "", fmt.Errorf("s3 put json: %w", err)
	}
	return key, nil
}

func (s *s3Storage) PutPreview(ctx context.Context, userID, projectID string, body []byte) (string, error) {
	key := previewKey(userID, projectID)
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String("image/png"),
	})
	if err != nil {
		return "", fmt.Errorf("s3 put preview: %w", err)
	}
	return key, nil
}

func (s *s3Storage) GetJSON(ctx context.Context, key string) ([]byte, error) {
	return s.getObject(ctx, key, 8<<20)
}

// GetPreview reads a PNG preview. Capped at 2MB to mirror the upload quota.
func (s *s3Storage) GetPreview(ctx context.Context, key string) ([]byte, error) {
	return s.getObject(ctx, key, 2<<20)
}

func (s *s3Storage) getObject(ctx context.Context, key string, cap int64) ([]byte, error) {
	if key == "" {
		return nil, errors.New("empty key")
	}
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("s3 get: %w", err)
	}
	defer out.Body.Close()
	return io.ReadAll(io.LimitReader(out.Body, cap))
}

// PublicPreviewURL returns a direct URL only when a public CDN prefix is
// configured. Empty string tells the caller to serve the preview via the
// backend proxy — necessary when S3 is on a private network (Tailscale, VPC).
func (s *s3Storage) PublicPreviewURL(key string) string {
	if key == "" || s.publicBaseURL == "" {
		return ""
	}
	return trimRightSlash(s.publicBaseURL) + "/" + key
}

func (s *s3Storage) Delete(ctx context.Context, userID, projectID string, jKey, pKey string) error {
	for _, k := range []string{jKey, pKey} {
		if k == "" {
			continue
		}
		_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(k),
		})
		if err != nil {
			return fmt.Errorf("s3 delete %s: %w", k, err)
		}
	}
	return nil
}

func trimRightSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

// --- DB storage (fallback / dev mode) ---

type dbStorage struct{}

// NewDBStorage returns a no-op blob store: Put/Get are unused because the
// service writes the JSON directly into the user_projects.data column.
func NewDBStorage() Storage { return &dbStorage{} }

func (d *dbStorage) Kind() string                                                    { return "db" }
func (d *dbStorage) PutJSON(context.Context, string, string, []byte) (string, error) { return "", nil }
func (d *dbStorage) PutPreview(context.Context, string, string, []byte) (string, error) {
	return "", nil
}
func (d *dbStorage) GetJSON(context.Context, string) ([]byte, error) {
	return nil, errors.New("db storage does not use keys")
}
func (d *dbStorage) GetPreview(context.Context, string) ([]byte, error) {
	return nil, errors.New("db storage does not use keys")
}
func (d *dbStorage) PublicPreviewURL(string) string                               { return "" }
func (d *dbStorage) Delete(context.Context, string, string, string, string) error { return nil }
