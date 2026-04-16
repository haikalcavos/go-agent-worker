package openai

import (
	"os"

	"go-agent-worker/library/config"

	openaiAdapter "github.com/cavos-io/conversation-worker/adapter/openai"
	"github.com/cavos-io/conversation-worker/core/llm"
)

func NewLLM(cfg config.LLMConfig) (llm.LLM, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	return openaiAdapter.NewOpenAILLM(apiKey, cfg.Model), nil
}
