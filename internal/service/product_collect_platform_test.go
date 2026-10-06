package service

import (
	"testing"

	"productcore/internal/dto"
)

func TestDetectCollectPlatform(t *testing.T) {
	ok := []string{
		"https://item.taobao.com/item.htm?id=123",
		"https://detail.tmall.com/item.htm?id=456",
		"https://m.tb.cn/h.abc",
		"http://world.taobao.com/item/1",
	}
	for _, raw := range ok {
		platform, normalized, err := DetectCollectPlatform(raw)
		if err != nil {
			t.Fatalf("%s: %v", raw, err)
		}
		if platform != "taobao" || normalized != raw {
			t.Fatalf("%s => %s %s", raw, platform, normalized)
		}
	}

	bad := []string{
		"",
		"not a url",
		"ftp://item.taobao.com/item.htm",
		"https://item.jd.com/1.html",
		"https://nottaobao.com/item.htm",
		"https://taobao.com.evil.com/item.htm",
	}
	for _, raw := range bad {
		if _, _, err := DetectCollectPlatform(raw); err == nil {
			t.Fatalf("expected reject: %q", raw)
		}
	}
}

func TestIngestFromExtensionEmpty(t *testing.T) {
	s := &ProductCollectService{}
	if _, err := s.IngestFromExtension(1, 1, "", nil, nil); err != ErrCollectEmpty {
		t.Fatalf("nil product: %v", err)
	}
	if _, err := s.IngestFromExtension(1, 1, "https://item.taobao.com/item.htm?id=1", &dto.ProductDTO{Name: "x"}, nil); err != ErrCollectEmpty {
		t.Fatalf("no pics: %v", err)
	}
}
