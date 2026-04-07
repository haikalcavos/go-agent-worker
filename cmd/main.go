package main

import (
	"go-agent-worker/infrastructure/bootstrap"
	"go-agent-worker/application/orchestrator"

	"github.com/cavos-io/conversation-worker/interface/cli"
	"github.com/cavos-io/conversation-worker/interface/worker"
)

func main() {
	// Initialize logging, OTEL, heartbeat
	bootstrap.Init()

	cfg := bootstrap.Config()

	server := worker.NewAgentServer(worker.WorkerOptions{
		AgentName: cfg.AgentName,
		WSRL:      cfg.LiveKitURL,
		APIKey:    cfg.LiveKitAPIKey,
		APISecret: cfg.LiveKitAPISecret,
	})

	server.RTCSession(orchestrator.Run, nil, nil)

	cli.RunApp(server)
}
