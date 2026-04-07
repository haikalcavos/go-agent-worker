package orchestrator

import (
	"context"
	"log"
	"log/slog"
	"strings"

	"go-agent-worker/domain/callsession"
	"go-agent-worker/infrastructure/config"

	"github.com/cavos-io/conversation-worker/core/agent"
	"github.com/cavos-io/conversation-worker/interface/worker"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

// Run is the entrypoint called by AgentServer for each incoming job.
// It wires the full STT→LLM→TTS pipeline and starts the conversation.
func Run(ctx *worker.JobContext) error {
	log.Println("Orchestrator Run called with job:", ctx.Job.Id)
	cfg := config.Get()
	log := slog.With("jobId", ctx.Job.Id, "room", ctx.Job.Room.Name)

	// --- Build providers from config ---
	sttProvider, err := newSTT(cfg.STT)
	if err != nil {
		log.Error("failed to initialize STT provider", "err", err)
		return err
	}

	llmProvider, err := newLLM(cfg.LLM)
	if err != nil {
		log.Error("failed to initialize LLM provider", "err", err)
		return err
	}

	ttsProvider, err := newTTS(cfg.TTS)
	if err != nil {
		log.Error("failed to initialize TTS provider", "err", err)
		return err
	}

	vadProvider, err := newVAD(cfg.VAD)
	if err != nil {
		log.Error("failed to initialize VAD provider", "err", err)
		return err
	}

	// --- Build agent ---
	a := agent.NewAgent(renderInstructions(cfg))
	a.STT = sttProvider
	a.LLM = llmProvider
	a.TTS = ttsProvider
	a.VAD = vadProvider
	a.AllowInterruptions = true

	// --- Connect to LiveKit room ---
	// disconnectCh is closed when the room disconnects.
	disconnectCh := make(chan struct{})
	cb := lksdk.NewRoomCallback()
	cb.OnDisconnected = func() { close(disconnectCh) }

	if err := ctx.Connect(context.Background(), cb); err != nil {
		log.Error("failed to connect to room", "err", err)
		return err
	}
	log.Info("connected to room")

	// --- Create agent session ---
	session := agent.NewAgentSession(a, ctx.Room, agent.AgentSessionOptions{
		AllowInterruptions:  true,
		MinEndpointingDelay: 0.3,
		MaxEndpointingDelay: 1.5,
	})

	log.Info("After new agent session")

	// --- Wire audio I/O ---
	// NewRoomIO wires VAD→STT→LLM→TTS and hooks PublishAudio to the room track.
	rio := worker.NewRoomIO(ctx.Room, session, worker.RoomOptions{})

	log.Info("After new RoomIO")
	rioCb := rio.GetCallback()
	log.Info("After get RoomIO callback")
	cb.OnTrackSubscribed = rioCb.OnTrackSubscribed

	if err := rio.Start(context.Background()); err != nil {
		log.Error("failed to start RoomIO", "err", err)
		return err
	}

	log.Info("RoomIO started, waiting for audio...")
	// --- Domain: start call session tracking ---
	cs := callsession.New(ctx.Job.Id, ctx.Job.Room.Name)
	cs.Start()
	defer cs.End()

	// --- Start pipeline ---
	if err := session.Start(context.Background()); err != nil {
		log.Error("failed to start agent session", "err", err)
		return err
	}
	log.Info("session started")

	// --- Greet the caller ---
	if cfg.Greeting != "" {
		if err := session.GenerateReply(context.Background(), cfg.Greeting); err != nil {
			log.Warn("failed to send greeting", "err", err)
		}

		log.Info("Sent greeting to caller")
	}

	log.Info("After greeting")
	// --- Block until room disconnects ---
	<-disconnectCh

	log.Info("session ended", "duration_sec", cs.DurationSec())
	return nil
}

// renderInstructions substitutes {{key}} variables from cfg.Context into the instructions.
func renderInstructions(cfg *config.Config) string {
	instructions := cfg.Instructions
	log.Println("Original instructions:", instructions)
	for k, v := range cfg.Context {
		instructions = strings.ReplaceAll(instructions, "{{"+k+"}}", v)
	}
	return instructions
}
