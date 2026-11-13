package handler

import "github.com/djdees/port-simulation/internal/domain"

// Store is an in-memory read-only view of the port dataset.
// All fields are populated at startup from the embedded port.json.
type Store struct {
	data *domain.PortData
}

// NewStore creates a Store backed by validated port data.
func NewStore(d *domain.PortData) *Store {
	return &Store{data: d}
}
