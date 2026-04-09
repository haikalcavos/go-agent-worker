package google

import (
	"os"

	"go-agent-worker/library/config"

	"github.com/cavos-io/conversation-worker/core/tts"
	googleAdapter "github.com/cavos-io/conversation-worker/adapter/google"
)

func NewTTS(cfg config.TTSConfig) (tts.TTS, error) {
	credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	return googleAdapter.NewGoogleTTS(credentialsFile)
}
