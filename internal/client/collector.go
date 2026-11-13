// Package client - periodic collector writing API snapshots to CSV files.
package client

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// preset describes one collection query.
type preset struct {
	id   string
	path string
}

// All five presets in collection order.
var presets = []preset{
	{id: "arrivals", path: "/api/v1/arrivals"},
	{id: "departures", path: "/api/v1/departures"},
	{id: "vessels", path: "/api/v1/vessels"},
	{id: "cargo", path: "/api/v1/cargo"},
	{id: "berths", path: "/api/v1/berths"},
}

// Snapshot is the latest successful collection for a preset.
type Snapshot struct {
	QueryID     string    `json:"query_id"`
	CollectedAt time.Time `json:"collected_at"`
	Items       []any     `json:"items"`
}

// Collector runs the periodic preset collection loop and maintains snapshots.
type Collector struct {
	client    *Client
	outputDir string
	interval  time.Duration
	logger    *slog.Logger

	mu        sync.RWMutex
	snapshots map[string]*Snapshot
}

// NewCollector creates a Collector. Call Run to start collecting.
func NewCollector(c *Client, outputDir string, interval time.Duration, logger *slog.Logger) *Collector {
	return &Collector{
		client:    c,
		outputDir: outputDir,
		interval:  interval,
		logger:    logger,
		snapshots: make(map[string]*Snapshot),
	}
}

// GetSnapshot returns the latest snapshot for query, or nil if none yet.
func (col *Collector) GetSnapshot(query string) *Snapshot {
	col.mu.RLock()
	defer col.mu.RUnlock()
	return col.snapshots[query]
}

// Run starts the collection loop. It blocks until ctx is cancelled.
// It collects one preset per tick (interval), cycling through all five.
// Each preset repeats every interval*5 seconds.
func (col *Collector) Run(ctx context.Context) {
	if err := os.MkdirAll(col.outputDir, 0o750); err != nil {
		col.logger.Error("cannot create output directory", "dir", col.outputDir, "err", err)
		return
	}

	col.logger.Info("collector starting", "interval", col.interval, "output_dir", col.outputDir)

	// Collect the first preset immediately.
	idx := 0
	col.collect(ctx, presets[idx])
	idx = (idx + 1) % len(presets)

	ticker := time.NewTicker(col.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			col.logger.Info("collector stopped")
			return
		case <-ticker.C:
			col.collect(ctx, presets[idx])
			idx = (idx + 1) % len(presets)
		}
	}
}

func (col *Collector) collect(ctx context.Context, p preset) {
	now := time.Now().UTC()
	col.logger.Info("collecting preset", "preset", p.id)

	var resp struct {
		Items []any `json:"items"`
	}
	if err := col.client.Get(ctx, p.path, &resp); err != nil {
		col.logger.Error("collection failed", "preset", p.id, "err", err)
		return
	}

	// Update in-memory snapshot.
	snap := &Snapshot{
		QueryID:     p.id,
		CollectedAt: now,
		Items:       resp.Items,
	}
	col.mu.Lock()
	col.snapshots[p.id] = snap
	col.mu.Unlock()

	// Append to CSV.
	if err := col.appendCSV(p.id, now, resp.Items); err != nil {
		col.logger.Error("CSV write failed", "preset", p.id, "err", err)
	}
}

func (col *Collector) appendCSV(presetID string, collectedAt time.Time, items []any) error {
	filename := filepath.Join(col.outputDir, presetID+".csv")
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o640)
	if err != nil {
		return fmt.Errorf("opening CSV file: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	for _, item := range items {
		raw, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("marshalling item: %w", err)
		}
		if err := w.Write([]string{collectedAt.Format(time.RFC3339), string(raw)}); err != nil {
			return fmt.Errorf("writing CSV row: %w", err)
		}
	}
	return nil
}
