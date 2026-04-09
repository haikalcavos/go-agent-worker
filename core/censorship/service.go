package censorship

import (
	"regexp"
	"unicode"
)

// Rule holds a compiled regex pattern and its replacement string.
type Rule struct {
	pattern     *regexp.Regexp
	replacement string
}

// Service applies a list of censorship rules to text.
type Service struct {
	rules      []Rule
	TailLength int // Length of the longest raw pattern (used by stream buffers)
}

// New creates a Service from a list of string patterns.
//
// Parameters:
//   - patterns: list of raw strings or regex patterns to censor
//   - replacement: string to substitute matched content (default "[redacted]")
//   - matchWholeWords: if true, wraps word-only patterns with \b boundaries
//
// Patterns are compiled case-insensitively ((?i) prefix).
// Invalid regex patterns are escaped as literals (fallback to QuoteMeta).
func New(patterns []string, replacement string, matchWholeWords bool) (*Service, error) {
	rules := make([]Rule, 0, len(patterns))
	maxLen := 0

	for _, raw := range patterns {
		if raw == "" {
			continue
		}
		if len(raw) > maxLen {
			maxLen = len(raw)
		}

		src := raw
		if matchWholeWords && isWordOnly(raw) {
			src = `\b` + regexp.QuoteMeta(raw) + `\b`
		}

		compiled, err := regexp.Compile("(?i)" + src)
		if err != nil {
			// Fallback: escape as literal
			compiled = regexp.MustCompile("(?i)" + regexp.QuoteMeta(raw))
		}

		rules = append(rules, Rule{pattern: compiled, replacement: replacement})
	}

	return &Service{rules: rules, TailLength: maxLen}, nil
}

// ApplyRules applies all censorship rules sequentially to the input text.
func (s *Service) ApplyRules(text string) string {
	result := text
	for _, rule := range s.rules {
		result = rule.pattern.ReplaceAllString(result, rule.replacement)
	}
	return result
}

// ProcessChunk is an alias for ApplyRules, for stream-processing compatibility.
func (s *Service) ProcessChunk(chunk string) string {
	return s.ApplyRules(chunk)
}

// isWordOnly returns true if s contains only Unicode letters, digits, or underscores.
func isWordOnly(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}
	return true
}
