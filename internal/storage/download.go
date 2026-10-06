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

const maxRemoteVideoSize = 200 << 20 // 200MB，与后台上传视频上限一致

var remoteDownloadClient = &http.Client{Timeout: 3 * time.Minute}

// DownloadRemote 下载远程文件到临时路径，调用方负责删除 tmpPath。
func DownloadRemote(rawURL string, maxBytes int64) (tmpPath, filename string, err error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", "", fmt.Errorf("empty url")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid url")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", "", fmt.Errorf("unsupported url scheme")
	}
	if maxBytes <= 0 {
		maxBytes = maxRemoteVideoSize
	}

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) ProductCore/1.0")
	req.Header.Set("Referer", "https://item.taobao.com/")

	resp, err := remoteDownloadClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	tmp, err := os.CreateTemp("", "collect-video-*")
	if err != nil {
		return "", "", err
	}
	tmpPath = tmp.Name()
	ok := false
	defer func() {
		_ = tmp.Close()
		if !ok {
			_ = os.Remove(tmpPath)
		}
	}()

	n, err := io.Copy(tmp, io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return "", "", fmt.Errorf("read response: %w", err)
	}
	if n > maxBytes {
		return "", "", fmt.Errorf("file too large (max %dMB)", maxBytes>>20)
	}
	if err := tmp.Close(); err != nil {
		return "", "", err
	}

	filename = videoFilenameFromURL(parsed, resp.Header.Get("Content-Type"))
	ok = true
	return tmpPath, filename, nil
}

func videoFilenameFromURL(parsed *url.URL, contentType string) string {
	name := filepath.Base(parsed.Path)
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == "/" {
		name = "video"
	}
	ext := strings.ToLower(filepath.Ext(name))
	if !isVideoExt(ext) {
		switch strings.ToLower(strings.Split(contentType, ";")[0]) {
		case "video/mp4":
			ext = ".mp4"
		case "video/webm":
			ext = ".webm"
		case "video/quicktime":
			ext = ".mov"
		case "video/x-msvideo":
			ext = ".avi"
		default:
			ext = ".mp4"
		}
		name = strings.TrimSuffix(name, filepath.Ext(name)) + ext
	}
	return safeFilename(name)
}

func isVideoExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".mp4", ".mov", ".webm", ".avi":
		return true
	default:
		return false
	}
}
