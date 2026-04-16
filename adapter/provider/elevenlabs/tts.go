package elevenlabs

import (
	"os"

	"go-agent-worker/library/config"

	elevenlabsAdapter "github.com/cavos-io/conversation-worker/adapter/elevenlabs"
	"github.com/cavos-io/conversation-worker/core/tts"
)

func NewTTS(cfg config.TTSConfig) (tts.TTS, error) {
	apiKey := os.Getenv("ELEVENLABS_API_KEY")
	return elevenlabsAdapter.NewElevenLabsTTS(apiKey, cfg.VoiceName, cfg.Model)
}
