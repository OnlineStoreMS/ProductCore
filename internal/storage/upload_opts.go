package storage

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrProductIDRequired = errors.New("productId is required for product upload")

// UploadOptions 上传上下文：决定对象存储路径层级
type UploadOptions struct {
	Scope     string // product | common
	ProductID uint64
	SkuID     uint64
	Resource  string // main, album, detail, video_11, sku_pic, spec, platform_logo ...
}

// UploadOptionsFromForm 由 handler 从 gin 表单字段构建
func UploadOptionsFromForm(values map[string]string, isVideo bool) UploadOptions {
	opts := UploadOptions{}

	if scope := strings.TrimSpace(values["scope"]); scope != "" {
		opts.Scope = scope
	}
	if r := strings.TrimSpace(values["resource"]); r != "" {
		opts.Resource = r
	}
	if pid := strings.TrimSpace(values["productId"]); pid != "" {
		if id, err := strconv.ParseUint(pid, 10, 64); err == nil {
			opts.ProductID = id
		}
	}
	if sid := strings.TrimSpace(values["skuId"]); sid != "" {
		if id, err := strconv.ParseUint(sid, 10, 64); err == nil {
			opts.SkuID = id
		}
	}

	if opts.Scope == "" {
		if opts.ProductID > 0 {
			opts.Scope = "product"
		} else {
			opts.Scope = "common"
		}
	}
	if opts.Resource == "" {
		if isVideo {
			opts.Resource = "video"
		} else {
			opts.Resource = "image"
		}
	}
	return opts
}

func (opts UploadOptions) Validate() error {
	if opts.Scope == "product" && opts.ProductID == 0 {
		return ErrProductIDRequired
	}
	return nil
}

func sanitizeResourceHint(resource string) string {
	r := strings.TrimSpace(resource)
	if r == "" {
		return "file"
	}
	return strings.Map(func(c rune) rune {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			return c
		}
		if c == '.' {
			return '_'
		}
		return '_'
	}, r)
}

// BuildStoredFilename 重命名：{resource}_{timestamp}_{random}.{ext}
func BuildStoredFilename(resource, originalName, contentType string) string {
	ext := strings.ToLower(filepath.Ext(originalName))
	if ext == "" {
		ext = extFromContentType(contentType)
	}
	hint := sanitizeResourceHint(resource)
	return fmt.Sprintf("%s_%d_%s%s", hint, time.Now().Unix(), uuid.NewString()[:6], ext)
}

func extFromContentType(ct string) string {
	switch strings.ToLower(ct) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	case "video/mp4":
		return ".mp4"
	case "video/webm":
		return ".webm"
	case "video/quicktime":
		return ".mov"
	default:
		return ".bin"
	}
}

// BuildRelativePath 生成相对存储根的路径（不含 leading slash）
func BuildRelativePath(rootPrefix string, opts UploadOptions, filename string) string {
	segments := make([]string, 0, 8)
	if p := strings.Trim(rootPrefix, "/"); p != "" {
		segments = append(segments, p)
	}

	if opts.Scope != "product" || opts.ProductID == 0 {
		segments = append(segments, "common")
		segments = append(segments, commonResourceDir(opts.Resource)...)
		segments = append(segments, filename)
		return strings.Join(segments, "/")
	}

	segments = append(segments, "products", strconv.FormatUint(opts.ProductID, 10))
	segments = append(segments, productResourceSegments(opts)...)
	segments = append(segments, filename)
	return strings.Join(segments, "/")
}

func commonResourceDir(resource string) []string {
	switch resource {
	case "platform_logo":
		return []string{"platform", "logo"}
	case "video", "video_common":
		return []string{"videos"}
	default:
		return []string{"images"}
	}
}

func productResourceSegments(opts UploadOptions) []string {
	switch opts.Resource {
	case "main", "import_main":
		return []string{"main"}
	case "album":
		return []string{"album"}
	case "pics34":
		return []string{"pics34"}
	case "detail", "import_detail":
		return []string{"detail"}
	case "material_white":
		return []string{"materials", "white"}
	case "material_transparent":
		return []string{"materials", "transparent"}
	case "material_guide34":
		return []string{"materials", "guide34"}
	case "material_long":
		return []string{"materials", "long"}
	case "video", "video_11", "import_video_11":
		return []string{"videos", "ratio11"}
	case "video_34", "import_video_34":
		return []string{"videos", "ratio34"}
	case "video_169":
		return []string{"videos", "ratio169"}
	case "video_916":
		return []string{"videos", "ratio916"}
	case "spec", "sku_spec", "import_sku":
		return []string{"specs"}
	case "sku_pic":
		if opts.SkuID > 0 {
			return []string{"skus", strconv.FormatUint(opts.SkuID, 10), "pic"}
		}
		return []string{"skus", "pic"}
	default:
		return []string{"files"}
	}
}

func detectContentType(filename, headerCT string) string {
	if headerCT != "" && headerCT != "application/octet-stream" {
		return strings.ToLower(headerCT)
	}
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mov":
		return "video/quicktime"
	case ".avi":
		return "video/x-msvideo"
	default:
		return "application/octet-stream"
	}
}
