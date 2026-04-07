package bootstrap

import (
	"log/slog"
	"os"

	"go-agent-worker/infrastructure/config"

	"github.com/joho/godotenv"
)

// Init sets up logging and loads configuration.
// Call once at the start of main().
func Init() {

	// Load .env file
	godotenv.Load()

	setupLogging()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configuration.json"
	}

	if _, err := config.Load(configPath); err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	slog.Info("agent-worker started", "agent", config.Get().AgentName)
}

// Config is a convenience accessor for the loaded config.
func Config() *config.Config {
	return config.Get()
}

func setupLogging() {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
}
