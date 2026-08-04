package mediainfo

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"os"
)

const aspectTolerance = 0.05

// ImageDimensions 读取图片宽高
func ImageDimensions(path string) (int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, err
	}
	return cfg.Width, cfg.Height, nil
}

// VideoDimensions 读取 mp4 视频宽高（解析 tkhd box）
func VideoDimensions(path string) (int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return 0, 0, err
	}
	return readMP4Dimensions(f, stat.Size())
}

func readMP4Dimensions(r io.Reader, total int64) (int, int, error) {
	var width, height int
	err := walkBoxes(r, total, func(boxType string, payload []byte) error {
		if boxType != "tkhd" {
			return nil
		}
		w, h, err := parseTkhd(payload)
		if err != nil {
			return nil
		}
		if w > 0 && h > 0 {
			width, height = w, h
		}
		return nil
	})
	if err != nil {
		return 0, 0, err
	}
	if width <= 0 || height <= 0 {
		return 0, 0, errors.New("无法读取视频尺寸")
	}
	return width, height, nil
}

func parseTkhd(payload []byte) (int, int, error) {
	if len(payload) < 8 {
		return 0, 0, errors.New("tkhd too short")
	}
	version := payload[0]
	var widthOff, heightOff int
	if version == 1 {
		widthOff, heightOff = 88, 96
	} else {
		widthOff, heightOff = 76, 80
	}
	if len(payload) < heightOff+4 {
		return 0, 0, errors.New("tkhd truncated")
	}
	w := int(binary.BigEndian.Uint32(payload[widthOff : widthOff+4]) >> 16)
	h := int(binary.BigEndian.Uint32(payload[heightOff : heightOff+4]) >> 16)
	return w, h, nil
}

func isContainer(boxType string) bool {
	switch boxType {
	case "moov", "trak", "mdia", "minf", "stbl", "meta", "edts":
		return true
	default:
		return false
	}
}

func walkBoxes(r io.Reader, limit int64, fn func(boxType string, payload []byte) error) error {
	var read int64
	for read+8 <= limit {
		var hdr [8]byte
		if _, err := io.ReadFull(r, hdr[:]); err != nil {
			return err
		}
		read += 8
		boxSize := int64(binary.BigEndian.Uint32(hdr[0:4]))
		boxType := string(hdr[4:8])
		headerSize := int64(8)
		if boxSize == 1 {
			var ext [8]byte
			if _, err := io.ReadFull(r, ext[:]); err != nil {
				return err
			}
			read += 8
			headerSize = 16
			boxSize = int64(binary.BigEndian.Uint64(ext[:]))
		}
		if boxSize == 0 {
			boxSize = limit - read + headerSize
		}
		if boxSize < headerSize {
			return fmt.Errorf("invalid mp4 box %s", boxType)
		}
		contentSize := boxSize - headerSize
		if read+contentSize > limit {
			contentSize = limit - read
		}
		if isContainer(boxType) {
			sub := io.LimitReader(r, contentSize)
			if err := walkBoxes(sub, contentSize, fn); err != nil {
				return err
			}
		} else {
			payload := make([]byte, contentSize)
			if contentSize > 0 {
				if _, err := io.ReadFull(r, payload); err != nil {
					return err
				}
			}
			if err := fn(boxType, payload); err != nil {
				return err
			}
		}
		read += contentSize
	}
	return nil
}

// DetectVideoRatio 识别视频宽高比：1:1 / 3:4 / 16:9 / 9:16
func DetectVideoRatio(width, height int) (string, error) {
	if width <= 0 || height <= 0 {
		return "", errors.New("无法读取视频尺寸")
	}
	actual := float64(width) / float64(height)
	targets := []struct {
		label string
		ratio float64
	}{
		{"1:1", 1.0},
		{"3:4", 0.75},
		{"16:9", 16.0 / 9.0},
		{"9:16", 9.0 / 16.0},
	}
	bestLabel := ""
	bestRel := math.MaxFloat64
	for _, t := range targets {
		rel := math.Abs(actual-t.ratio) / t.ratio
		if rel <= aspectTolerance && rel < bestRel {
			bestRel = rel
			bestLabel = t.label
		}
	}
	if bestLabel != "" {
		return bestLabel, nil
	}
	return "", fmt.Errorf("视频宽高比需为 1:1、3:4、16:9 或 9:16（当前 %d×%d）", width, height)
}

// MatchVideoRatio 兼容旧调用：识别主图/商品视频支持的比例。
func MatchVideoRatio(width, height int) (string, error) {
	return DetectVideoRatio(width, height)
}
