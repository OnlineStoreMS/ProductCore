package storage

import (
	"testing"

	"productcore/internal/config"
)

func TestPublicURLResolver_Resolve(t *testing.T) {
	r := NewPublicURLResolver(&config.StorageConfig{
		PublicBaseURL: "http://192.168.3.41:9100/productcore",
		Prefix:        "uploads",
	})

	stored := "http://127.0.0.1:9100/productcore/uploads/products/42/main/a.jpg"
	want := "http://192.168.3.41:9100/productcore/uploads/products/42/main/a.jpg"
	if got := r.Resolve(stored); got != want {
		t.Fatalf("Resolve() = %q, want %q", got, want)
	}
}

func TestPublicURLResolver_ResolveExternal(t *testing.T) {
	r := NewPublicURLResolver(&config.StorageConfig{
		PublicBaseURL: "http://192.168.3.41:9100/productcore",
		Prefix:        "uploads",
	})
	external := "https://cdn.example.com/logo.png"
	if got := r.Resolve(external); got != external {
		t.Fatalf("external URL should be unchanged, got %q", got)
	}
}
