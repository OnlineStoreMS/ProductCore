package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"productcore/internal/config"
)

type LocalStorage struct {
	baseDir    string
	baseURL    string
	rootPrefix string
	resolver   *PublicURLResolver
}

func NewLocal(cfg *config.StorageConfig) (*LocalStorage, error) {
	contentDir := cfg.LocalContentDir()
	if err := os.MkdirAll(contentDir, 0o755); err != nil {
		return nil, err
	}
	return &LocalStorage{
		baseDir:    contentDir,
		baseURL:    strings.TrimRight(cfg.PublicBaseURL, "/"),
		rootPrefix: cfg.ObjectKeyPrefix("local"),
		resolver:   NewPublicURLResolver(cfg),
	}, nil
}

func (s *LocalStorage) Upload(file *multipart.FileHeader, opts UploadOptions) (string, error) {
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
	return s.writeFromReader(src, objectKey, contentType)
}

func (s *LocalStorage) UploadPath(srcPath, originalName string, opts UploadOptions) (string, error) {
	if err := opts.Validate(); err != nil {
		return "", err
	}
	src, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer src.Close()

	contentType := detectContentType(originalName, "")
	objectKey := BuildObjectKey(s.rootPrefix, opts, originalName, contentType)
	return s.writeFromReader(src, objectKey, contentType)
}

func (s *LocalStorage) writeFromReader(src io.Reader, objectKey, contentType string) (string, error) {
	destPath := filepath.Join(s.baseDir, filepath.FromSlash(objectKey))
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return "", err
	}
	dst, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}
	_ = contentType
	return s.publicURL(objectKey), nil
}

func (s *LocalStorage) publicURL(objectKey string) string {
	return fmt.Sprintf("%s/%s", s.baseURL, filepath.ToSlash(objectKey))
}

func (s *LocalStorage) ResolvePublicURL(stored string) string {
	return s.resolver.Resolve(stored)
}

func (s *LocalStorage) CopyStoredToPath(stored, destPath string) error {
	stored = strings.TrimSpace(stored)
	if stored == "" {
		return fmt.Errorf("empty stored url")
	}
	key := s.resolver.ObjectKey(stored)
	if key == "" && strings.HasPrefix(stored, s.baseURL+"/") {
		key = strings.TrimPrefix(stored, s.baseURL+"/")
	}
	if key == "" {
		return fmt.Errorf("cannot resolve object key from %q", stored)
	}
	srcPath := filepath.Join(s.baseDir, filepath.FromSlash(key))
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", srcPath, err)
	}
	defer src.Close()
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	dst, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}
