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
func Run(server *worker.AgentServer, jobCtx *worker.JobContext) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	// --- Create agent session ---
	session := agent.NewAgentSession(a, nil, agent.AgentSessionOptions{
		AllowInterruptions:  true,
		MinEndpointingDelay: 0.5,
		MaxEndpointingDelay: 3.0,
	})
	session.ChatCtx = chatCtx

	// --- Register session with server (for console UI) ---
	server.SetConsoleSession(session)

	// --- Connect to LiveKit room ---
	// Callbacks delegate through a closure that reads the eventual *RoomIO
	// once it is assigned after Connect returns.
	var rio *worker.RoomIO
	cb := lksdk.NewRoomCallback()
	cb.OnDisconnected = func() { cancel() }
	cb.OnTrackSubscribed = func(track *webrtc.TrackRemote, pub *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) {
		log.Info("track subscribed", "participant", string(rp.Identity()), "kind", track.Kind().String())
		if rio == nil {
			return
		}
		rio.GetCallback().OnTrackSubscribed(track, pub, rp)
	}
	cb.OnTrackUnsubscribed = func(track *webrtc.TrackRemote, pub *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) {
		if rio == nil {
			return
		}
		rio.GetCallback().OnTrackUnsubscribed(track, pub, rp)
	}
	cb.OnParticipantDisconnected = func(rp *lksdk.RemoteParticipant) {
		if rio == nil {
			return
		}
		rio.GetCallback().OnParticipantDisconnected(rp)
	}

	if err := jobCtx.Connect(ctx, cb); err != nil {
		log.Error("failed to connect to room", "err", err)
		return err
	}
	log.Info("connected to room")

	// --- Wire audio I/O ---
	rio = worker.NewRoomIO(jobCtx.Room, session, worker.RoomOptions{})
	defer rio.Close()

	if err := rio.Start(ctx); err != nil {
		log.Error("failed to start RoomIO", "err", err)
		return err
	}
	log.Info("room I/O started")

	// --- Domain: start call session tracking ---
	cs := callsession.New(jobCtx.Job.Id, jobCtx.Job.Room.Name)
	cs.Start()
	defer cs.End()

	// --- Start pipeline ---
	if err := session.Start(ctx); err != nil {
		log.Error("failed to start agent session", "err", err)
		return err
	}
	log.Info("agent session started")

	// --- Greet the caller ---
	// Wait for the full audio pipeline to be ready before sending greeting.
	log.Info("waiting for audio pipeline to be fully ready...")
	time.Sleep(10 * time.Second)

	greeting := cfg.Greeting
	if censorService != nil && greeting != "" {
		greeting = censorService.ApplyRules(greeting)
	}

	// if greeting != "" {
	// 	log.Info("sending greeting", "text", greeting)

	// 	maxRetries := 3
	// 	var lastErr error
	// 	for attempt := 1; attempt <= maxRetries; attempt++ {
	// 		greetCtx, greetCancel := context.WithTimeout(ctx, 30*time.Second)
	// 		_, lastErr = session.GenerateReply(greetCtx, greeting, false)
	// 		greetCancel()

	// 		if lastErr == nil {
	// 			log.Info("greeting scheduled", "attempt", attempt)
	// 			time.Sleep(1000 * time.Millisecond)
	// 			break
	// 		}
	// 		log.Warn("greeting attempt failed", "attempt", attempt, "err", lastErr)
	// 		if attempt < maxRetries {
	// 			time.Sleep(500 * time.Millisecond)
	// 		}
	// 	}
	// 	if lastErr != nil {
	// 		log.Error("failed to send greeting after retries", "err", lastErr, "attempts", maxRetries)
	// 	}
	// } else {
	// 	log.Warn("greeting is empty, skipping")
	// }

	// --- Block until room disconnects (cancel() is called by OnDisconnected) ---
	<-ctx.Done()

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
