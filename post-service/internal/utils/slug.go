package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Match any character that is not a letter, number, or hyphen
	nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9-]`)
	// Match multiple hyphens
	multipleHyphenRegex = regexp.MustCompile(`-+`)
)

// GenerateSlug creates a URL-friendly slug from a string
func GenerateSlug(s string) string {
	// Convert to lowercase
	slug := strings.ToLower(s)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove all non-alphanumeric characters except hyphens
	slug = nonAlphanumericRegex.ReplaceAllString(slug, "")

	// Replace multiple hyphens with a single hyphen
	slug = multipleHyphenRegex.ReplaceAllString(slug, "-")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	return slug
}

// GenerateUniqueSlug creates a unique slug by appending a number if necessary
func GenerateUniqueSlug(base string, exists func(string) bool) string {
	slug := GenerateSlug(base)
	if !exists(slug) {
		return slug
	}

	// If the slug exists, append a number
	counter := 1
	for {
		newSlug := fmt.Sprintf("%s-%d", slug, counter)
		if !exists(newSlug) {
			return newSlug
		}
		counter++
	}
}
