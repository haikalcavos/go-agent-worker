package censoring

import (
	"context"
	"log"
	"strings"

	"go-agent-worker/core/censorship"

	"github.com/cavos-io/conversation-worker/core/tts"
)

// TTSWrapper wraps a tts.TTS provider and applies censorship rules
// to every text chunk before it is pushed to the underlying TTS stream.
type TTSWrapper struct {
	inner  tts.TTS
	censor *censorship.Service
}

// NewTTSWrapper creates a TTSWrapper. If censor is nil, returns inner unchanged.
func NewTTSWrapper(inner tts.TTS, censor *censorship.Service) tts.TTS {
	if censor == nil {
		return inner
	}
	return &TTSWrapper{inner: inner, censor: censor}
}

// --- Delegate all non-text methods to inner ---

func (w *TTSWrapper) Label() string {
	return w.inner.Label()
}

func (w *TTSWrapper) Capabilities() tts.TTSCapabilities {
	return w.inner.Capabilities()
}

func (w *TTSWrapper) SampleRate() int {
	return w.inner.SampleRate()
}

func (w *TTSWrapper) NumChannels() int {
	return w.inner.NumChannels()
}

// Synthesize applies censorship to the full text string before synthesis.
func (w *TTSWrapper) Synthesize(ctx context.Context, text string) (tts.ChunkedStream, error) {
	return w.inner.Synthesize(ctx, w.censor.ApplyRules(text))
}

// Stream returns a wrapped SynthesizeStream that censors each PushText call.
func (w *TTSWrapper) Stream(ctx context.Context) (tts.SynthesizeStream, error) {
	innerStream, err := w.inner.Stream(ctx)
	if err != nil {
		return nil, err
	}
	return &censoringStream{inner: innerStream, censor: w.censor}, nil
}

// censoringStream wraps SynthesizeStream and intercepts PushText.
// It buffers text until sentence boundaries to ensure multi-word patterns are matched correctly.
type censoringStream struct {
	inner  tts.SynthesizeStream
	censor *censorship.Service
	buf    strings.Builder
}

// PushText accumulates text until a sentence boundary is found, then applies censorship
// and forwards the complete sentence. This ensures multi-word patterns can be matched correctly.
func (s *censoringStream) PushText(text string) error {
	log.Println("Push Text=============================")
	s.buf.WriteString(text)

	// Look for sentence boundary (., !, ?, \n followed by space or end-of-string)
	raw := s.buf.String()
	boundary := lastSentenceBoundary(raw)
	if boundary < 0 {
		return nil // no complete sentence yet, keep buffering
	}

	toSend := raw[:boundary]
	s.buf.Reset()
	s.buf.WriteString(raw[boundary:])

	censored := s.censor.ApplyRules(toSend)
	if censored == "" {
		return nil
	}
	return s.inner.PushText(censored)
}

// Flush flushes any remaining buffered text (end of LLM response) after censoring.
func (s *censoringStream) Flush() error {
	if s.buf.Len() > 0 {
		censored := s.censor.ApplyRules(s.buf.String())
		s.buf.Reset()
		if censored != "" {
			if err := s.inner.PushText(censored); err != nil {
				return err
			}
		}
	}
	return s.inner.Flush()
}

func (s *censoringStream) Close() error {
	return s.inner.Close()
}

func (s *censoringStream) Next() (*tts.SynthesizedAudio, error) {
	return s.inner.Next()
}

// lastSentenceBoundary returns the index AFTER the last sentence-ending character
// (., !, ?, or \n) that is followed by a space, newline, or end-of-string.
// Returns -1 if no sentence boundary is found.
func lastSentenceBoundary(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		c := s[i]
		if c == '.' || c == '!' || c == '?' || c == '\n' {
			after := i + 1
			if after == len(s) || s[after] == ' ' || s[after] == '\n' {
				return after
			}
		}
	}
	return -1
}
