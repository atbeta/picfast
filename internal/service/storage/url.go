package storage

import "strings"

func joinPublicURL(base, pathname string) string {
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(pathname, "/")
}
