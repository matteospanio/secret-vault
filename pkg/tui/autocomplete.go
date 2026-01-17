package tui

import (
	"sort"
	"strings"
)

// FuzzyMatch returns true if pattern matches text using fuzzy matching.
// Fuzzy matching means all characters in pattern appear in text in order,
// but not necessarily contiguously. Matching is case-insensitive.
func FuzzyMatch(pattern, text string) bool {
	if pattern == "" {
		return true
	}
	if text == "" {
		return false
	}

	pattern = strings.ToLower(pattern)
	text = strings.ToLower(text)

	patternIdx := 0
	for i := 0; i < len(text) && patternIdx < len(pattern); i++ {
		if text[i] == pattern[patternIdx] {
			patternIdx++
		}
	}

	return patternIdx == len(pattern)
}

// FuzzyScore returns a score indicating how well pattern matches text.
// Higher scores indicate better matches. Returns 0 or negative for no match.
// Scoring priorities:
// - Exact match: highest score
// - Prefix match: high score
// - Substring match: medium score
// - Fuzzy match: lower score
func FuzzyScore(pattern, text string) int {
	if pattern == "" {
		return 1 // Empty pattern matches everything with minimal score
	}
	if text == "" {
		return 0
	}

	patternLower := strings.ToLower(pattern)
	textLower := strings.ToLower(text)

	// Exact match - highest priority
	if textLower == patternLower {
		return 1000 + len(pattern)*10
	}

	// Prefix match - high priority
	if strings.HasPrefix(textLower, patternLower) {
		return 500 + len(pattern)*5
	}

	// Substring match - medium priority
	if strings.Contains(textLower, patternLower) {
		// Earlier in string = higher score
		idx := strings.Index(textLower, patternLower)
		return 200 + len(pattern)*2 - idx
	}

	// Fuzzy match - check if it matches and calculate score
	if FuzzyMatch(pattern, text) {
		// Score based on how compact the match is
		score := 50
		// Bonus for matching at start
		if len(textLower) > 0 && len(patternLower) > 0 && textLower[0] == patternLower[0] {
			score += 25
		}
		return score
	}

	return 0
}

// FilterSecrets filters a list of secret names based on the query using fuzzy matching.
// Returns all matching secrets.
func FilterSecrets(secrets []string, query string) []string {
	if secrets == nil {
		return []string{}
	}

	if query == "" {
		// Return a copy of the original slice
		result := make([]string, len(secrets))
		copy(result, secrets)
		return result
	}

	var matches []string
	for _, secret := range secrets {
		if FuzzyMatch(query, secret) {
			matches = append(matches, secret)
		}
	}

	if matches == nil {
		return []string{}
	}
	return matches
}

// SortByScore sorts secrets by their fuzzy match score against the query.
// Higher scoring matches appear first. Non-matching secrets are excluded.
func SortByScore(secrets []string, query string) []string {
	if query == "" {
		// Return a copy maintaining original order
		result := make([]string, len(secrets))
		copy(result, secrets)
		return result
	}

	// Create scored items
	type scoredItem struct {
		name  string
		score int
	}

	var scored []scoredItem
	for _, secret := range secrets {
		score := FuzzyScore(query, secret)
		if score > 0 {
			scored = append(scored, scoredItem{name: secret, score: score})
		}
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Extract names
	result := make([]string, len(scored))
	for i, item := range scored {
		result[i] = item.name
	}

	return result
}
