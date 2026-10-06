package taobaoimg

import "testing"

func TestUnwrap(t *testing.T) {
	orig := "http://img.alicdn.com/imgextra/i1/1726878240/O1CN01m4BOvc2AjyQTRKuZm_!!1726878240.jpg"
	cases := []struct {
		in, want string
	}{
		{orig, orig},
		{orig + "_.webp", orig},
		{orig + "_400x400.jpg_.webp", orig},
		{orig + "_400x400q90.jpg_.webp", orig},
		{orig + "_Q90.jpg_.webp", orig},
		{"//img.alicdn.com/imgextra/i1/x/foo.jpg_.webp", "http://img.alicdn.com/imgextra/i1/x/foo.jpg"},
		{"https://img.alicdn.com/imgextra/i1/x/foo.png_.webp", "https://img.alicdn.com/imgextra/i1/x/foo.png"},
		{"", ""},
	}
	for _, c := range cases {
		got := Unwrap(c.in)
		if got != c.want {
			t.Fatalf("Unwrap(%q)\n got %q\nwant %q", c.in, got, c.want)
		}
	}
}
