package domain_test

import (
	"testing"

	"github.com/djdees/port-simulation/internal/domain"
	"github.com/djdees/port-simulation/internal/server/data"
)

// TestEmbeddedDatasetValid ensures the embedded port.json passes all validation rules.
func TestEmbeddedDatasetValid(t *testing.T) {
	t.Helper()
	d, err := data.Load()
	if err != nil {
		t.Fatalf("embedded dataset failed to load/validate: %v", err)
	}
	if len(d.Vessels) == 0 {
		t.Error("expected at least one vessel")
	}
	if len(d.Barges) == 0 {
		t.Error("expected at least one barge")
	}
	if len(d.Tows) == 0 {
		t.Error("expected at least one tow")
	}
	if len(d.Cargo) == 0 {
		t.Error("expected at least one cargo record")
	}
	if len(d.Facilities) == 0 {
		t.Error("expected at least one facility")
	}
	if len(d.Terminals) == 0 {
		t.Error("expected at least one terminal")
	}
	if len(d.Berths) == 0 {
		t.Error("expected at least one berth")
	}
	if len(d.Events) == 0 {
		t.Error("expected at least one event")
	}
}

// TestValidateCatchesDuplicateID ensures validation rejects duplicate IDs.
func TestValidateCatchesDuplicateID(t *testing.T) {
	d := &domain.PortData{
		Vessels: []domain.Vessel{
			{ID: "VSL-001", Name: "A", IMO: "IMO1234567", MMSI: "123456789",
				CallSign: "AA", FlagCountry: "US", Kind: domain.VesselKindFreighter,
				Owner: "Owner A", LengthM: 100, DraftM: 5},
			{ID: "VSL-001", Name: "B", IMO: "IMO7654321", MMSI: "987654321",
				CallSign: "BB", FlagCountry: "US", Kind: domain.VesselKindFreighter,
				Owner: "Owner B", LengthM: 100, DraftM: 5},
		},
	}
	err := domain.Validate(d)
	if err == nil {
		t.Fatal("expected validation error for duplicate vessel ID, got nil")
	}
}

// TestValidateCatchesBadCrossRef ensures validation rejects unknown cross-references.
func TestValidateCatchesBadCrossRef(t *testing.T) {
	v := domain.Vessel{
		ID: "VSL-001", Name: "Ship", IMO: "IMO1234567", MMSI: "123456789",
		CallSign: "AA", FlagCountry: "US", Kind: domain.VesselKindFreighter,
		Owner: "Owner", LengthM: 100, DraftM: 5,
	}
	tow := domain.Tow{
		ID:        "TOW-001",
		TowboatID: "VSL-999", // does not exist
		BargeIDs:  []string{},
		Name:      "Bad Tow",
	}
	d := &domain.PortData{Vessels: []domain.Vessel{v}, Tows: []domain.Tow{tow}}
	err := domain.Validate(d)
	if err == nil {
		t.Fatal("expected validation error for bad towboat_id cross-ref, got nil")
	}
}

// TestValidateCatchesMissingCarrier ensures cargo must reference a vessel or barge.
func TestValidateCatchesMissingCarrier(t *testing.T) {
	c := domain.Cargo{
		ID:          "CGO-001",
		Commodity:   domain.CommodityGrain,
		Origin:      "A",
		Destination: "B",
		WeightMT:    100,
	}
	d := &domain.PortData{Cargo: []domain.Cargo{c}}
	err := domain.Validate(d)
	if err == nil {
		t.Fatal("expected validation error for cargo with no carrier, got nil")
	}
}

// TestEnumValidation exercises enum IsValid methods.
func TestEnumValidation(t *testing.T) {
	if domain.VesselKind("unknown").IsValid() {
		t.Error("'unknown' should not be a valid VesselKind")
	}
	if !domain.VesselKindFreighter.IsValid() {
		t.Error("VesselKindFreighter should be valid")
	}
	if domain.FacilityKind("xyz").IsValid() {
		t.Error("'xyz' should not be a valid FacilityKind")
	}
	if !domain.FacilityKindCommercial.IsValid() {
		t.Error("FacilityKindCommercial should be valid")
	}
}
