package callsession

import "time"

// Event is the marker interface for all call session domain events.
type Event interface {
	eventType() string
}

// SessionStarted fires when the agent connects and the session begins.
type SessionStarted struct {
	SessionID string
	RoomName  string
	At        time.Time
}

func (SessionStarted) eventType() string { return "session.started" }

// SessionEnded fires when the room disconnects.
type SessionEnded struct {
	SessionID   string
	DurationSec float64
	At          time.Time
}

func (SessionEnded) eventType() string { return "session.ended" }

// TranscriptReceived fires when the STT produces a final transcript.
type TranscriptReceived struct {
	SessionID string
	Text      string
	Speaker   string // "user" | "agent"
	At        time.Time
}

func (TranscriptReceived) eventType() string { return "session.transcript_received" }
