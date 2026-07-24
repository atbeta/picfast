package service

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

func GeneratePathname(pathRule, fileRule, extension string, fileMD5 string, userID int64) string {
	if pathRule == "" {
		pathRule = "{Y}/{m}/{d}"
	}
	if fileRule == "" {
		fileRule = "{uniqid}"
	}
	path := expandRule(pathRule, fileMD5, userID)
	name := expandRule(fileRule, fileMD5, userID)
	if path == "" || path == "." {
		return name + "." + extension
	}
	return path + "/" + name + "." + extension
}

func expandRule(rule, fileMD5 string, userID int64) string {
	now := time.Now()
	result := rule

	placeholders := map[string]string{
		"{Y}":             now.Format("2006"),
		"{y}":             now.Format("06"),
		"{m}":             now.Format("01"),
		"{d}":             now.Format("02"),
		"{timestamp}":     fmt.Sprintf("%d", now.Unix()),
		"{uniqid}":        generateRandomString(13),
		"{md5}":           fileMD5,
		"{md5-16}":        truncate(fileMD5, 16),
		"{str-random-16}": generateRandomString(16),
		"{str-random-10}": generateRandomString(10),
		"{uid}":           fmt.Sprintf("%d", userID),
	}

	for key, val := range placeholders {
		result = strings.ReplaceAll(result, key, val)
	}

	result = strings.TrimLeft(result, "./")

	return result
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

func GenerateImageKey(length int) string {
	return generateRandomString(length)
}

// MinImageKeyLength is the smallest allowed image key length. Anything below
// this is rejected at generation time and clamped at the lookup layer so the
// invariant is upheld even for misconfigured instances.
const MinImageKeyLength = 4

// MaxImageKeyLength is the largest allowed image key length. The underlying
// charset has 36 symbols, so 10 characters already offer ~3.6e15 combinations
// and effectively eliminate collisions for any realistic deployment.
const MaxImageKeyLength = 10

// ClampImageKeyLength normalises a configured minimum length into the supported
// [MinImageKeyLength, MaxImageKeyLength] range. Values outside the range are
// silently clamped instead of returning an error, so a typo in the config file
// cannot prevent the service from starting.
func ClampImageKeyLength(minLength int) int {
	if minLength < MinImageKeyLength {
		return MinImageKeyLength
	}
	if minLength > MaxImageKeyLength {
		return MaxImageKeyLength
	}
	return minLength
}

// BaseKeyLength returns the recommended key length for the given total image
// count, targeting <0.1% collision probability per generation (<1 retry per
// 1000 uploads). The configured minLength acts as a floor: a deployment that
// always wants long keys can raise it (e.g. 8 or 10) and the tier table no
// longer matters. Values outside [MinImageKeyLength, MaxImageKeyLength] are
// clamped to keep callers safe against misconfiguration.
func BaseKeyLength(totalImages int64, minLength int) int {
	tier := 4
	switch {
	case totalImages < 1680:
		tier = 4
	case totalImages < 60466:
		tier = 5
	case totalImages < 2176782:
		tier = 6
	case totalImages < 78364164:
		tier = 7
	default:
		tier = 8
	}
	min := ClampImageKeyLength(minLength)
	if tier < min {
		return min
	}
	return tier
}

func ComputeMD5(data []byte) string {
	h := md5.Sum(data)
	return hex.EncodeToString(h[:])
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
