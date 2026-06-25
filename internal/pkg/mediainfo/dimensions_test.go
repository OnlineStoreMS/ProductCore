package mediainfo_test

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"productcore/internal/pkg/mediainfo"
)

func TestVideoDimensionsSample(t *testing.T) {
	zipPath := "../../../data/show/淘宝_商品ID_723068433145.zip"
	videoPath := extractSampleVideo(t, zipPath)
	w, h, err := mediainfo.VideoDimensions(videoPath)
	if err != nil {
		t.Fatalf("VideoDimensions: %v", err)
	}
	ratio, err := mediainfo.MatchVideoRatio(w, h)
	if err != nil {
		t.Fatalf("MatchVideoRatio(%d,%d): %v", w, h, err)
	}
	if ratio != "1:1" && ratio != "3:4" {
		t.Fatalf("unexpected ratio %q", ratio)
	}
}

func extractSampleVideo(t *testing.T, zipPath string) string {
	t.Helper()
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	for _, f := range r.File {
		if filepath.Base(f.Name) == "视频01.mp4" {
			tmp, err := os.CreateTemp("", "video-*.mp4")
			if err != nil {
				t.Fatal(err)
			}
			rc, err := f.Open()
			if err != nil {
				t.Fatal(err)
			}
			if _, err := io.Copy(tmp, rc); err != nil {
				rc.Close()
				t.Fatal(err)
			}
			rc.Close()
			tmp.Close()
			t.Cleanup(func() { _ = os.Remove(tmp.Name()) })
			return tmp.Name()
		}
	}
	t.Fatal("sample video not found")
	return ""
}
