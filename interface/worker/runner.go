package worker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go-agent-worker/adapter/censoring"
	"go-agent-worker/adapter/provider"
	"go-agent-worker/core/callsession"
	"go-agent-worker/core/censorship"
	"go-agent-worker/library/config"

	"github.com/cavos-io/conversation-worker/core/agent"
	"github.com/cavos-io/conversation-worker/core/llm"
	"github.com/cavos-io/conversation-worker/interface/worker"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/webrtc/v4"
)

// Run is the entrypoint called by AgentServer for each incoming job.
// It wires the full STT→LLM→TTS pipeline and starts the conversation.
func Run(jobCtx *worker.JobContext) error {
	cfg := config.Get()
	log := slog.With("jobId", jobCtx.Job.Id, "room", jobCtx.Job.Room.Name)

	log.Info("orchestrator started")

	// --- Build providers from config ---
	sttProvider, err := provider.NewSTT(cfg.STT)
	if err != nil {
		log.Error("failed to initialize STT provider", "err", err)
		return err
	}

	llmProvider, err := provider.NewLLM(cfg.LLM)
	if err != nil {
		log.Error("failed to initialize LLM provider", "err", err)
		return err
	}

	ttsProvider, err := provider.NewTTS(cfg.TTS)
	if err != nil {
		log.Error("failed to initialize TTS provider", "err", err)
		return err
	}

	vadProvider, err := provider.NewVAD(cfg.VAD)
	if err != nil {
		log.Error("failed to initialize VAD provider", "err", err)
		return err
	}

	// --- Build censorship service (optional) ---
	var censorService *censorship.Service
	fmt.Println("censorship patterns:", cfg.Censorship.Patterns, len(cfg.Censorship.Patterns))
	if len(cfg.Censorship.Patterns) > 0 {
		cs, err := censorship.New(
			cfg.Censorship.Patterns,
			cfg.Censorship.Replacement,
			cfg.Censorship.MatchWholeWords,
		)
		if err != nil {
			log.Error("failed to create censorship service", "err", err)
		} else {
			censorService = cs
			log.Info("censorship service enabled", "patterns", len(cfg.Censorship.Patterns))
		}
	}

	// --- Wrap TTS with censorship filter (if enabled) ---
	if censorService != nil {
		ttsProvider = censoring.NewTTSWrapper(ttsProvider, censorService)
		log.Info("TTS censorship wrapper applied")
	}

	// --- Build chat context with system prompt ---
	instructions := renderInstructions(cfg)
	chatCtx := llm.NewChatContext()
	chatCtx.Append(&llm.ChatMessage{
		Role:    llm.ChatRoleSystem,
		Content: []llm.ChatContent{{Text: instructions}},
	})

	// --- Build agent ---
	a := agent.NewAgent(instructions)
	a.STT = sttProvider
	a.LLM = llmProvider
	a.TTS = ttsProvider
	a.VAD = vadProvider
	a.ChatCtx = chatCtx
	a.TurnDetection = agent.TurnDetectionModeVAD
	a.AllowInterruptions = true
	a.MinEndpointingDelay = 0.5
	a.MaxEndpointingDelay = 3.0

	// --- Connect to LiveKit room ---
	// Buffer early track subscriptions that arrive before RoomIO is ready.
	// Without this, audio tracks from the user are lost during the Connect→RoomIO
	// window, causing the agent to never hear the user.
	type earlyTrack struct {
		track *webrtc.TrackRemote
		pub   *lksdk.RemoteTrackPublication
		rp    *lksdk.RemoteParticipant
	}
	var earlyTracks []earlyTrack
	var roomIO *worker.RoomIO

	disconnectCh := make(chan struct{})
	cb := lksdk.NewRoomCallback()
	cb.OnDisconnected = func() { close(disconnectCh) }
	cb.OnTrackSubscribed = func(track *webrtc.TrackRemote, pub *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) {
		log.Info("track subscribed", "participant", string(rp.Identity()), "kind", track.Kind().String())
		if roomIO != nil {
			roomIO.GetCallback().OnTrackSubscribed(track, pub, rp)
		} else {
			log.Info("buffering early track")
			earlyTracks = append(earlyTracks, earlyTrack{track, pub, rp})
		}
	}

	if err := jobCtx.Connect(context.Background(), cb); err != nil {
		log.Error("failed to connect to room", "err", err)
		return err
	}
	log.Info("connected to room")

	// --- Create agent session ---
	session := agent.NewAgentSession(a, jobCtx.Room, agent.AgentSessionOptions{
		AllowInterruptions:  true,
		MinEndpointingDelay: 0.5,
		MaxEndpointingDelay: 3.0,
	})
	session.ChatCtx = chatCtx

	// --- Wire audio I/O and replay buffered tracks ---
	roomIO = worker.NewRoomIO(jobCtx.Room, session, worker.RoomOptions{})

	if len(earlyTracks) > 0 {
		log.Info("replaying buffered tracks", "count", len(earlyTracks))
		for _, et := range earlyTracks {
			roomIO.GetCallback().OnTrackSubscribed(et.track, et.pub, et.rp)
		}
	}

	if err := roomIO.Start(context.Background()); err != nil {
		log.Error("failed to start RoomIO", "err", err)
		return err
	}
	log.Info("room I/O started")

	// --- Domain: start call session tracking ---
	cs := callsession.New(jobCtx.Job.Id, jobCtx.Job.Room.Name)
	cs.Start()
	defer cs.End()

	// --- Start pipeline ---
	if err := session.Start(context.Background()); err != nil {
		log.Error("failed to start agent session", "err", err)
		return err
	}
	log.Info("agent session started")

	// --- Greet the caller ---
	// CRITICAL: Wait longer for full pipeline initialization
	// The audio pipeline needs time for:
	// 1. VAD to initialize
	// 2. Audio track to be published
	// 3. Queue processing goroutines to start
	// 4. TTS provider to be ready
	log.Info("waiting for audio pipeline to be fully ready...")
	time.Sleep(10 * time.Second) // Increased from 500ms to 1.5s

	greeting := cfg.Greeting
	log.Info("preparing greeting", "original", greeting)

	if censorService != nil && greeting != "" {
		greeting = censorService.ApplyRules(greeting)
		log.Info("greeting after censorship", "final", greeting)
	}

	if greeting != "" {
		log.Info("sending greeting to user", "text", greeting)

		// Try to send greeting with retry logic
		// since GenerateReply is async and may fail if timing is off
		maxRetries := 3
		var lastErr error

		for attempt := 1; attempt <= maxRetries; attempt++ {
			greetCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

			log.Info("attempting to send greeting", "attempt", attempt, "max", maxRetries)
			err := session.GenerateReply(greetCtx, greeting)

			cancel()

			if err != nil {
				lastErr = err
				log.Warn("greeting send attempt failed", "attempt", attempt, "err", err)

				// Wait before retry - gives async pipeline more time
				if attempt < maxRetries {
					time.Sleep(500 * time.Millisecond)
				}
			} else {
				log.Info("greeting scheduled successfully", "attempt", attempt)
				// GenerateReply is async, so wait a bit for it to process
				// This gives the pipeline time to actually send the audio
				time.Sleep(1000 * time.Millisecond)
				break
			}
		}

		if lastErr != nil && maxRetries > 0 {
			log.Error("failed to send greeting after retries", "err", lastErr, "text", greeting, "attempts", maxRetries)
		}
	} else {
		log.Warn("greeting is empty, skipping greeting")
	}

	// --- Block until room disconnects ---
	<-disconnectCh

	log.Info("session ended", "duration_sec", cs.DurationSec())
	return nil
}

// renderInstructions substitutes {{key}} placeholders in instructions with values from config context.
func renderInstructions(cfg *config.Config) string {
	instructions := cfg.Instructions
	for k, v := range cfg.Context {
		instructions = strings.ReplaceAll(instructions, "{{"+k+"}}", v)
	}
	return instructions
}
