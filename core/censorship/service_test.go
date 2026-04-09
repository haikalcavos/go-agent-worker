package censorship

import (
	"testing"
)

func TestApplyRules_CaseInsensitiveReplacement(t *testing.T) {
	service, err := New([]string{"secret"}, "[redacted]", false)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	result := service.ApplyRules("Secret code")
	expected := "[redacted] code"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestApplyRules_InvalidRegexFallback(t *testing.T) {
	// Invalid regex pattern should be escaped and treated as literal
	service, err := New([]string{"(["}, "[redacted]", false)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	result := service.ApplyRules("([")
	expected := "[redacted]"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestApplyRules_MultiplePatterns(t *testing.T) {
	service, err := New([]string{"secret", "password"}, "[redacted]", false)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	result := service.ApplyRules("secret password")
	expected := "[redacted] [redacted]"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestApplyRules_WholeWordMatching(t *testing.T) {
	// matchWholeWords=true should not replace "secret" within "secretive"
	service, err := New([]string{"secret"}, "[redacted]", true)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	result := service.ApplyRules("This is secretive and secret")
	// "secretive" should remain, "secret" should be redacted
	if result != "This is secretive and [redacted]" {
		t.Errorf("got %q, want %q", result, "This is secretive and [redacted]")
	}
}

func TestApplyRules_EmptyInput(t *testing.T) {
	service, err := New([]string{"test"}, "[redacted]", false)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	result := service.ApplyRules("")
	expected := ""
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestApplyRules_EmptyPattern(t *testing.T) {
	// Empty patterns should be ignored
	service, err := New([]string{""}, "[redacted]", false)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	// No rules created, so input should be unchanged
	result := service.ApplyRules("test")
	expected := "test"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestApplyRules_RegexPattern(t *testing.T) {
	// Regex pattern for API key format
	service, err := New([]string{"sk-[a-zA-Z0-9]+"}, "[REDACTED_KEY]", false)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	result := service.ApplyRules("Your API key is sk-abc123def456")
	expected := "Your API key is [REDACTED_KEY]"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestProcessChunk(t *testing.T) {
	service, err := New([]string{"api.key"}, "***", false)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	input := "my api.key is secret"
	result := service.ProcessChunk(input)
	expected := "my *** is secret"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestTailLength(t *testing.T) {
	patterns := []string{"short", "veryverylongpattern"}
	service, err := New(patterns, "[redacted]", false)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	// TailLength should be the length of the longest pattern
	expected := len("veryverylongpattern")
	if service.TailLength != expected {
		t.Errorf("got TailLength=%d, want %d", service.TailLength, expected)
	}
}

func TestTailLength_WithEmptyPatterns(t *testing.T) {
	patterns := []string{"", "short", "", "medium"}
	service, err := New(patterns, "[redacted]", false)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	// TailLength should be length of "medium" (longest non-empty)
	expected := len("medium")
	if service.TailLength != expected {
		t.Errorf("got TailLength=%d, want %d", service.TailLength, expected)
	}
}

func BenchmarkApplyRules(b *testing.B) {
	service, _ := New(
		[]string{"password", "secret", "api", "token", "key"},
		"[redacted]",
		false,
	)
	text := "this is a secret password with api key and token"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ApplyRules(text)
	}
}

func TestApplyRules_CaseSensitiveIntegration(t *testing.T) {
	// Verify case-insensitive matching with various cases
	service, err := New([]string{"SECRET"}, "[redacted]", false)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	tests := []struct {
		input    string
		expected string
	}{
		{"SECRET", "[redacted]"},
		{"secret", "[redacted]"},
		{"Secret", "[redacted]"},
		{"sEcReT", "[redacted]"},
	}
	for _, tt := range tests {
		result := service.ApplyRules(tt.input)
		if result != tt.expected {
			t.Errorf("for input %q: got %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestApplyRules_MultipleOccurrences(t *testing.T) {
	service, err := New([]string{"test"}, "X", false)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	result := service.ApplyRules("test test test")
	expected := "X X X"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestApplyRules_SequentialRuleApplication(t *testing.T) {
	// Verify that rules are applied sequentially
	// Pattern 1: "a" -> "b", Pattern 2: "b" -> "c"
	// So "a" should become "b" then "c"
	service, err := New([]string{"a", "b"}, "X", false)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	result := service.ApplyRules("a")
	// "a" matches first rule -> "X"
	expected := "X"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestApplyRules_SpecialCharacters(t *testing.T) {
	service, err := New([]string{"@example.com"}, "[EMAIL]", false)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	result := service.ApplyRules("user@example.com")
	expected := "user[EMAIL]"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}
