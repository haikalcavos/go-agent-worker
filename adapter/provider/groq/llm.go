package groq

import (
	"os"

	"go-agent-worker/library/config"

	groqAdapter "github.com/cavos-io/conversation-worker/adapter/groq"
	"github.com/cavos-io/conversation-worker/core/llm"
)

func NewLLM(cfg config.LLMConfig) (llm.LLM, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	return groqAdapter.NewGroqLLM(apiKey, cfg.Model), nil
}
