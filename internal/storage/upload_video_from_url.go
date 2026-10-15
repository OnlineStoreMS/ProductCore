package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// UploadVideoFromURL 下载远程视频并写入存储（素材中心选取等）
func UploadVideoFromURL(store Storage, rawURL string, opts UploadOptions) (string, error) {
	if err := opts.Validate(); err != nil {
		return "", err
	}
	tmpPath, filename, err := DownloadRemote(rawURL, maxRemoteVideoSize)
	if err != nil {
		return "", err
	}
	defer func() { _ = os.Remove(tmpPath) }()

	if !isVideoExt(filepath.Ext(filename)) {
		return "", fmt.Errorf("unsupported video type (mp4/webm/mov/avi)")
	}
	// 素材相对路径可能经网关代理；文件名已规范化
	filename = safeFilename(strings.TrimSpace(filename))
	return store.UploadPath(tmpPath, filename, opts)
}
