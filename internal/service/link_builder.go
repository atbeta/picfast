package service

import (
	"fmt"
	"strings"

	"github.com/atbeta/picfast/internal/domain"
)

type LinkBuilder struct {
	BaseURL string
}

func (b LinkBuilder) BuildImageLinks(key, extension, md5, originName string) domain.ImageLinks {
	baseURL := strings.TrimRight(b.BaseURL, "/")
	url := fmt.Sprintf("%s/i/%s.%s", baseURL, key, extension)

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
