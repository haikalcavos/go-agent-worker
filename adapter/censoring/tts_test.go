package censoring

import (
	"io"
	"testing"

	"go-agent-worker/core/censorship"
	"github.com/cavos-io/conversation-worker/core/tts"
)

// mockTTSStream implements tts.SynthesizeStream for testing
type mockTTSStream struct {
	pushedTexts []string
	flushed     bool
	closed      bool
}

func (m *mockTTSStream) PushText(text string) error {
	m.pushedTexts = append(m.pushedTexts, text)
	return nil
}

func (m *mockTTSStream) Flush() error {
	m.flushed = true
	return nil
}

func (m *mockTTSStream) Close() error {
	m.closed = true
	return nil
}

func (m *mockTTSStream) Next() (*tts.SynthesizedAudio, error) {
	return nil, io.EOF
}

// TestSentenceBoundaryBuffering verifies that censoring waits for sentence boundaries
// before applying rules, so multi-word patterns can be matched correctly
func TestSentenceBoundaryBuffering(t *testing.T) {
	censorService, err := censorship.New(
		[]string{"Kentang Goreng"},
		"[redacted]",
		false,
	)
	if err != nil {
		t.Fatalf("failed to create censorship service: %v", err)
	}

	mockStream := &mockTTSStream{}
	stream := &censoringStream{
		inner:  mockStream,
		censor: censorService,
	}

	// Simulate LLM streaming token by token: "Kent" | "ang" | " goreng" | "!"
	tokens := []string{"Kent", "ang", " goreng", "!"}
	for _, token := range tokens {
		err := stream.PushText(token)
		if err != nil {
			t.Fatalf("PushText failed: %v", err)
		}
	}

	// When "!" is pushed, it triggers sentence boundary, so text should be flushed
	if len(mockStream.pushedTexts) != 1 {
		t.Errorf("expected 1 pushed text after sentence boundary, got %d: %v", len(mockStream.pushedTexts), mockStream.pushedTexts)
	}

	expected := "[redacted]!"
	if mockStream.pushedTexts[0] != expected {
		t.Errorf("expected censored text %q, got %q", expected, mockStream.pushedTexts[0])
	}

	// Explicit flush should be a no-op since buffer is empty
	err = stream.Flush()
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// Still only 1 pushed text
	if len(mockStream.pushedTexts) != 1 {
		t.Errorf("expected 1 pushed text after explicit flush, got %d", len(mockStream.pushedTexts))
	}

	if !mockStream.flushed {
		t.Error("expected inner stream to be flushed")
	}
}

// TestSentenceBoundaryWithMultipleSentences verifies correct behavior with multiple sentences
func TestSentenceBoundaryWithMultipleSentences(t *testing.T) {
	censorService, err := censorship.New(
		[]string{"badword"},
		"[X]",
		false,
	)
	if err != nil {
		t.Fatalf("failed to create censorship service: %v", err)
	}

	mockStream := &mockTTSStream{}
	stream := &censoringStream{
		inner:  mockStream,
		censor: censorService,
	}

	// Sentence 1: "Hello badword." (ends with .)
	stream.PushText("Hello ")
	stream.PushText("badword")
	stream.PushText(". ")

	// Should have flushed first sentence (period is boundary, space is kept in buffer)
	if len(mockStream.pushedTexts) != 1 {
		t.Errorf("after first sentence, expected 1 text, got %d: %v", len(mockStream.pushedTexts), mockStream.pushedTexts)
	}
	if mockStream.pushedTexts[0] != "Hello [X]." {
		t.Errorf("expected %q, got %q", "Hello [X].", mockStream.pushedTexts[0])
	}

	// Sentence 2: "This is badword again!" (ends with !)
	// The space from prev boundary + new text becomes: " This is badword again!"
	stream.PushText("This is ")
	stream.PushText("badword")
	stream.PushText(" again!")

	// Should have flushed second sentence
	if len(mockStream.pushedTexts) != 2 {
		t.Errorf("after second sentence, expected 2 texts, got %d: %v", len(mockStream.pushedTexts), mockStream.pushedTexts)
	}
	// Note: the leading space from previous sentence boundary is preserved
	if mockStream.pushedTexts[1] != " This is [X] again!" {
		t.Errorf("expected %q, got %q", " This is [X] again!", mockStream.pushedTexts[1])
	}

	// Remaining text: "Still buffering..." (ends with . so it has a sentence boundary!)
	stream.PushText("Still buffering...")

	// The "..." at the end is recognized as a sentence boundary, so it auto-flushes
	if len(mockStream.pushedTexts) != 3 {
		t.Errorf("expected 3 texts after 'Still buffering...' (dot boundary), got %d", len(mockStream.pushedTexts))
	}
	if mockStream.pushedTexts[2] != "Still buffering..." {
		t.Errorf("expected %q, got %q", "Still buffering...", mockStream.pushedTexts[2])
	}

	// Final flush should be a no-op
	stream.Flush()

	if len(mockStream.pushedTexts) != 3 {
		t.Errorf("after final flush, expected still 3 texts, got %d", len(mockStream.pushedTexts))
	}
}

// TestLastSentenceBoundary tests the helper function
func TestLastSentenceBoundary(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		desc     string
	}{
		{"Hello.", 6, "period at end (index after .)"},
		{"Hello! World.", 13, "multiple sentences, finds last ."},
		{"Hello! ", 6, "exclamation followed by space"},
		{"What?", 5, "question mark at end"},
		{"No boundary", -1, "no sentence boundary"},
		{"Hello. ", 6, "period followed by space"},
		{"Test\n", 5, "newline boundary"},
		{"", -1, "empty string"},
		{"3.14 is pi", -1, "period not at word boundary (period not followed by space)"},
		{"Dr. Smith said hello.", 21, "abbreviation then real sentence"},
	}

	for _, tt := range tests {
		result := lastSentenceBoundary(tt.input)
		if result != tt.expected {
			t.Errorf("%s: lastSentenceBoundary(%q) = %d, expected %d", tt.desc, tt.input, result, tt.expected)
		}
	}
}
