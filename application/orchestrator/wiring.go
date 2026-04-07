package orchestrator

import (
	"fmt"
	"log"
	"os"

	"go-agent-worker/infrastructure/config"

	"github.com/cavos-io/conversation-worker/core/llm"
	"github.com/cavos-io/conversation-worker/core/stt"
	"github.com/cavos-io/conversation-worker/core/tts"
	coreVAD "github.com/cavos-io/conversation-worker/core/vad"

	anthropicAdapter "github.com/cavos-io/conversation-worker/adapter/anthropic"
	azureAdapter "github.com/cavos-io/conversation-worker/adapter/azure"
	deepgramAdapter "github.com/cavos-io/conversation-worker/adapter/deepgram"
	elevenlabsAdapter "github.com/cavos-io/conversation-worker/adapter/elevenlabs"
	googleAdapter "github.com/cavos-io/conversation-worker/adapter/google"
	groqAdapter "github.com/cavos-io/conversation-worker/adapter/groq"
	openaiAdapter "github.com/cavos-io/conversation-worker/adapter/openai"
	sileroVADAdapter "github.com/cavos-io/conversation-worker/adapter/silero_vad"

	goopenai "github.com/sashabaranov/go-openai"
)

func newSTT(cfg config.STTConfig) (stt.STT, error) {
	switch cfg.Provider {
	case "deepgram":
		return deepgramAdapter.NewDeepgramSTT(os.Getenv("DEEPGRAM_API_KEY"), cfg.Model), nil
	case "google":
		return googleAdapter.NewGoogleSTT(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	default:
		return nil, fmt.Errorf("unsupported STT provider: %s", cfg.Provider)
	}
}

func newLLM(cfg config.LLMConfig) (llm.LLM, error) {
	log.Println("newLLM with provider:", cfg.Provider) // Debug log to check provider value
	switch cfg.Provider {
	case "openai":
		return openaiAdapter.NewOpenAILLM(os.Getenv("OPENAI_API_KEY"), cfg.Model), nil
	case "google":
		return googleAdapter.NewGoogleLLM(os.Getenv("GOOGLE_API_KEY"), cfg.Model)
	case "anthropic":
		return anthropicAdapter.NewAnthropicLLM(os.Getenv("ANTHROPIC_API_KEY"), cfg.Model)
	case "groq":
		return groqAdapter.NewGroqLLM(os.Getenv("GROQ_API_KEY"), cfg.Model), nil
	case "azure":
		// Azure uses OpenAI-compatible API with custom base URL
		return openaiAdapter.NewOpenAILLMWithBaseURL(
			os.Getenv("AZURE_OPENAI_API_KEY"),
			cfg.Model,
			os.Getenv("AZURE_OPENAI_ENDPOINT"),
		), nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.Provider)
	}
}

func newTTS(cfg config.TTSConfig) (tts.TTS, error) {
	switch cfg.Provider {
	case "elevenlabs":
		return elevenlabsAdapter.NewElevenLabsTTS(
			os.Getenv("ELEVENLABS_API_KEY"),
			cfg.VoiceName,
			cfg.Model,
		)
	case "openai":
		model := goopenai.TTSModel1
		if cfg.Model != "" {
			model = goopenai.SpeechModel(cfg.Model)
		}
		voice := goopenai.VoiceAlloy
		if cfg.VoiceName != "" {
			voice = goopenai.SpeechVoice(cfg.VoiceName)
		}
		return openaiAdapter.NewOpenAITTS(os.Getenv("OPENAI_API_KEY"), model, voice), nil
	case "google":
		return googleAdapter.NewGoogleTTS(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	case "azure":
		return azureAdapter.NewAzureTTS(
			os.Getenv("AZURE_SPEECH_KEY"),
			os.Getenv("AZURE_SPEECH_REGION"),
			cfg.VoiceName,
		), nil
	default:
		return nil, fmt.Errorf("unsupported TTS provider: %s", cfg.Provider)
	}
}

func newVAD(cfg config.VADConfig) (coreVAD.VAD, error) {
	switch cfg.Provider {
	case "silero":
		return sileroVADAdapter.NewSileroVAD(sileroVADAdapter.SileroVADOptions{
			MinSpeechDuration:   cfg.MinSpeechDuration,
			MinSilenceDuration:  cfg.MinSilenceDuration,
			ActivationThreshold: cfg.ActivationThreshold,
		}), nil
	case "simple", "":
		return coreVAD.NewSimpleVAD(cfg.ActivationThreshold), nil
	default:
		return nil, fmt.Errorf("unsupported VAD provider: %s", cfg.Provider)
	}
}
