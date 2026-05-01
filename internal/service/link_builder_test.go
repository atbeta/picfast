package service

import "testing"

func TestLinkBuilderBuildImageLinks(t *testing.T) {
	builder := LinkBuilder{BaseURL: "https://img.example.com"}

	links := builder.BuildImageLinks("abc123", "png", "md5hash", "cat.png")

	if links.URL != "https://img.example.com/i/abc123.png" {
		t.Fatalf("URL = %q", links.URL)
	}
	if links.ThumbnailURL != "https://img.example.com/t/md5hash.png" {
		t.Fatalf("ThumbnailURL = %q", links.ThumbnailURL)
	}
	if links.HTML != `<img src="https://img.example.com/i/abc123.png" alt="cat.png" />` {
		t.Fatalf("HTML = %q", links.HTML)
	}
	if links.BBCode != "[img]https://img.example.com/i/abc123.png[/img]" {
		t.Fatalf("BBCode = %q", links.BBCode)
	}
	if links.Markdown != "![cat.png](https://img.example.com/i/abc123.png)" {
		t.Fatalf("Markdown = %q", links.Markdown)
	}
}
