package openai

import (
	"os"

	"go-agent-worker/library/config"

	openaiAdapter "github.com/cavos-io/conversation-worker/adapter/openai"
	"github.com/cavos-io/conversation-worker/core/stt"
	coreVAD "github.com/cavos-io/conversation-worker/core/vad"
)

func NewSTT(cfg config.STTConfig) (stt.STT, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	openaiSTT := openaiAdapter.NewOpenAISTT(apiKey, "")
	internalVAD := coreVAD.NewSimpleVAD(0.0005)
	stt := stt.NewStreamAdapter(openaiSTT, internalVAD)
	return stt, nil
}
