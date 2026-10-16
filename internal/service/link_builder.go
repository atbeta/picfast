package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/atbeta/picfast/internal/domain"
)

// linkMode reads the strategy's "link_mode" value. Strategy configs also carry
// non-string values (for example "use_path_style": true), so the config must be
// decoded into map[string]any: decoding into map[string]string fails on those
// documents, which used to make an explicit "proxy" silently behave as
// "direct" (image links pointing straight at a private bucket).
func linkMode(configs json.RawMessage) string {
	var raw map[string]any
	if err := json.Unmarshal(configs, &raw); err != nil {
		return ""
	}
	mode, _ := raw["link_mode"].(string)
	return mode
}

// IsProxyLinkMode returns true if the strategy's link_mode is set to "proxy",
// meaning image URLs should be served through PicFast's /i/ proxy route
// instead of direct storage URLs. An absent or empty link_mode defaults to
// "direct" (passthrough).
func IsProxyLinkMode(configs json.RawMessage) bool {
	return linkMode(configs) == "proxy"
}

// IsDirectLinkMode returns true when the strategy's link_mode is explicitly
// set to "direct". This is currently WebDAV-only: it enables 302 redirect at
// serve time instead of proxying content through PicFast. For S3/OSS/COS/Kodo
// strategies, the absence of "proxy" already means direct CDN passthrough at
// link-construction time — do NOT broaden this check to other strategy types
// without understanding the different semantics.
func IsDirectLinkMode(configs json.RawMessage) bool {
	return linkMode(configs) == "direct"
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
