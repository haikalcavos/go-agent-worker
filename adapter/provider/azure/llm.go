package azure

import (
	"os"

	"go-agent-worker/library/config"

	openaiAdapter "github.com/cavos-io/conversation-worker/adapter/openai"
	"github.com/cavos-io/conversation-worker/core/llm"
)

func NewLLM(cfg config.LLMConfig) (llm.LLM, error) {
	apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	return openaiAdapter.NewOpenAILLMWithBaseURL(apiKey, cfg.Model, endpoint), nil
}
