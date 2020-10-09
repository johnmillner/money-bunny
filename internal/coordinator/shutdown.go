package coordinator

import "github.com/google/uuid"

type Shutdown struct {
	From     uuid.UUID
	To       uuid.UUID
	Shutdown bool
}

func (s Shutdown) GetTo() uuid.UUID {
	return s.To
}
func (s Shutdown) GetFrom() uuid.UUID {
	return s.From
}
