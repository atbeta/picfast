package handler

import (
	"os"
	"path/filepath"
	"testing"
)

// Thumbnails live in each image's storage strategy, so the dashboard badge must
// follow the strategies and not the legacy on-disk directory (which is normally
// absent, and used to make every deployment report 未启用).
func TestThumbnailStoreHealth(t *testing.T) {
	cases := []struct {
		name          string
		strategies    []map[string]any
		wantStatus    string
		wantHealthy   bool
		wantHealthyNo int
	}{
		{name: "no strategies", wantStatus: "disabled"},
		{name: "all healthy", strategies: []map[string]any{{"healthy": true}}, wantHealthy: true, wantHealthyNo: 1},
		{name: "all unhealthy", strategies: []map[string]any{{"healthy": false}}, wantHealthy: false},
		{name: "one unhealthy one healthy", strategies: []map[string]any{{"healthy": false}, {"healthy": true}}, wantHealthy: true, wantHealthyNo: 1},
		{name: "missing healthy key", strategies: []map[string]any{{"error": "failed to init"}}, wantHealthy: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := thumbnailStoreHealth(tc.strategies, "")

			if tc.wantStatus != "" {
				if got["status"] != tc.wantStatus {
					t.Fatalf("status = %v, want %v", got["status"], tc.wantStatus)
				}
				return
			}
			if got["healthy"] != tc.wantHealthy {
				t.Fatalf("healthy = %v, want %v", got["healthy"], tc.wantHealthy)
			}
			if got["strategies"] != len(tc.strategies) {
				t.Fatalf("strategies = %v, want %d", got["strategies"], len(tc.strategies))
			}
			if tc.wantHealthyNo > 0 && got["strategies_healthy"] != tc.wantHealthyNo {
				t.Fatalf("strategies_healthy = %v, want %d", got["strategies_healthy"], tc.wantHealthyNo)
			}
			if got["thumbnail_prefix"] != "_picfast/thumbnails/" {
				t.Fatalf("thumbnail_prefix = %v, want _picfast/thumbnails/", got["thumbnail_prefix"])
			}
		})
	}
}

func TestThumbnailStoreHealthReportsLegacyDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "legacy"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	got := thumbnailStoreHealth([]map[string]any{{"healthy": true}}, filepath.Join(dir, "legacy"))
	if got["legacy_dir"] == nil {
		t.Fatal("legacy_dir missing")
	}
	if got["legacy_dir_present"] != true {
		t.Fatalf("legacy_dir_present = %v, want true", got["legacy_dir_present"])
	}

	missing := thumbnailStoreHealth([]map[string]any{{"healthy": true}}, filepath.Join(dir, "absent"))
	if _, ok := missing["legacy_dir_present"]; ok {
		t.Fatalf("legacy_dir_present = %v, want absent", missing["legacy_dir_present"])
	}
}
