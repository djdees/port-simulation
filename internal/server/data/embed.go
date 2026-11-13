// Package data embeds the fictional port dataset and provides a validated loader.
// All data is simulated and does not represent real operational information.
package data

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/djdees/port-simulation/internal/domain"
)

//go:embed port.json
var portJSON []byte

// Load parses and validates the embedded port dataset.
// It returns an error if JSON is malformed or domain validation fails.
// Call once at server startup; the returned PortData is read-only.
func Load() (*domain.PortData, error) {
	var d domain.PortData
	if err := json.Unmarshal(portJSON, &d); err != nil {
		return nil, fmt.Errorf("parsing embedded port.json: %w", err)
	}
	if err := domain.Validate(&d); err != nil {
		return nil, fmt.Errorf("validating embedded port.json: %w", err)
	}
	return &d, nil
}
