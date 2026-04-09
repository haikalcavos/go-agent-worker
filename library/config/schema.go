package config

// Config is the top-level runtime configuration, sourced from env vars + configuration.json.
type Config struct {
	// LiveKit / worker identity
	AgentName        string `env:"AGENT_NAME"`
	LiveKitURL       string `env:"LIVEKIT_URL"`
	LiveKitAPIKey    string `env:"LIVEKIT_API_KEY"`
	LiveKitAPISecret string `env:"LIVEKIT_API_SECRET"`

	// Conversation
	Instructions string `json:"instruct"`
	Greeting     string `json:"greeting"`
	Language     string `json:"language"` // e.g. "id-ID"

	// Context variables rendered into Instructions
	Context map[string]string `json:"context"`

	// Provider configs
	VAD VADConfig `json:"vad"`
	STT STTConfig `json:"stt"`
	LLM LLMConfig `json:"llm"`
	TTS TTSConfig `json:"tts"`

	// Session behaviour
	TurnDetection     TurnDetectionConfig     `json:"turn_detection"`
	CallPolicy        CallPolicyConfig        `json:"call_policy"`
	Recording         RecordingConfig         `json:"recording"`
	NoiseCancellation NoiseCancellationConfig `json:"noise_cancellation"`
	Censorship        CensorshipConfig        `json:"censorship"`
}

type VADConfig struct {
	Provider            string  `json:"provider"`             // "silero" | "simple"
	MinSpeechDuration   float64 `json:"min_speech_duration"`
	MinSilenceDuration  float64 `json:"min_silence_duration"`
	ActivationThreshold float64 `json:"activation_threshold"`
	SampleRate          int     `json:"sample_rate"`
}

type STTConfig struct {
	Provider       string `json:"provider"` // "deepgram" | "google" | "assemblyai" | ...
	Model          string `json:"model"`
	Language       string `json:"language"`
	DetectLanguage bool   `json:"detect_language"`
}

type LLMConfig struct {
	Provider    string  `json:"provider"` // "openai" | "google" | "anthropic" | "groq" | ...
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
}

type TTSConfig struct {
	Provider   string `json:"provider"`   // "google" | "elevenlabs" | "azure" | ...
	VoiceName  string `json:"voice_name"`
	Model      string `json:"model"`
	SampleRate int    `json:"sample_rate"`
}

type TurnDetectionConfig struct {
	Use string `json:"use"` // "vad" | "stt" | "manual"
}

type CallPolicyConfig struct {
	MaxDurationSec int        `json:"max_duration_sec"`
	Idle           IdlePolicy `json:"idle"`
}

type IdlePolicy struct {
	ReminderEnabled     bool   `json:"reminder_enabled"`
	ReminderAfterSec    int    `json:"reminder_after_sec"`
	TerminationAfterSec int    `json:"termination_after_sec"`
	ReminderMessage     string `json:"reminder_message"`
}

type RecordingConfig struct {
	Enabled bool   `json:"enabled"`
	Bucket  string `json:"bucket"`
	Prefix  string `json:"prefix"`
}

type NoiseCancellationConfig struct {
	Provider string `json:"provider"`
}

type CensorshipConfig struct {
	Patterns        []string `json:"patterns"`
	Replacement     string   `json:"replacement"`
	MatchWholeWords bool     `json:"match_whole_words"`
}
