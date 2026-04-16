package google

import (
	"os"

	"go-agent-worker/library/config"

	googleAdapter "github.com/cavos-io/conversation-worker/adapter/google"
	"github.com/cavos-io/conversation-worker/core/llm"
)

func NewLLM(cfg config.LLMConfig) (llm.LLM, error) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	return googleAdapter.NewGoogleLLM(apiKey, cfg.Model)
}
