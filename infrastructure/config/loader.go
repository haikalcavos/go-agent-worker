package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

var (
	once     sync.Once
	instance *Config
)

// Load reads configuration from a JSON file then overlays env vars.
// Call once at startup. Get() returns the cached result.
func Load(configPath string) (*Config, error) {
	var err error
	once.Do(func() {
		cfg := defaultConfig()

		if configPath != "" {
			data, readErr := os.ReadFile(configPath)
			if readErr != nil {
				err = fmt.Errorf("reading config file: %w", readErr)
				return
			}
			if jsonErr := json.Unmarshal(data, cfg); jsonErr != nil {
				err = fmt.Errorf("parsing config file: %w", jsonErr)
				return
			}
		}

		overlayEnv(cfg)
		instance = cfg
	})
	return instance, err
}

// Get returns the loaded config. Panics if Load was not called first.
func Get() *Config {
	if instance == nil {
		panic("config.Load must be called before config.Get")
	}
	return instance
}

func defaultConfig() *Config {
	return &Config{
		AgentName: "cavos-agent",
		Language:  "id-ID",
		VAD: VADConfig{
			Provider:            "simple",
			MinSpeechDuration:   0.05,
			MinSilenceDuration:  0.3,
			ActivationThreshold: 0.5,
			SampleRate:          16000,
		},
		STT: STTConfig{Provider: "deepgram", Model: "nova-2"},
		LLM: LLMConfig{Provider: "openai", Model: "gpt-4o-mini"},
		TTS: TTSConfig{Provider: "elevenlabs", SampleRate: 24000},
		TurnDetection: TurnDetectionConfig{Use: "vad"},
		CallPolicy: CallPolicyConfig{
			MaxDurationSec: 300,
			Idle: IdlePolicy{
				ReminderEnabled:     true,
				ReminderAfterSec:    20,
				TerminationAfterSec: 45,
				ReminderMessage:     "Are you still there?",
			},
		},
	}
}

// overlayEnv overrides string fields from environment variables.
func overlayEnv(cfg *Config) {
	if v := os.Getenv("AGENT_NAME"); v != "" {
		cfg.AgentName = v
	}
	if v := os.Getenv("LIVEKIT_URL"); v != "" {
		cfg.LiveKitURL = v
	}
	if v := os.Getenv("LIVEKIT_API_KEY"); v != "" {
		cfg.LiveKitAPIKey = v
	}
	if v := os.Getenv("LIVEKIT_API_SECRET"); v != "" {
		cfg.LiveKitAPISecret = v
	}
	// Provider API keys are read directly by the wiring layer via os.Getenv
}
