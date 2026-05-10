package maintenance

import "testing"

func TestJoinObjectPath(t *testing.T) {
	tests := []struct {
		dir  string
		name string
		want string
	}{
		{dir: "", name: "cat.png", want: "cat.png"},
		{dir: "2026/05", name: "cat.png", want: "2026/05/cat.png"},
		{dir: "/2026/05/", name: "/cat.png", want: "2026/05/cat.png"},
	}

	for _, tt := range tests {
		if got := joinObjectPath(tt.dir, tt.name); got != tt.want {
			t.Fatalf("joinObjectPath(%q, %q) = %q, want %q", tt.dir, tt.name, got, tt.want)
		}
	}
}

func TestInventoryItemThumbnailName(t *testing.T) {
	item := InventoryItem{Extension: "png", MD5: "abc"}
	if got := item.ThumbnailName(); got != "abc.png" {
		t.Fatalf("ThumbnailName() = %q", got)
	}

	item.Extension = "svg"
	if got := item.ThumbnailName(); got != "" {
		t.Fatalf("ThumbnailName() for svg = %q, want empty", got)
	}
}
