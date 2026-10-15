package storage

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
)

const maxAdaptFromURLSize = 40 << 20 // 素材原图可能很大

// ImageAdaptOptions 电商图适配：可选比例裁剪 + 限边缩放 + JPEG 压缩
type ImageAdaptOptions struct {
	// AspectW/H >0 时按比例裁剪；CropX/Y/W/H 为相对原图 0-1（左上+宽高），缺省居中最大内接矩形
	AspectW int
	AspectH int
	CropX   float64
	CropY   float64
	CropW   float64
	CropH   float64
	// MaxEdge 输出长边上限（1:1 时即边长）；DetailMaxW/H 详情图用
	MaxEdge     int
	DetailMaxW  int
	DetailMaxH  int
	Quality     int // jpeg 1-100，默认 82
	PreferPNG   bool
}

// UploadFromURLAdapt 下载远程图 → 裁剪/缩放/压缩 → 写入商品存储
func UploadFromURLAdapt(store Storage, rawURL string, uploadOpts UploadOptions, adapt ImageAdaptOptions) (string, error) {
	if err := uploadOpts.Validate(); err != nil {
		return "", err
	}
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", fmt.Errorf("empty url")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", fmt.Errorf("invalid url")
	}

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 ProductCore-MaterialAdapt/1.0")

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxAdaptFromURLSize+1))
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if len(data) > maxAdaptFromURLSize {
		return "", fmt.Errorf("file too large (max 40MB)")
	}

	img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		return "", fmt.Errorf("decode image: %w", err)
	}
	img = imaging.Clone(img)
	out := adaptImage(img, adapt)

	quality := adapt.Quality
	if quality <= 0 {
		quality = 82
	}
	if quality > 95 {
		quality = 95
	}

	ext := ".jpg"
	mimeHint := "image/jpeg"
	var buf bytes.Buffer
	if adapt.PreferPNG {
		ext = ".png"
		mimeHint = "image/png"
		if err := png.Encode(&buf, out); err != nil {
			return "", err
		}
	} else {
		if err := jpeg.Encode(&buf, out, &jpeg.Options{Quality: quality}); err != nil {
			return "", err
		}
	}

	tmp, err := os.CreateTemp("", "adapt-*"+ext)
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()
	if _, err := tmp.Write(buf.Bytes()); err != nil {
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}

	base := filenameFromURL(parsed, mimeHint)
	base = strings.TrimSuffix(base, filepath.Ext(base)) + ext
	_ = mimeHint
	return store.UploadPath(tmpPath, base, uploadOpts)
}

func adaptImage(src image.Image, opt ImageAdaptOptions) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 1 || h < 1 {
		return src
	}

	cropped := src
	if opt.AspectW > 0 && opt.AspectH > 0 {
		cw, ch, cx, cy := resolveCropRect(w, h, opt)
		cropped = imaging.Crop(src, image.Rect(cx, cy, cx+cw, cy+ch))
		w, h = cw, ch
	}

	maxEdge := opt.MaxEdge
	if maxEdge <= 0 {
		maxEdge = 1200
	}
	targetW, targetH := w, h

	if opt.DetailMaxW > 0 || opt.DetailMaxH > 0 {
		dw := opt.DetailMaxW
		dh := opt.DetailMaxH
		if dw <= 0 {
			dw = 1290
		}
		if dh <= 0 {
			dh = 2000
		}
		scale := 1.0
		if targetW > dw {
			scale = float64(dw) / float64(targetW)
		}
		nh := int(math.Round(float64(targetH) * scale))
		nw := int(math.Round(float64(targetW) * scale))
		if nh > dh {
			scale2 := float64(dh) / float64(nh)
			nw = int(math.Round(float64(nw) * scale2))
			nh = dh
		}
		targetW, targetH = max(1, nw), max(1, nh)
	} else if opt.AspectW > 0 && opt.AspectH > 0 {
		// 按比例输出，长边=maxEdge
		aw, ah := float64(opt.AspectW), float64(opt.AspectH)
		if aw >= ah {
			targetW = maxEdge
			targetH = int(math.Round(float64(maxEdge) * ah / aw))
		} else {
			targetH = maxEdge
			targetW = int(math.Round(float64(maxEdge) * aw / ah))
		}
	} else if w > maxEdge || h > maxEdge {
		if w >= h {
			targetW = maxEdge
			targetH = int(math.Round(float64(h) * float64(maxEdge) / float64(w)))
		} else {
			targetH = maxEdge
			targetW = int(math.Round(float64(w) * float64(maxEdge) / float64(h)))
		}
	}

	if targetW == w && targetH == h {
		return cropped
	}
	return imaging.Resize(cropped, max(1, targetW), max(1, targetH), imaging.Lanczos)
}

func resolveCropRect(w, h int, opt ImageAdaptOptions) (cw, ch, cx, cy int) {
	aw, ah := float64(opt.AspectW), float64(opt.AspectH)
	if opt.CropW > 0.001 && opt.CropH > 0.001 {
		cx = int(math.Round(clamp01(opt.CropX) * float64(w)))
		cy = int(math.Round(clamp01(opt.CropY) * float64(h)))
		cw = int(math.Round(clamp01(opt.CropW) * float64(w)))
		ch = int(math.Round(clamp01(opt.CropH) * float64(h)))
		if cw < 1 {
			cw = 1
		}
		if ch < 1 {
			ch = 1
		}
		if cx+cw > w {
			cx = max(0, w-cw)
		}
		if cy+ch > h {
			cy = max(0, h-ch)
		}
		return
	}
	// 居中最大内接目标比例
	if float64(w)/float64(h) > aw/ah {
		ch = h
		cw = int(math.Round(float64(h) * aw / ah))
	} else {
		cw = w
		ch = int(math.Round(float64(w) * ah / aw))
	}
	if cw > w {
		cw = w
	}
	if ch > h {
		ch = h
	}
	cx = (w - cw) / 2
	cy = (h - ch) / 2
	return
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
