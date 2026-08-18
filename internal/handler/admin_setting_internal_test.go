package handler

import (
	"encoding/json"
	"testing"
)

func TestSettingsDiff(t *testing.T) {
	before := map[string]interface{}{
		"app_name":          "PicFast",
		"allow_guest_upload": false,
		"theme_config":      json.RawMessage(`{"mode":"light"}`),
		"unchanged_nested":  map[string]any{"a": 1},
	}
	after := map[string]interface{}{
		"app_name":          "PicFast Pro",
		"allow_guest_upload": false,
		"theme_config":      json.RawMessage(`{"mode":"dark"}`),
		"unchanged_nested":  map[string]any{"a": 1},
	}

	changes := settingsDiff(before, after)

	if len(changes) != 2 {
		t.Fatalf("changes count = %d, want 2: %v", len(changes), changes)
	}
	if _, ok := changes["allow_guest_upload"]; ok {
		t.Fatalf("unchanged scalar should not appear in changes")
	}
	if _, ok := changes["unchanged_nested"]; ok {
		t.Fatalf("deep-equal nested value should not appear in changes")
	}
	nameChange, ok := changes["app_name"].(map[string]any)
	if !ok || nameChange["before"] != "PicFast" || nameChange["after"] != "PicFast Pro" {
		t.Fatalf("app_name change malformed: %v", changes["app_name"])
	}
	if _, ok := changes["theme_config"]; !ok {
		t.Fatalf("changed nested JSON should be recorded as a whole")
	}
}
