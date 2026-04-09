package google

import (
	"os"

	"go-agent-worker/library/config"

	"github.com/cavos-io/conversation-worker/core/stt"
	googleAdapter "github.com/cavos-io/conversation-worker/adapter/google"
)

func NewSTT(cfg config.STTConfig) (stt.STT, error) {
	credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	return googleAdapter.NewGoogleSTT(credentialsFile)
}
