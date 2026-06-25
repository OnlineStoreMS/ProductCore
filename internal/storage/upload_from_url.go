package storage

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxUploadFromURLSize = 10 << 20 // 10MB

var uploadFromURLClient = &http.Client{Timeout: 45 * time.Second}

// UploadFromURL 下载远程图片并写入存储（用于 SKU 导入等）
func UploadFromURL(store Storage, rawURL string, opts UploadOptions) (string, error) {
	if err := opts.Validate(); err != nil {
		return "", err
	}
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", fmt.Errorf("empty url")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid url")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("unsupported url scheme")
	}

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "ProductCore/1.0")

	resp, err := uploadFromURLClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	body := io.LimitReader(resp.Body, maxUploadFromURLSize+1)
	tmp, err := os.CreateTemp("", "upload-url-*")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()

	n, err := io.Copy(tmp, body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if n > maxUploadFromURLSize {
		return "", fmt.Errorf("file too large (max 10MB)")
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}

	filename := filenameFromURL(parsed, resp.Header.Get("Content-Type"))
	if !isAllowedImageName(filename, resp.Header.Get("Content-Type")) {
		return "", fmt.Errorf("unsupported image type (jpg/jpeg/png/webp)")
	}
	return store.UploadPath(tmpPath, filename, opts)
}

func filenameFromURL(parsed *url.URL, contentType string) string {
	name := filepath.Base(parsed.Path)
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == "/" {
		name = "image"
	}
	ext := filepath.Ext(name)
	if ext == "" {
		switch strings.ToLower(strings.Split(contentType, ";")[0]) {
		case "image/jpeg", "image/jpg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		default:
			ext = ".jpg"
		}
		name += ext
	}
	return safeFilename(name)
}

func isAllowedImageName(filename, contentType string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	}
	ct := strings.ToLower(strings.Split(contentType, ";")[0])
	switch ct {
	case "image/jpeg", "image/jpg", "image/png", "image/webp":
		return true
	}
	return false
}
