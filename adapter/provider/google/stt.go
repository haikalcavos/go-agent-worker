package google

import (
	"os"

	"go-agent-worker/library/config"

	googleAdapter "github.com/cavos-io/conversation-worker/adapter/google"
	"github.com/cavos-io/conversation-worker/core/stt"
)

func NewSTT(cfg config.STTConfig) (stt.STT, error) {
	credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	return googleAdapter.NewGoogleSTT(credentialsFile)
}
