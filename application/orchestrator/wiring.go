package orchestrator

import (
	"fmt"
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
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		openaiSTT := openaiAdapter.NewOpenAISTT(apiKey, cfg.Model)
		internalVAD := coreVAD.NewSimpleVAD(0.0005)
		return stt.NewStreamAdapter(openaiSTT, internalVAD), nil
	case "deepgram":
		apiKey := os.Getenv("DEEPGRAM_API_KEY")
		return deepgramAdapter.NewDeepgramSTT(apiKey, cfg.Model), nil
	case "google":
		credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		return googleAdapter.NewGoogleSTT(credentialsFile)
	default:
		return nil, fmt.Errorf("unsupported STT provider: %s", cfg.Provider)
	}
}

func newLLM(cfg config.LLMConfig) (llm.LLM, error) {
	switch cfg.Provider {
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		return openaiAdapter.NewOpenAILLM(apiKey, cfg.Model), nil
	case "google":
		apiKey := os.Getenv("GOOGLE_API_KEY")
		return googleAdapter.NewGoogleLLM(apiKey, cfg.Model)
	case "anthropic":
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		return anthropicAdapter.NewAnthropicLLM(apiKey, cfg.Model)
	case "groq":
		apiKey := os.Getenv("GROQ_API_KEY")
		return groqAdapter.NewGroqLLM(apiKey, cfg.Model), nil
	case "azure":
		apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
		endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
		return openaiAdapter.NewOpenAILLMWithBaseURL(apiKey, cfg.Model, endpoint), nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.Provider)
	}
}

func newTTS(cfg config.TTSConfig) (tts.TTS, error) {
	switch cfg.Provider {
	case "elevenlabs":
		apiKey := os.Getenv("ELEVENLABS_API_KEY")
		return elevenlabsAdapter.NewElevenLabsTTS(apiKey, cfg.VoiceName, cfg.Model)
	case "openai":
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
	case "google":
		credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		return googleAdapter.NewGoogleTTS(credentialsFile)
	case "azure":
		speechKey := os.Getenv("AZURE_SPEECH_KEY")
		speechRegion := os.Getenv("AZURE_SPEECH_REGION")
		return azureAdapter.NewAzureTTS(speechKey, speechRegion, cfg.VoiceName), nil
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
