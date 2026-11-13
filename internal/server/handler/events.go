package handler

import (
	"log/slog"
	"net/http"
	"sort"
	"time"

	"github.com/djdees/port-simulation/internal/domain"
)

func makeArrivalsHandler(store *Store, _ *slog.Logger) http.HandlerFunc {
	return makeEventHandler(store, domain.EventTypeArrival)
}

func makeDeparturesHandler(store *Store, _ *slog.Logger) http.HandlerFunc {
	return makeEventHandler(store, domain.EventTypeDeparture)
}

func makeEventHandler(store *Store, eventType domain.EventType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := parsePagination(r)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, err.Error())
			return
		}

		q := r.URL.Query()
		facilityID := q.Get("facility_id")

		var since, until time.Time
		if v := q.Get("since"); v != "" {
			since, err = time.Parse(time.RFC3339, v)
			if err != nil {
				writeProblem(w, http.StatusBadRequest, "since must be RFC3339 format")
				return
			}
		}
		if v := q.Get("until"); v != "" {
			until, err = time.Parse(time.RFC3339, v)
			if err != nil {
				writeProblem(w, http.StatusBadRequest, "until must be RFC3339 format")
				return
			}
		}

		events := make([]domain.TrafficEvent, 0)
		for _, e := range store.data.Events {
			if e.EventType != eventType {
				continue
			}
			if facilityID != "" && e.FacilityID != facilityID {
				continue
			}
			if !since.IsZero() && e.OccurredAt.Before(since) {
				continue
			}
			if !until.IsZero() && e.OccurredAt.After(until) {
				continue
			}
			events = append(events, e)
		}

		// Deterministic order: ascending occurred_at, then ID
		sort.Slice(events, func(i, j int) bool {
			if events[i].OccurredAt.Equal(events[j].OccurredAt) {
				return events[i].ID < events[j].ID
			}
			return events[i].OccurredAt.Before(events[j].OccurredAt)
		})

		page := paginate(events, p)
		writeJSON(w, http.StatusOK, collectionResponse[domain.TrafficEvent]{
			Items:  page,
			Total:  len(events),
			Limit:  p.limit,
			Offset: p.offset,
		})
	}
}
