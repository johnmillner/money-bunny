package gatherers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInsert_Multiple(t *testing.T) {
	original := []Equity{{
		Name: "1",
	}, {
		Name: "4",
	}}

	added := []Equity{{
		Name: "2",
	}, {
		Name: "3",
	}}

	expected := []Equity{{
		Name: "1",
	}, {
		Name: "2",
	}, {
		Name: "3",
	}, {
		Name: "4",
	}}

	assert.Equal(t, expected, Insert(original, 1, added...))
}

func TestInsert_Single(t *testing.T) {
	original := []Equity{{
		Name: "1",
	}, {
		Name: "3",
	}}

	added := Equity{
		Name: "2",
	}

	expected := []Equity{{
		Name: "1",
	}, {
		Name: "2",
	}, {
		Name: "3",
	}}

	assert.Equal(t, expected, Insert(original, 1, added))
}

func TestInsert_Nothing(t *testing.T) {
	original := []Equity{{
		Name: "1",
	}, {
		Name: "2",
	}}

	expected := []Equity{{
		Name: "1",
	}, {
		Name: "2",
	}}

	assert.Equal(t, expected, Insert(original, 1))
}

func TestInsert_Low(t *testing.T) {
	original := []Equity{{
		Name: "2",
	}, {
		Name: "3",
	}}

	added := Equity{
		Name: "1",
	}

	expected := []Equity{{
		Name: "1",
	}, {
		Name: "2",
	}, {
		Name: "3",
	}}

	assert.Equal(t, expected, Insert(original, -1, added))
}

func TestInsert_High(t *testing.T) {
	original := []Equity{{
		Name: "1",
	}, {
		Name: "2",
	}}

	added := Equity{
		Name: "3",
	}

	expected := []Equity{{
		Name: "1",
	}, {
		Name: "2",
	}, {
		Name: "3",
	}}

	assert.Equal(t, expected, Insert(original, 100, added))
}
