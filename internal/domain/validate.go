package domain

import (
	"fmt"
	"strings"
)

// ValidationError is returned when dataset integrity checks fail.
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("dataset validation failed with %d error(s):\n  %s",
		len(e.Errors), strings.Join(e.Errors, "\n  "))
}

// Validate checks the full PortData dataset for integrity.
// It accumulates all errors rather than stopping at the first.
func Validate(d *PortData) error {
	var errs []string
	add := func(format string, args ...any) {
		errs = append(errs, fmt.Sprintf(format, args...))
	}

	// Build ID sets for cross-reference checks.
	vesselIDs := make(map[string]bool)
	bargeIDs := make(map[string]bool)
	towIDs := make(map[string]bool)
	cargoIDs := make(map[string]bool)
	facilityIDs := make(map[string]bool)
	terminalIDs := make(map[string]bool)
	berthIDs := make(map[string]bool)

	// --- Vessels ---
	for i, v := range d.Vessels {
		if v.ID == "" {
			add("vessels[%d]: missing id", i)
		} else if vesselIDs[v.ID] {
			add("vessels[%d]: duplicate id %q", i, v.ID)
		} else {
			vesselIDs[v.ID] = true
		}
		if v.Name == "" {
			add("vessel %q: missing name", v.ID)
		}
		if v.IMO == "" {
			add("vessel %q: missing imo", v.ID)
		}
		if v.FlagCountry == "" {
			add("vessel %q: missing flag_country", v.ID)
		}
		if !v.Kind.IsValid() {
			add("vessel %q: invalid kind %q", v.ID, v.Kind)
		}
		if v.LengthM <= 0 {
			add("vessel %q: length_m must be > 0", v.ID)
		}
		if v.DraftM <= 0 {
			add("vessel %q: draft_m must be > 0", v.ID)
		}
	}

	// --- Barges ---
	for i, b := range d.Barges {
		if b.ID == "" {
			add("barges[%d]: missing id", i)
		} else if bargeIDs[b.ID] {
			add("barges[%d]: duplicate id %q", i, b.ID)
		} else {
			bargeIDs[b.ID] = true
		}
		if b.Name == "" {
			add("barge %q: missing name", b.ID)
		}
		if !b.Capability.IsValid() {
			add("barge %q: invalid capability %q", b.ID, b.Capability)
		}
		if b.CapacityMT < 0 {
			add("barge %q: capacity_mt must be >= 0", b.ID)
		}
		if b.LengthM <= 0 {
			add("barge %q: length_m must be > 0", b.ID)
		}
	}

	// --- Tows ---
	for i, t := range d.Tows {
		if t.ID == "" {
			add("tows[%d]: missing id", i)
		} else if towIDs[t.ID] {
			add("tows[%d]: duplicate id %q", i, t.ID)
		} else {
			towIDs[t.ID] = true
		}
		if t.TowboatID == "" {
			add("tow %q: missing towboat_id", t.ID)
		}
		if len(t.BargeIDs) == 0 {
			add("tow %q: must have at least one barge_id", t.ID)
		}
	}
	// Deferred cross-ref: tow towboat and barge IDs (need vessel/barge sets fully built first)
	for _, t := range d.Tows {
		if t.TowboatID != "" && !vesselIDs[t.TowboatID] {
			add("tow %q: towboat_id %q does not reference a known vessel", t.ID, t.TowboatID)
		}
		for _, bid := range t.BargeIDs {
			if !bargeIDs[bid] {
				add("tow %q: barge_id %q does not reference a known barge", t.ID, bid)
			}
		}
	}

	// --- Cargo ---
	for i, c := range d.Cargo {
		if c.ID == "" {
			add("cargo[%d]: missing id", i)
		} else if cargoIDs[c.ID] {
			add("cargo[%d]: duplicate id %q", i, c.ID)
		} else {
			cargoIDs[c.ID] = true
		}
		if !c.Commodity.IsValid() {
			add("cargo %q: invalid commodity %q", c.ID, c.Commodity)
		}
		if c.Origin == "" {
			add("cargo %q: missing origin", c.ID)
		}
		if c.Destination == "" {
			add("cargo %q: missing destination", c.ID)
		}
		if c.VesselID == "" && c.BargeID == "" {
			add("cargo %q: must have vessel_id or barge_id", c.ID)
		}
		if c.VesselID != "" && c.BargeID != "" {
			add("cargo %q: cannot have both vessel_id and barge_id", c.ID)
		}
		if c.WeightMT <= 0 {
			add("cargo %q: weight_mt must be > 0", c.ID)
		}
	}
	// Deferred cargo carrier cross-ref
	for _, c := range d.Cargo {
		if c.VesselID != "" && !vesselIDs[c.VesselID] {
			add("cargo %q: vessel_id %q does not reference a known vessel", c.ID, c.VesselID)
		}
		if c.BargeID != "" && !bargeIDs[c.BargeID] {
			add("cargo %q: barge_id %q does not reference a known barge", c.ID, c.BargeID)
		}
	}

	// --- Facilities ---
	for i, f := range d.Facilities {
		if f.ID == "" {
			add("facilities[%d]: missing id", i)
		} else if facilityIDs[f.ID] {
			add("facilities[%d]: duplicate id %q", i, f.ID)
		} else {
			facilityIDs[f.ID] = true
		}
		if f.Name == "" {
			add("facility %q: missing name", f.ID)
		}
		if !f.Kind.IsValid() {
			add("facility %q: invalid kind %q", f.ID, f.Kind)
		}
	}

	// --- Terminals ---
	for i, t := range d.Terminals {
		if t.ID == "" {
			add("terminals[%d]: missing id", i)
		} else if terminalIDs[t.ID] {
			add("terminals[%d]: duplicate id %q", i, t.ID)
		} else {
			terminalIDs[t.ID] = true
		}
		if t.Name == "" {
			add("terminal %q: missing name", t.ID)
		}
		if t.FacilityID == "" {
			add("terminal %q: missing facility_id", t.ID)
		} else if !facilityIDs[t.FacilityID] {
			add("terminal %q: facility_id %q does not reference a known facility", t.ID, t.FacilityID)
		}
	}

	// --- Berths ---
	for i, b := range d.Berths {
		if b.ID == "" {
			add("berths[%d]: missing id", i)
		} else if berthIDs[b.ID] {
			add("berths[%d]: duplicate id %q", i, b.ID)
		} else {
			berthIDs[b.ID] = true
		}
		if b.Name == "" {
			add("berth %q: missing name", b.ID)
		}
		if b.TerminalID == "" {
			add("berth %q: missing terminal_id", b.ID)
		} else if !terminalIDs[b.TerminalID] {
			add("berth %q: terminal_id %q does not reference a known terminal", b.ID, b.TerminalID)
		}
		if b.MaxDraftM <= 0 {
			add("berth %q: max_draft_m must be > 0", b.ID)
		}
		if b.MaxLengthM <= 0 {
			add("berth %q: max_length_m must be > 0", b.ID)
		}
	}

	// --- Events ---
	eventIDs := make(map[string]bool)
	for i, e := range d.Events {
		if e.ID == "" {
			add("events[%d]: missing id", i)
		} else if eventIDs[e.ID] {
			add("events[%d]: duplicate id %q", i, e.ID)
		} else {
			eventIDs[e.ID] = true
		}
		if !e.EventType.IsValid() {
			add("event %q: invalid event_type %q", e.ID, e.EventType)
		}
		// At least one subject
		if e.VesselID == "" && e.BargeID == "" && e.TowID == "" {
			add("event %q: must have at least one of vessel_id, barge_id, tow_id", e.ID)
		}
		if e.OccurredAt.IsZero() {
			add("event %q: missing occurred_at", e.ID)
		}
		// Cross-refs
		if e.VesselID != "" && !vesselIDs[e.VesselID] {
			add("event %q: vessel_id %q does not reference a known vessel", e.ID, e.VesselID)
		}
		if e.BargeID != "" && !bargeIDs[e.BargeID] {
			add("event %q: barge_id %q does not reference a known barge", e.ID, e.BargeID)
		}
		if e.TowID != "" && !towIDs[e.TowID] {
			add("event %q: tow_id %q does not reference a known tow", e.ID, e.TowID)
		}
		if e.FacilityID != "" && !facilityIDs[e.FacilityID] {
			add("event %q: facility_id %q does not reference a known facility", e.ID, e.FacilityID)
		}
		if e.BerthID != "" && !berthIDs[e.BerthID] {
			add("event %q: berth_id %q does not reference a known berth", e.ID, e.BerthID)
		}
		for _, cid := range e.CargoIDs {
			if !cargoIDs[cid] {
				add("event %q: cargo_id %q does not reference a known cargo record", e.ID, cid)
			}
		}
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}
