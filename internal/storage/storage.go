package storage

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"productcore/internal/config"
)

type Storage interface {
	Upload(file *multipart.FileHeader, opts UploadOptions) (url string, err error)
	UploadPath(srcPath, originalName string, opts UploadOptions) (url string, err error)
	ResolvePublicURL(stored string) string
	CopyStoredToPath(stored, destPath string) error
}

func New(cfg *config.StorageConfig) (Storage, error) {
	switch cfg.Driver {
	case "minio":
		return NewMinIO(cfg)
	case "local", "":
		return NewLocal(cfg)
	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", cfg.Driver)
	}
}

func safeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, name)
	if name == "" || name == "." {
		return "file.bin"
	}
	return name
}
