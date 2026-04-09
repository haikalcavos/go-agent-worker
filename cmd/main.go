package main

import (
	"go-agent-worker/library/config"
	"log"

	agentworker "go-agent-worker/interface/worker"

	"github.com/cavos-io/conversation-worker/interface/cli"
	"github.com/cavos-io/conversation-worker/interface/worker"
)

func main() {
	// Initialize logging, load .env and configuration.json
	InitApp()

	cfg := config.Get()

	server := worker.NewAgentServer(worker.WorkerOptions{
		AgentName: cfg.AgentName,
		WSRL:      cfg.LiveKitURL,
		APIKey:    cfg.LiveKitAPIKey,
		APISecret: cfg.LiveKitAPISecret,
	})

	log.Printf("%+v", worker.WorkerOptions{
		AgentName: cfg.AgentName,
		WSRL:      cfg.LiveKitURL,
		APIKey:    cfg.LiveKitAPIKey,
		APISecret: cfg.LiveKitAPISecret,
	})

	server.RTCSession(agentworker.Run, nil, nil)

	cli.RunApp(server)
}
