package internal

import (
	"github.com/magiconair/properties/assert"
	"testing"
	"time"
)

func TestOuroboros_Insert_Filling(t *testing.T) {
	o := NewOuroboros(make([]Snapshot, 3))

	t1 := Snapshot{
		timestamp: time.Time{}.Add(time.Minute),
	}
	t2 := Snapshot{
		timestamp: time.Time{}.Add(2 * time.Minute),
	}
	t3 := Snapshot{
		timestamp: time.Time{}.Add(3 * time.Minute),
	}

	o.Insert(t1)
	assert.Equal(t, o.Get(), []Snapshot{{}, {}, t1})
	o.Insert(t2)
	assert.Equal(t, o.Get(), []Snapshot{{}, t1, t2})
	o.Insert(t3)
	assert.Equal(t, o.Get(), []Snapshot{t1, t2, t3})
}

func TestOuroboros_Insert_Overflow(t *testing.T) {
	t1 := Snapshot{
		timestamp: time.Time{}.Add(time.Minute),
	}
	t2 := Snapshot{
		timestamp: time.Time{}.Add(2 * time.Minute),
	}
	t3 := Snapshot{
		timestamp: time.Time{}.Add(3 * time.Minute),
	}
	t4 := Snapshot{
		timestamp: time.Time{}.Add(4 * time.Minute),
	}
	t5 := Snapshot{
		timestamp: time.Time{}.Add(5 * time.Minute),
	}
	t6 := Snapshot{
		timestamp: time.Time{}.Add(6 * time.Minute),
	}

	o := NewOuroboros([]Snapshot{t1, t2, t3})

	o.Insert(t4)
	assert.Equal(t, o.Get(), []Snapshot{t2, t3, t4})
	o.Insert(t5)
	assert.Equal(t, o.Get(), []Snapshot{t3, t4, t5})
	o.Insert(t6)
	assert.Equal(t, o.Get(), []Snapshot{t4, t5, t6})
}

func TestOuroboros_Insert_Insertion(t *testing.T) {
	t1 := Snapshot{
		timestamp: time.Time{}.Add(time.Minute),
	}
	t2 := Snapshot{
		timestamp: time.Time{}.Add(2 * time.Minute),
	}
	t3 := Snapshot{
		timestamp: time.Time{}.Add(3 * time.Minute),
	}
	t4 := Snapshot{
		timestamp: time.Time{}.Add(4 * time.Minute),
	}

	o := NewOuroboros([]Snapshot{t1, t2, t4})

	o.Insert(t3)
	assert.Equal(t, o.Get(), []Snapshot{t2, t3, t4})
}

func TestOuroboros_Insert_Replace(t *testing.T) {
	t1 := Snapshot{
		timestamp: time.Time{}.Add(time.Minute),
	}
	t2 := Snapshot{
		Price:     1,
		timestamp: time.Time{}.Add(2 * time.Minute),
	}
	t22 := Snapshot{
		Price:     2,
		timestamp: time.Time{}.Add(2 * time.Minute),
	}
	t3 := Snapshot{
		timestamp: time.Time{}.Add(3 * time.Minute),
	}

	o := NewOuroboros([]Snapshot{t1, t2, t3})

	o.Insert(t22)
	assert.Equal(t, o.Get(), []Snapshot{t1, t22, t3})
}

func TestOuroboros_Insert_Early(t *testing.T) {
	t1 := Snapshot{
		timestamp: time.Time{}.Add(time.Minute),
	}
	t2 := Snapshot{
		timestamp: time.Time{}.Add(2 * time.Minute),
	}
	t3 := Snapshot{
		timestamp: time.Time{}.Add(3 * time.Minute),
	}
	t4 := Snapshot{
		timestamp: time.Time{}.Add(4 * time.Minute),
	}

	o := NewOuroboros([]Snapshot{t2, t3, t4})

	o.Insert(t1)
	assert.Equal(t, o.Get(), []Snapshot{t2, t3, t4})
}
