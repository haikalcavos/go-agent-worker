package anthropic

import (
	"os"

	"go-agent-worker/library/config"

	anthropicAdapter "github.com/cavos-io/conversation-worker/adapter/anthropic"
	"github.com/cavos-io/conversation-worker/core/llm"
)

func NewLLM(cfg config.LLMConfig) (llm.LLM, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	return anthropicAdapter.NewAnthropicLLM(apiKey, cfg.Model)
}
