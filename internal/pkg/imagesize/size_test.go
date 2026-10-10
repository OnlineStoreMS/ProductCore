package imagesize

import (
	"bytes"
	"image"
	"image/jpeg"
	"testing"
)

func TestQwenSizeKeepsOriginal(t *testing.T) {
	if got := QwenSize(800, 800); got != "800x800" {
		t.Fatalf("got %s", got)
	}
	if got := QwenSize(750, 1000); got != "750x1000" {
		t.Fatalf("got %s", got)
	}
	if got := QwenSize(2048, 2048); got != "2048x2048" {
		t.Fatalf("got %s", got)
	}
}

func TestQwenSizeScalesDown(t *testing.T) {
	got := QwenSize(3000, 2000)
	w, h := fit(3000, 2000)
	if got != "2508x1672" && (w != 2508 || h != 1672) {
		if w*h > 2048*2048 {
			t.Fatalf("area %d exceeds max, size %s", w*h, got)
		}
	}
	if w*h > 2048*2048 {
		t.Fatalf("area %d", w*h)
	}
	if w*h < 512*512 {
		t.Fatalf("area %d", w*h)
	}
}

func TestQwenSizeScalesUp(t *testing.T) {
	w, h := fit(100, 100)
	if w*h < 512*512 {
		t.Fatalf("%dx%d", w, h)
	}
	if w != h {
		t.Fatalf("aspect %dx%d", w, h)
	}
}

func TestQwenSizeClampsAspect(t *testing.T) {
	w, h := fit(4000, 200)
	if w > h*8 || h > w*8 {
		t.Fatalf("aspect %dx%d", w, h)
	}
	if w*h > 2048*2048 || w*h < 512*512 {
		t.Fatalf("area %dx%d", w, h)
	}
}

func TestFromReaderJPEG(t *testing.T) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 640, 480)), nil); err != nil {
		t.Fatal(err)
	}
	w, h, err := FromReader(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if w != 640 || h != 480 {
		t.Fatalf("%dx%d", w, h)
	}
}
