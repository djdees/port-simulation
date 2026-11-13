// Package domain defines the fictional port data model for the New Orleans Port
// Information Service. All entities, identifiers, and relationships are simulated
// and do not represent real vessels, facilities, or operational data.
package domain

import "time"

// PortData is the top-level container for the embedded dataset.
type PortData struct {
	Vessels    []Vessel       `json:"vessels"`
	Barges     []Barge        `json:"barges"`
	Tows       []Tow          `json:"tows"`
	Cargo      []Cargo        `json:"cargo"`
	Facilities []Facility     `json:"facilities"`
	Terminals  []Terminal     `json:"terminals"`
	Berths     []Berth        `json:"berths"`
	Events     []TrafficEvent `json:"events"`
}

// --- Enumerations ---

// VesselKind classifies commercial vessel types in the dataset.
type VesselKind string

const (
	VesselKindFreighter VesselKind = "freighter"
	VesselKindTanker    VesselKind = "tanker"
	VesselKindOSV       VesselKind = "osv" // offshore service vessel
	VesselKindTowboat   VesselKind = "towboat"
	VesselKindBulker    VesselKind = "bulker"
)

var validVesselKinds = map[VesselKind]bool{
	VesselKindFreighter: true,
	VesselKindTanker:    true,
	VesselKindOSV:       true,
	VesselKindTowboat:   true,
	VesselKindBulker:    true,
}

// IsValid reports whether v is a recognized vessel kind.
func (v VesselKind) IsValid() bool { return validVesselKinds[v] }

// BargeCapability describes what a barge is equipped to carry.
type BargeCapability string

const (
	BargeCapabilityDryBulk    BargeCapability = "dry_bulk"
	BargeCapabilityLiquidBulk BargeCapability = "liquid_bulk"
	BargeCapabilityHopper     BargeCapability = "hopper"
	BargeCapabilityCover      BargeCapability = "covered"
	BargeCapabilityOpen       BargeCapability = "open"
)

var validBargeCapabilities = map[BargeCapability]bool{
	BargeCapabilityDryBulk:    true,
	BargeCapabilityLiquidBulk: true,
	BargeCapabilityHopper:     true,
	BargeCapabilityCover:      true,
	BargeCapabilityOpen:       true,
}

// IsValid reports whether b is a recognized barge capability.
func (b BargeCapability) IsValid() bool { return validBargeCapabilities[b] }

// FacilityKind classifies port facilities.
type FacilityKind string

const (
	FacilityKindCommercial FacilityKind = "commercial"
	FacilityKindIndustrial FacilityKind = "industrial"
	FacilityKindService    FacilityKind = "service"
	FacilityKindAnchorage  FacilityKind = "anchorage"
)

var validFacilityKinds = map[FacilityKind]bool{
	FacilityKindCommercial: true,
	FacilityKindIndustrial: true,
	FacilityKindService:    true,
	FacilityKindAnchorage:  true,
}

// IsValid reports whether f is a recognized facility kind.
func (f FacilityKind) IsValid() bool { return validFacilityKinds[f] }

// EventType classifies port traffic events.
type EventType string

const (
	EventTypeArrival        EventType = "arrival"
	EventTypeDeparture      EventType = "departure"
	EventTypeBerthAssign    EventType = "berth_assignment"
	EventTypeLoadStart      EventType = "load_start"
	EventTypeLoadComplete   EventType = "load_complete"
	EventTypeUnloadStart    EventType = "unload_start"
	EventTypeUnloadComplete EventType = "unload_complete"
	EventTypeService        EventType = "service"
	EventTypeDelay          EventType = "delay"
	EventTypeAnchor         EventType = "anchor"
)

var validEventTypes = map[EventType]bool{
	EventTypeArrival:        true,
	EventTypeDeparture:      true,
	EventTypeBerthAssign:    true,
	EventTypeLoadStart:      true,
	EventTypeLoadComplete:   true,
	EventTypeUnloadStart:    true,
	EventTypeUnloadComplete: true,
	EventTypeService:        true,
	EventTypeDelay:          true,
	EventTypeAnchor:         true,
}

// IsValid reports whether e is a recognized event type.
func (e EventType) IsValid() bool { return validEventTypes[e] }

// CommodityType classifies cargo commodities.
type CommodityType string

const (
	CommodityGrain      CommodityType = "grain"
	CommoditySoybeans   CommodityType = "soybeans"
	CommoditySteel      CommodityType = "steel"
	CommodityPetroleum  CommodityType = "petroleum"
	CommodityLNG        CommodityType = "lng"
	CommodityChemicals  CommodityType = "chemicals"
	CommodityContainers CommodityType = "containers"
	CommodityCoal       CommodityType = "coal"
	CommodityFertilizer CommodityType = "fertilizer"
	CommodityTimber     CommodityType = "timber"
)

var validCommodities = map[CommodityType]bool{
	CommodityGrain:      true,
	CommoditySoybeans:   true,
	CommoditySteel:      true,
	CommodityPetroleum:  true,
	CommodityLNG:        true,
	CommodityChemicals:  true,
	CommodityContainers: true,
	CommodityCoal:       true,
	CommodityFertilizer: true,
	CommodityTimber:     true,
}

// IsValid reports whether c is a recognized commodity type.
func (c CommodityType) IsValid() bool { return validCommodities[c] }

// --- Core Entity Types ---

// Vessel represents a commercial or service vessel. All records are fictional.
type Vessel struct {
	ID          string     `json:"id"` // e.g. "VSL-001"
	Name        string     `json:"name"`
	IMO         string     `json:"imo"`  // fictional IMO number
	MMSI        string     `json:"mmsi"` // fictional MMSI
	CallSign    string     `json:"call_sign"`
	FlagCountry string     `json:"flag_country"` // ISO 3166-1 alpha-2
	Kind        VesselKind `json:"kind"`
	Owner       string     `json:"owner"`
	LengthM     float64    `json:"length_m"`
	DraftM      float64    `json:"draft_m"`
}

// Barge represents a river barge. All records are fictional.
type Barge struct {
	ID         string          `json:"id"` // e.g. "BGE-001"
	Name       string          `json:"name"`
	Owner      string          `json:"owner"`
	CapacityMT float64         `json:"capacity_mt"` // metric tonnes
	Capability BargeCapability `json:"capability"`
	LengthM    float64         `json:"length_m"`
}

// Tow represents a towboat pushing a fleet of barges. All records are fictional.
type Tow struct {
	ID        string   `json:"id"`         // e.g. "TOW-001"
	TowboatID string   `json:"towboat_id"` // references a Vessel with Kind=towboat
	BargeIDs  []string `json:"barge_ids"`
	Name      string   `json:"name"`
}

// Cargo represents a cargo record associated with a vessel or barge. All records are fictional.
type Cargo struct {
	ID          string        `json:"id"` // e.g. "CGO-001"
	Commodity   CommodityType `json:"commodity"`
	Origin      string        `json:"origin"` // port/city name
	Destination string        `json:"destination"`
	VesselID    string        `json:"vessel_id,omitempty"`
	BargeID     string        `json:"barge_id,omitempty"`
	WeightMT    float64       `json:"weight_mt"`
	Notes       string        `json:"notes,omitempty"`
}

// Facility represents a port facility. All records are fictional.
type Facility struct {
	ID   string       `json:"id"` // e.g. "FAC-001"
	Name string       `json:"name"`
	Kind FacilityKind `json:"kind"`
}

// Terminal represents a terminal within a port facility. All records are fictional.
type Terminal struct {
	ID         string `json:"id"` // e.g. "TRM-001"
	Name       string `json:"name"`
	FacilityID string `json:"facility_id"`
}

// Berth represents a berth within a terminal. All records are fictional.
type Berth struct {
	ID         string  `json:"id"` // e.g. "BTH-001"
	Name       string  `json:"name"`
	TerminalID string  `json:"terminal_id"`
	MaxDraftM  float64 `json:"max_draft_m"`
	MaxLengthM float64 `json:"max_length_m"`
}

// TrafficEvent represents a port traffic event. All records are fictional.
type TrafficEvent struct {
	ID         string    `json:"id"` // e.g. "EVT-001"
	EventType  EventType `json:"event_type"`
	VesselID   string    `json:"vessel_id,omitempty"`
	BargeID    string    `json:"barge_id,omitempty"`
	TowID      string    `json:"tow_id,omitempty"`
	FacilityID string    `json:"facility_id,omitempty"`
	BerthID    string    `json:"berth_id,omitempty"`
	CargoIDs   []string  `json:"cargo_ids,omitempty"`
	OccurredAt time.Time `json:"occurred_at"` // UTC RFC3339
	Notes      string    `json:"notes,omitempty"`
}
