package maintenance

import "testing"

func TestManifestValidate(t *testing.T) {
	manifest := NewManifest("0.8.0", 24)
	manifest.Objects.Count = 3
	manifest.Objects.Bytes = 128

	if err := manifest.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestManifestValidateRejectsUnsupportedFormat(t *testing.T) {
	manifest := NewManifest("0.8.0", 24)
	manifest.Format = "other"

	if err := manifest.Validate(); err == nil {
		t.Fatal("Validate() expected error")
	}
}
