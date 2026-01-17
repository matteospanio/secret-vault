package tui

import (
	"testing"
)

// --- Fuzzy Matching Tests ---

func TestFuzzyMatch(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		text     string
		expected bool
	}{
		// Exact matches
		{"exact match", "github", "github", true},
		{"exact match uppercase", "GITHUB", "github", true},

		// Prefix matches
		{"prefix match", "git", "github-token", true},
		{"prefix match case insensitive", "GIT", "github-token", true},

		// Substring matches
		{"substring match", "hub", "github", true},
		{"substring at end", "token", "github-token", true},

		// Fuzzy matches (non-contiguous)
		{"fuzzy match", "ghb", "github", true},
		{"fuzzy match with gaps", "gtok", "github-token", true},
		{"fuzzy match letters spread", "gttk", "github-token", true},

		// No matches
		{"no match", "xyz", "github", false},
		{"no match wrong order", "bga", "github", false},
		{"partial no match", "githubx", "github", false},

		// Edge cases
		{"empty pattern matches all", "", "github", true},
		{"empty text no match", "a", "", false},
		{"both empty", "", "", true},
		{"single char match", "g", "github", true},
		{"single char no match", "x", "github", false},

		// Special characters
		{"match with hyphen", "api-key", "my-api-key", true},
		{"match with underscore", "api_key", "my_api_key", true},
		{"match numbers", "123", "secret123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FuzzyMatch(tt.pattern, tt.text)
			if result != tt.expected {
				t.Errorf("FuzzyMatch(%q, %q) = %v, want %v", tt.pattern, tt.text, result, tt.expected)
			}
		})
	}
}

// --- FilterSecrets Tests ---

func TestFilterSecrets(t *testing.T) {
	secrets := []string{
		"github-token",
		"gitlab-api-key",
		"aws-secret",
		"azure-key",
		"personal-github",
	}

	tests := []struct {
		name     string
		query    string
		expected []string
	}{
		{
			name:     "empty query returns all",
			query:    "",
			expected: secrets,
		},
		{
			name:  "prefix filter includes fuzzy matches",
			query: "git",
			// "git" fuzzy matches: github-token, gitlab-api-key, AND personal-github (g-i-t)
			expected: []string{"github-token", "gitlab-api-key", "personal-github"},
		},
		{
			name:  "case insensitive filter",
			query: "GIT",
			// Same as above, case insensitive
			expected: []string{"github-token", "gitlab-api-key", "personal-github"},
		},
		{
			name:     "substring filter",
			query:    "key",
			expected: []string{"gitlab-api-key", "azure-key"},
		},
		{
			name:     "fuzzy filter",
			query:    "ghb",
			expected: []string{"github-token", "personal-github"},
		},
		{
			name:     "no matches",
			query:    "xyz",
			expected: []string{},
		},
		{
			name:     "single result",
			query:    "aws",
			expected: []string{"aws-secret"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterSecrets(secrets, tt.query)
			if len(result) != len(tt.expected) {
				t.Errorf("FilterSecrets(%q) returned %d items, want %d", tt.query, len(result), len(tt.expected))
				t.Errorf("got: %v", result)
				t.Errorf("want: %v", tt.expected)
				return
			}
			for i, r := range result {
				if r != tt.expected[i] {
					t.Errorf("FilterSecrets(%q)[%d] = %q, want %q", tt.query, i, r, tt.expected[i])
				}
			}
		})
	}
}

func TestFilterSecretsEmptyList(t *testing.T) {
	result := FilterSecrets([]string{}, "test")
	if len(result) != 0 {
		t.Errorf("FilterSecrets with empty list should return empty, got %v", result)
	}
}

func TestFilterSecretsNilList(t *testing.T) {
	result := FilterSecrets(nil, "test")
	if result == nil {
		t.Error("FilterSecrets should return empty slice, not nil")
	}
	if len(result) != 0 {
		t.Errorf("FilterSecrets with nil list should return empty, got %v", result)
	}
}

// --- FuzzyScore Tests ---

func TestFuzzyScore(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		text        string
		shouldScore bool
		minScore    int // minimum expected score (0 = any positive)
	}{
		// Exact match should score highest
		{"exact match high score", "github", "github", true, 100},

		// Prefix matches score higher than substring
		{"prefix scores well", "git", "github", true, 50},

		// Substring matches
		{"substring scores", "hub", "github", true, 1},

		// Fuzzy matches score lower
		{"fuzzy lower score", "ghb", "github", true, 1},

		// No match should return 0 or negative
		{"no match zero score", "xyz", "github", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := FuzzyScore(tt.pattern, tt.text)
			if tt.shouldScore && score < tt.minScore {
				t.Errorf("FuzzyScore(%q, %q) = %d, want >= %d", tt.pattern, tt.text, score, tt.minScore)
			}
			if !tt.shouldScore && score > 0 {
				t.Errorf("FuzzyScore(%q, %q) = %d, want <= 0", tt.pattern, tt.text, score)
			}
		})
	}
}

func TestFuzzyScoreOrdering(t *testing.T) {
	// Exact match should score higher than prefix
	exactScore := FuzzyScore("github", "github")
	prefixScore := FuzzyScore("git", "github")
	if exactScore <= prefixScore {
		t.Errorf("Exact match score (%d) should be > prefix score (%d)", exactScore, prefixScore)
	}

	// Prefix should score higher than fuzzy
	fuzzyScore := FuzzyScore("ghb", "github")
	if prefixScore <= fuzzyScore {
		t.Errorf("Prefix score (%d) should be > fuzzy score (%d)", prefixScore, fuzzyScore)
	}
}

// --- SortByScore Tests ---

func TestSortByScore(t *testing.T) {
	secrets := []string{
		"personal-github",
		"github-token",
		"github",
		"gitlab-api",
	}

	sorted := SortByScore(secrets, "github")

	// "github" should be first (exact match)
	if sorted[0] != "github" {
		t.Errorf("First result should be exact match 'github', got %q", sorted[0])
	}

	// "github-token" should be before "personal-github" (prefix vs substring)
	githubTokenIdx := -1
	personalGithubIdx := -1
	for i, s := range sorted {
		if s == "github-token" {
			githubTokenIdx = i
		}
		if s == "personal-github" {
			personalGithubIdx = i
		}
	}
	if githubTokenIdx > personalGithubIdx {
		t.Errorf("'github-token' should come before 'personal-github' in sorted results")
	}
}

func TestSortByScoreEmptyQuery(t *testing.T) {
	secrets := []string{"c", "a", "b"}
	sorted := SortByScore(secrets, "")

	// With empty query, should maintain original order or alphabetical
	if len(sorted) != 3 {
		t.Errorf("SortByScore with empty query should return all items, got %d", len(sorted))
	}
}

func TestSortByScoreNoMatches(t *testing.T) {
	secrets := []string{"github", "gitlab"}
	sorted := SortByScore(secrets, "xyz")

	if len(sorted) != 0 {
		t.Errorf("SortByScore with no matches should return empty, got %v", sorted)
	}
}
