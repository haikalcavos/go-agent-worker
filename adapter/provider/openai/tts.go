package openai

import (
	"os"

	"go-agent-worker/library/config"

	"github.com/cavos-io/conversation-worker/core/tts"
	openaiAdapter "github.com/cavos-io/conversation-worker/adapter/openai"
	goopenai "github.com/sashabaranov/go-openai"
)

func NewTTS(cfg config.TTSConfig) (tts.TTS, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	model := goopenai.TTSModel1
	if cfg.Model != "" {
		model = goopenai.SpeechModel(cfg.Model)
	}
	voice := goopenai.VoiceAlloy
	if cfg.VoiceName != "" {
		voice = goopenai.SpeechVoice(cfg.VoiceName)
	}
	return openaiAdapter.NewOpenAITTS(apiKey, model, voice), nil
}
