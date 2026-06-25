package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"productcore/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOStorage struct {
	client     *minio.Client
	bucket     string
	baseURL    string
	rootPrefix string
	resolver   *PublicURLResolver
}

func NewMinIO(cfg *config.StorageConfig) (*MinIOStorage, error) {
	m := cfg.MinIO
	if m.Endpoint == "" {
		return nil, fmt.Errorf("minio endpoint required")
	}
	if m.AccessKey == "" || m.SecretKey == "" {
		return nil, fmt.Errorf("minio access_key and secret_key required")
	}
	if m.Bucket == "" {
		return nil, fmt.Errorf("minio bucket required")
	}

	client, err := minio.New(m.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(m.AccessKey, m.SecretKey, ""),
		Secure: m.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}

	ctx := context.Background()
	if _, err := client.ListBuckets(ctx); err != nil {
		return nil, fmt.Errorf("minio connect %s: %w", m.Endpoint, err)
	}

	exists, err := client.BucketExists(ctx, m.Bucket)
	if err != nil {
		return nil, fmt.Errorf("minio bucket check: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, m.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("minio make bucket: %w", err)
		}
	}

	if m.PublicRead {
		if err := setBucketPublicRead(ctx, client, m.Bucket); err != nil {
			return nil, fmt.Errorf("minio bucket policy: %w", err)
		}
	}

	baseURL := cfg.PublicBaseURL
	if baseURL == "" {
		scheme := "http"
		if m.UseSSL {
			scheme = "https"
		}
		baseURL = fmt.Sprintf("%s://%s/%s", scheme, m.Endpoint, m.Bucket)
	}

	return &MinIOStorage{
		client:     client,
		bucket:     m.Bucket,
		baseURL:    strings.TrimRight(baseURL, "/"),
		rootPrefix: cfg.ObjectKeyPrefix("minio"),
		resolver:   NewPublicURLResolver(cfg),
	}, nil
}

func setBucketPublicRead(ctx context.Context, client *minio.Client, bucket string) error {
	policy := map[string]any{
		"Version": "2012-10-17",
		"Statement": []map[string]any{{
			"Effect":    "Allow",
			"Principal": map[string]any{"AWS": []string{"*"}},
			"Action":    []string{"s3:GetObject"},
			"Resource":  []string{fmt.Sprintf("arn:aws:s3:::%s/*", bucket)},
		}},
	}
	raw, err := json.Marshal(policy)
	if err != nil {
		return err
	}
	return client.SetBucketPolicy(ctx, bucket, string(raw))
}

func (s *MinIOStorage) Upload(file *multipart.FileHeader, opts UploadOptions) (string, error) {
	if err := opts.Validate(); err != nil {
		return "", err
	}
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	contentType := detectContentType(file.Filename, file.Header.Get("Content-Type"))
	objectKey := BuildObjectKey(s.rootPrefix, opts, file.Filename, contentType)

	_, err = s.client.PutObject(context.Background(), s.bucket, objectKey, src, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return s.publicURL(objectKey), nil
}

func (s *MinIOStorage) UploadPath(srcPath, originalName string, opts UploadOptions) (string, error) {
	if err := opts.Validate(); err != nil {
		return "", err
	}
	src, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer src.Close()
	stat, err := src.Stat()
	if err != nil {
		return "", err
	}

	contentType := detectContentType(originalName, "")
	objectKey := BuildObjectKey(s.rootPrefix, opts, originalName, contentType)

	_, err = s.client.PutObject(context.Background(), s.bucket, objectKey, src, stat.Size(), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return s.publicURL(objectKey), nil
}

func (s *MinIOStorage) publicURL(objectKey string) string {
	return s.baseURL + "/" + objectKey
}

func (s *MinIOStorage) ResolvePublicURL(stored string) string {
	return s.resolver.Resolve(stored)
}

func (s *MinIOStorage) CopyStoredToPath(stored, destPath string) error {
	stored = strings.TrimSpace(stored)
	if stored == "" {
		return fmt.Errorf("empty stored url")
	}
	key := s.resolver.ObjectKey(stored)
	if key == "" {
		return fmt.Errorf("cannot resolve object key from %q", stored)
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	if err := s.client.FGetObject(context.Background(), s.bucket, key, destPath, minio.GetObjectOptions{}); err != nil {
		return fmt.Errorf("download %s: %w", key, err)
	}
	return nil
}
