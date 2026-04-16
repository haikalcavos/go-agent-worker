package silero

import (
	"go-agent-worker/library/config"

	sileroVADAdapter "github.com/cavos-io/conversation-worker/adapter/silero_vad"
	coreVAD "github.com/cavos-io/conversation-worker/core/vad"
)

func NewVAD(cfg config.VADConfig) (coreVAD.VAD, error) {
	return sileroVADAdapter.NewSileroVAD(sileroVADAdapter.SileroVADOptions{
		MinSpeechDuration:   cfg.MinSpeechDuration,
		MinSilenceDuration:  cfg.MinSilenceDuration,
		ActivationThreshold: cfg.ActivationThreshold,
	}), nil
}
