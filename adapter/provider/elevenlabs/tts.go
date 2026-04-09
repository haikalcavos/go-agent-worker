package elevenlabs

import (
	"os"

	"go-agent-worker/library/config"

	"github.com/cavos-io/conversation-worker/core/tts"
	elevenlabsAdapter "github.com/cavos-io/conversation-worker/adapter/elevenlabs"
)

func NewTTS(cfg config.TTSConfig) (tts.TTS, error) {
	apiKey := os.Getenv("ELEVENLABS_API_KEY")
	return elevenlabsAdapter.NewElevenLabsTTS(apiKey, cfg.VoiceName, cfg.Model)
}
