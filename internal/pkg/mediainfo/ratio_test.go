package mediainfo_test

import (
	"testing"

	"productcore/internal/pkg/mediainfo"
)

func TestDetectVideoRatio(t *testing.T) {
	cases := []struct {
		w, h int
		want string
	}{
		{800, 800, "1:1"},
		{750, 1000, "3:4"},
		{1280, 720, "16:9"},
		{1920, 1080, "16:9"},
		{1080, 1920, "9:16"},
		{720, 1280, "9:16"},
	}
	for _, tc := range cases {
		got, err := mediainfo.DetectVideoRatio(tc.w, tc.h)
		if err != nil {
			t.Fatalf("DetectVideoRatio(%d,%d): %v", tc.w, tc.h, err)
		}
		if got != tc.want {
			t.Fatalf("DetectVideoRatio(%d,%d)=%q want %q", tc.w, tc.h, got, tc.want)
		}
	}
}

func TestDetectVideoRatioUnsupported(t *testing.T) {
	if _, err := mediainfo.DetectVideoRatio(1024, 768); err == nil {
		t.Fatal("expected error for 4:3")
	}
}
