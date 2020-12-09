package internal

import (
	"sync"
)

// ouroboros contains a ring buffer style snapshots that can have a rasterized array of the current state printed out
type ouroboros struct {
	snapshots []Snapshot
	capacity  int
	mutex     sync.RWMutex
}

func NewOuroboros(snapshots []Snapshot) *ouroboros {
	return &ouroboros{
		snapshots: snapshots,
		capacity:  cap(snapshots),
		mutex:     sync.RWMutex{},
	}
}

func (o *ouroboros) Insert(s Snapshot) {
	defer o.trim()
	defer o.mutex.Unlock()

	// check if this can simply be appended to the end
	if s.timestamp.After(o.snapshots[len(o.snapshots)-1].timestamp) {
		o.mutex.Lock()
		o.snapshots = append(o.snapshots, s)
		return
	}

	// insert stock somewhere in the middle looking backwards
	for i := len(o.snapshots) - 2; i >= 0; i-- {
		// if the timestamp is equal, replace it
		if s.timestamp.Equal(o.snapshots[i].timestamp) {
			o.mutex.Lock()
			o.snapshots[i] = s
			return
		}

		// if this timestamp is after the current - insert
		if s.timestamp.After(o.snapshots[i].timestamp) {
			o.mutex.Lock()
			o.snapshots = append(o.snapshots[:i+1], append([]Snapshot{s}, o.snapshots[i+1:]...)...)
			return
		}
	}

	o.mutex.Lock()
}

func (o *ouroboros) trim() {
	o.snapshots = o.snapshots[len(o.snapshots)-o.capacity:]
}

func (o *ouroboros) Get() []Snapshot {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	return o.snapshots
}
