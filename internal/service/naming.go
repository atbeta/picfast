package service

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

func GeneratePathname(pathRule, fileRule, extension string, fileMD5 string, userID int64) string {
	path := expandRule(pathRule, fileMD5, userID)
	name := expandRule(fileRule, fileMD5, userID)
	return path + "/" + name + "." + extension
}

func expandRule(rule, fileMD5 string, userID int64) string {
	now := time.Now()
	result := rule

	placeholders := map[string]string{
		"{Y}":            now.Format("2006"),
		"{y}":            now.Format("06"),
		"{m}":            now.Format("01"),
		"{d}":            now.Format("02"),
		"{timestamp}":    fmt.Sprintf("%d", now.Unix()),
		"{uniqid}":       generateRandomString(13),
		"{md5}":          fileMD5,
		"{md5-16}":       truncate(fileMD5, 16),
		"{str-random-16}": generateRandomString(16),
		"{str-random-10}": generateRandomString(10),
		"{uid}":          fmt.Sprintf("%d", userID),
	}

	for key, val := range placeholders {
		result = strings.ReplaceAll(result, key, val)
	}

	return filepath.Clean(result)
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

func GenerateImageKey() string {
	return generateRandomString(6)
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
