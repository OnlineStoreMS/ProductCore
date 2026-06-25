package httputil_test

import (
	"net/url"
	"strings"
	"testing"

	"productcore/internal/pkg/httputil"
)

func TestContentDispositionAttachmentUTF8(t *testing.T) {
	got := httputil.ContentDispositionAttachment("商品中心_商品ID_51.zip", "productcore_51.zip")
	if !strings.Contains(got, `filename="productcore_51.zip"`) {
		t.Fatalf("missing ascii fallback: %q", got)
	}
	if !strings.Contains(got, "filename*=UTF-8''") {
		t.Fatalf("missing filename*: %q", got)
	}
	encoded := strings.TrimPrefix(strings.Split(got, "filename*=UTF-8''")[1], "")
	decoded, err := url.PathUnescape(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded != "商品中心_商品ID_51.zip" {
		t.Fatalf("decoded = %q", decoded)
	}
}
