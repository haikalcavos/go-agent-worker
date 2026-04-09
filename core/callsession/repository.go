package callsession

// Repository is the port interface for call session persistence.
// Implementations live in the adapter layer.
type Repository interface {
	Save(session *CallSession) error
	FindByID(id string) (*CallSession, error)
}
