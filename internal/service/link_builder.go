package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/atbeta/picfast/internal/domain"
)

// IsProxyLinkMode returns true if the strategy's link_mode is set to "proxy",
// meaning image URLs should be served through PicFast's /i/ proxy route
// instead of direct storage URLs. An absent or empty link_mode defaults to
// "direct" (passthrough).
func IsProxyLinkMode(configs json.RawMessage) bool {
	var raw map[string]string
	if err := json.Unmarshal(configs, &raw); err != nil {
		return false
	}
	return raw["link_mode"] == "proxy"
}

// IsDirectLinkMode returns true when the strategy's link_mode is explicitly
// set to "direct". This is currently WebDAV-only: it enables 302 redirect at
// serve time instead of proxying content through PicFast. For S3/OSS/COS/Kodo
// strategies, the absence of "proxy" already means direct CDN passthrough at
// link-construction time — do NOT broaden this check to other strategy types
// without understanding the different semantics.
func IsDirectLinkMode(configs json.RawMessage) bool {
	var raw map[string]string
	if err := json.Unmarshal(configs, &raw); err != nil {
		return false
	}
	return raw["link_mode"] == "direct"
}

type LinkBuilder struct {
	BaseURL     string
	StrategyURL string
}

func (b LinkBuilder) BuildImageLinks(key, extension, md5, originName string) domain.ImageLinks {
	baseURL := strings.TrimRight(b.BaseURL, "/")

	var url string
	if b.StrategyURL != "" {
		url = b.StrategyURL
	} else {
		url = fmt.Sprintf("%s/i/%s.%s", baseURL, key, extension)
	}

	var thumbURL string
	ext := strings.ToLower(extension)
	if ext != "svg" && ext != "ico" {
		thumbURL = fmt.Sprintf("%s/t/%s.png", baseURL, md5)
	}

	return domain.ImageLinks{
		URL:          url,
		HTML:         fmt.Sprintf(`<img src="%s" alt="%s" />`, url, originName),
		BBCode:       fmt.Sprintf("[img]%s[/img]", url),
		Markdown:     fmt.Sprintf("![%s](%s)", originName, url),
		ThumbnailURL: thumbURL,
	}
}
