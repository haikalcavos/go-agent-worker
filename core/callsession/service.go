package callsession

import (
	"time"
)

// Status represents the lifecycle state of a call session.
type Status string

const (
	StatusPending Status = "pending"
	StatusActive  Status = "active"
	StatusEnded   Status = "ended"
)

// CallSession is the root aggregate for a single voice call.
// It tracks lifecycle, duration, and accumulated events.
type CallSession struct {
	ID       string
	RoomName string
	Status   Status

	startedAt time.Time
	endedAt   time.Time

	events []Event
}

// New creates a new CallSession in pending state.
func New(id, roomName string) *CallSession {
	return &CallSession{
		ID:       id,
		RoomName: roomName,
		Status:   StatusPending,
	}
}

// Start transitions the session to active.
func (cs *CallSession) Start() {
	cs.Status = StatusActive
	cs.startedAt = time.Now()
	cs.raise(SessionStarted{SessionID: cs.ID, RoomName: cs.RoomName, At: cs.startedAt})
}

// End transitions the session to ended.
func (cs *CallSession) End() {
	cs.Status = StatusEnded
	cs.endedAt = time.Now()
	cs.raise(SessionEnded{SessionID: cs.ID, DurationSec: cs.DurationSec(), At: cs.endedAt})
}

// DurationSec returns elapsed seconds since Start().
func (cs *CallSession) DurationSec() float64 {
	if cs.startedAt.IsZero() {
		return 0
	}
	end := cs.endedAt
	if end.IsZero() {
		end = time.Now()
	}
	return end.Sub(cs.startedAt).Seconds()
}

// Events returns uncommitted domain events and clears the internal buffer.
func (cs *CallSession) Events() []Event {
	evts := cs.events
	cs.events = nil
	return evts
}

func (cs *CallSession) raise(e Event) {
	cs.events = append(cs.events, e)
}
