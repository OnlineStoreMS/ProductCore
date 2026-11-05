package storage

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

const maxUploadBytesSize = 10 << 20 // 10MB

// UploadBytes 把内存中的图片写入存储。
func UploadBytes(store Storage, data []byte, filename string, opts UploadOptions) (string, error) {
	if err := opts.Validate(); err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", fmt.Errorf("empty image")
	}
	if len(data) > maxUploadBytesSize {
		return "", fmt.Errorf("file too large (max 10MB)")
	}
	filename = strings.TrimSpace(filename)
	if filename == "" {
		filename = "image.jpg"
	}
	filename = safeFilename(filename)
	if !isAllowedImageName(filename, "") {
		return "", fmt.Errorf("unsupported image type (jpg/jpeg/jfif/png/webp/svg)")
	}
	tmp, err := os.CreateTemp("", "upload-bytes-*")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()
	if _, err := tmp.Write(data); err != nil {
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	return store.UploadPath(tmpPath, filename, opts)
}

// UploadBase64 解码 base64（可带 data:image/...;base64, 前缀）后写入存储。
func UploadBase64(store Storage, raw string, opts UploadOptions) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("empty image")
	}
	filename := "image.jpg"
	payload := raw
	if i := strings.Index(raw, ","); strings.HasPrefix(raw, "data:") && i > 0 {
		meta := raw[:i]
		payload = raw[i+1:]
		switch {
		case strings.Contains(meta, "image/png"):
			filename = "image.png"
		case strings.Contains(meta, "image/webp"):
			filename = "image.webp"
		case strings.Contains(meta, "image/jpeg"), strings.Contains(meta, "image/jpg"), strings.Contains(meta, "image/jfif"):
			filename = "image.jpg"
		case strings.Contains(meta, "image/svg"):
			filename = "image.svg"
		}
	}
	data, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		data, err = base64.RawStdEncoding.DecodeString(payload)
		if err != nil {
			return "", fmt.Errorf("invalid base64 image")
		}
	}
	if len(data) >= 8 && data[0] == 0x89 && data[1] == 0x50 {
		filename = "image.png"
	} else if len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		filename = "image.webp"
	}
	return UploadBytes(store, data, filename, opts)
}
