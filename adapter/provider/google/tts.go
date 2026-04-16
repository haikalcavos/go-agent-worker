package google

import (
	"os"

	"go-agent-worker/library/config"

	googleAdapter "github.com/cavos-io/conversation-worker/adapter/google"
	"github.com/cavos-io/conversation-worker/core/tts"
)

func NewTTS(cfg config.TTSConfig) (tts.TTS, error) {
	credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	return googleAdapter.NewGoogleTTS(credentialsFile)
}
