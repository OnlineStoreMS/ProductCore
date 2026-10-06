package service

import "testing"

func TestNeedsRehost(t *testing.T) {
	if !needsRehost("https://gw.alicdn.com/bao/uploaded/O1CN01iybyOD2AjyYEQsVbY_!!1726878240.jpg") {
		t.Fatal("alicdn should be rehosted")
	}
	if needsRehost("https://osms.zfcycle.com/apps/product/uploads/products/1/main/a.jpg") {
		t.Fatal("own url should stay")
	}
	if needsRehost("") {
		t.Fatal("empty")
	}
}

func TestRewriteCollectImgSrc(t *testing.T) {
	html := `<p><img src="https://img.alicdn.com/a.jpg" /></p><p><img src="https://osms.example.com/b.jpg" /></p>`
	got := rewriteCollectImgSrc(html, func(raw string) string {
		if needsRehost(raw) {
			return "https://minio.example/uploads/a.jpg"
		}
		return raw
	})
	want := `<p><img src="https://minio.example/uploads/a.jpg" /></p><p><img src="https://osms.example.com/b.jpg" /></p>`
	if got != want {
		t.Fatalf("got %s", got)
	}
}
