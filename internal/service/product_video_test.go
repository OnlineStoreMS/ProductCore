package service

import (
	"testing"

	"productcore/internal/dto"
)

func TestIsCollectableVideoURL(t *testing.T) {
	ok := []string{
		"https://cloud.video.taobao.com/play/u/foo/p/1/e/6/t/1/bar.mp4",
		"https://video.alicdn.com/a.mp4?foo=1",
		"https://cloudvideo.taobao.com/play/x",
	}
	for _, u := range ok {
		if !isCollectableVideoURL(u) {
			t.Fatalf("want collectable: %s", u)
		}
	}
	bad := []string{"", "not-a-url", "https://img.alicdn.com/a.jpg", "https://x.com/a.m3u8"}
	for _, u := range bad {
		if isCollectableVideoURL(u) {
			t.Fatalf("want skip: %s", u)
		}
	}
}

func TestCollectRemoteVideoURLsDedup(t *testing.T) {
	u := "https://cloud.video.taobao.com/a.mp4"
	got := collectRemoteVideoURLs(&dto.ProductDTO{
		ProductVideo: u,
		Media:        &dto.ProductMediaDTO{Videos: &dto.ProductVideosDTO{Ratio11: u}},
	}, []string{u, "https://img.alicdn.com/x.jpg"})
	if len(got) != 1 || got[0] != u {
		t.Fatalf("got %#v", got)
	}
}

func TestAssignVideoByRatio(t *testing.T) {
	dst := &dto.ProductVideosDTO{}
	if err := assignVideoByRatio(dst, "1:1", "u1"); err != nil {
		t.Fatal(err)
	}
	if err := assignVideoByRatio(dst, "3:4", "u2"); err != nil {
		t.Fatal(err)
	}
	if dst.Ratio11 != "u1" || dst.Ratio34 != "u2" {
		t.Fatalf("%+v", dst)
	}
	if err := assignVideoByRatio(dst, "1:1", "u3"); err == nil {
		t.Fatal("expected duplicate 1:1 error")
	}
}
