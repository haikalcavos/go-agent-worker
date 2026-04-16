package deepgram

import (
	"os"

	"go-agent-worker/library/config"

	deepgramAdapter "github.com/cavos-io/conversation-worker/adapter/deepgram"
	"github.com/cavos-io/conversation-worker/core/stt"
)

func NewSTT(cfg config.STTConfig) (stt.STT, error) {
	apiKey := os.Getenv("DEEPGRAM_API_KEY")
	return deepgramAdapter.NewDeepgramSTT(apiKey, cfg.Model), nil
}
