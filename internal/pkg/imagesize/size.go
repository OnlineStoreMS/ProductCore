package imagesize

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
)

const (
	minPixels = 512 * 512
	maxPixels = 2048 * 2048
	maxAspect = 8
)

// FromReader 读取图片宽高，支持 jpeg、png、gif、webp。
func FromReader(r io.Reader) (int, int, error) {
	var head [32]byte
	n, err := io.ReadFull(r, head[:])
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return 0, 0, err
	}
	if w, h, ok := webpDimensions(head[:n]); ok {
		return w, h, nil
	}
	cfg, _, err := image.DecodeConfig(io.MultiReader(bytes.NewReader(head[:n]), r))
	if err != nil {
		return 0, 0, err
	}
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return 0, 0, errors.New("invalid image size")
	}
	return cfg.Width, cfg.Height, nil
}

// QwenSize 把原图宽高收成 qwen-image 允许的输出尺寸（宽x高）。
// 原图落在范围内时保持原尺寸；过大或过小则按比例缩放到总像素 512×512 至 2048×2048，宽高比不超过 8:1。
func QwenSize(width, height int) string {
	w, h := fit(width, height)
	return fmt.Sprintf("%dx%d", w, h)
}

func fit(width, height int) (int, int) {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	w := float64(width)
	h := float64(height)
	if w > h*maxAspect {
		w = h * maxAspect
	} else if h > w*maxAspect {
		h = w * maxAspect
	}
	area := w * h
	switch {
	case area > maxPixels:
		scale := math.Sqrt(maxPixels / area)
		w *= scale
		h *= scale
	case area < minPixels:
		scale := math.Sqrt(minPixels / area)
		w *= scale
		h *= scale
	}
	iw := int(math.Round(w))
	ih := int(math.Round(h))
	if iw < 1 {
		iw = 1
	}
	if ih < 1 {
		ih = 1
	}
	if iw > ih*maxAspect {
		iw = ih * maxAspect
	}
	if ih > iw*maxAspect {
		ih = iw * maxAspect
	}
	for iw*ih > maxPixels && (iw > 1 || ih > 1) {
		if iw >= ih && iw > 1 {
			iw--
			continue
		}
		if ih > 1 {
			ih--
		}
	}
	for iw*ih < minPixels {
		if float64(iw)/float64(ih) <= 1 {
			iw++
		} else {
			ih++
		}
		if iw > ih*maxAspect {
			iw = ih * maxAspect
		}
		if ih > iw*maxAspect {
			ih = iw * maxAspect
		}
		if iw*ih >= minPixels || iw > 4096 || ih > 4096 {
			break
		}
	}
	return iw, ih
}

func webpDimensions(b []byte) (int, int, bool) {
	if len(b) < 30 || string(b[0:4]) != "RIFF" || string(b[8:12]) != "WEBP" {
		return 0, 0, false
	}
	switch string(b[12:16]) {
	case "VP8X":
		w := 1 + int(b[24]) + int(b[25])<<8 + int(b[26])<<16
		h := 1 + int(b[27]) + int(b[28])<<8 + int(b[29])<<16
		return w, h, w > 0 && h > 0
	case "VP8 ":
		w := int(uint16(b[26]) | uint16(b[27])<<8)
		h := int(uint16(b[28]) | uint16(b[29])<<8)
		w &= 0x3fff
		h &= 0x3fff
		return w, h, w > 0 && h > 0
	case "VP8L":
		if len(b) < 25 {
			return 0, 0, false
		}
		bits := uint32(b[21]) | uint32(b[22])<<8 | uint32(b[23])<<16 | uint32(b[24])<<24
		w := int(bits&0x3FFF) + 1
		h := int((bits>>14)&0x3FFF) + 1
		return w, h, true
	default:
		return 0, 0, false
	}
}
