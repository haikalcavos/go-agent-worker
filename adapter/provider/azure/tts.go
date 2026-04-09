package azure

import (
	"os"

	"go-agent-worker/library/config"

	"github.com/cavos-io/conversation-worker/core/tts"
	azureAdapter "github.com/cavos-io/conversation-worker/adapter/azure"
)

func NewTTS(cfg config.TTSConfig) (tts.TTS, error) {
	speechKey := os.Getenv("AZURE_SPEECH_KEY")
	speechRegion := os.Getenv("AZURE_SPEECH_REGION")
	return azureAdapter.NewAzureTTS(speechKey, speechRegion, cfg.VoiceName), nil
}
