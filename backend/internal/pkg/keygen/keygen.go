package keygen

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// Slug は文字列をスラッグ化する（小文字、スペースをハイフンに、英数字とハイフンのみ）
func Slug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = regexp.MustCompile(`[^a-z0-9\p{Hiragana}\p{Katakana}\p{Han}\-]+`).ReplaceAllString(s, "-")
	s = regexp.MustCompile(`\-+`).ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return ""
	}
	return s
}

// UUIDKey は UUID の文字列を返す
func UUIDKey(id uuid.UUID) string {
	return id.String()
}

// PrefixedID は "prefix-{id}" 形式を返す
func PrefixedID(prefix string, id uint) string {
	return fmt.Sprintf("%s-%d", prefix, id)
}

// CompositeKey は複合キーを "-" で連結する
func CompositeKey(parts ...string) string {
	return strings.Join(parts, "-")
}
