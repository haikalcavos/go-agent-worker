package anthropic

import (
	"os"

	"go-agent-worker/library/config"

	"github.com/cavos-io/conversation-worker/core/llm"
	anthropicAdapter "github.com/cavos-io/conversation-worker/adapter/anthropic"
)

func NewLLM(cfg config.LLMConfig) (llm.LLM, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	return anthropicAdapter.NewAnthropicLLM(apiKey, cfg.Model)
}
