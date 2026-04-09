package provider

import (
	"fmt"

	"go-agent-worker/library/config"

	"github.com/cavos-io/conversation-worker/core/llm"
	"github.com/cavos-io/conversation-worker/core/stt"
	"github.com/cavos-io/conversation-worker/core/tts"
	coreVAD "github.com/cavos-io/conversation-worker/core/vad"

	anthropicProvider "go-agent-worker/adapter/provider/anthropic"
	azureProvider "go-agent-worker/adapter/provider/azure"
	deepgramProvider "go-agent-worker/adapter/provider/deepgram"
	elevenlabsProvider "go-agent-worker/adapter/provider/elevenlabs"
	googleProvider "go-agent-worker/adapter/provider/google"
	groqProvider "go-agent-worker/adapter/provider/groq"
	openaiProvider "go-agent-worker/adapter/provider/openai"
	sileroProvider "go-agent-worker/adapter/provider/silero"
)

func NewSTT(cfg config.STTConfig) (stt.STT, error) {
	switch cfg.Provider {
	case "openai":
		return openaiProvider.NewSTT(cfg)
	case "deepgram":
		return deepgramProvider.NewSTT(cfg)
	case "google":
		return googleProvider.NewSTT(cfg)
	default:
		return nil, fmt.Errorf("unsupported STT provider: %s", cfg.Provider)
	}
}

func NewLLM(cfg config.LLMConfig) (llm.LLM, error) {
	switch cfg.Provider {
	case "openai":
		return openaiProvider.NewLLM(cfg)
	case "google":
		return googleProvider.NewLLM(cfg)
	case "anthropic":
		return anthropicProvider.NewLLM(cfg)
	case "groq":
		return groqProvider.NewLLM(cfg)
	case "azure":
		return azureProvider.NewLLM(cfg)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.Provider)
	}
}

func NewTTS(cfg config.TTSConfig) (tts.TTS, error) {
	switch cfg.Provider {
	case "elevenlabs":
		return elevenlabsProvider.NewTTS(cfg)
	case "openai":
		return openaiProvider.NewTTS(cfg)
	case "google":
		return googleProvider.NewTTS(cfg)
	case "azure":
		return azureProvider.NewTTS(cfg)
	default:
		return nil, fmt.Errorf("unsupported TTS provider: %s", cfg.Provider)
	}
}

func NewVAD(cfg config.VADConfig) (coreVAD.VAD, error) {
	switch cfg.Provider {
	case "silero":
		return sileroProvider.NewVAD(cfg)
	case "simple", "":
		return coreVAD.NewSimpleVAD(cfg.ActivationThreshold), nil
	default:
		return nil, fmt.Errorf("unsupported VAD provider: %s", cfg.Provider)
	}
}
