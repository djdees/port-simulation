package handler

import (
	"log/slog"
	"net/http"
	"sort"

	"github.com/djdees/port-simulation/internal/domain"
)

func makeBargessHandler(store *Store, _ *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := parsePagination(r)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, err.Error())
			return
		}
		q := r.URL.Query()
		capability := domain.BargeCapability(q.Get("capability"))
		towID := q.Get("tow_id")

		if capability != "" && !capability.IsValid() {
			writeProblem(w, http.StatusBadRequest, "invalid capability value")
			return
		}

		// Build set of barge IDs in the requested tow for fast lookup.
		bargesInTow := make(map[string]bool)
		if towID != "" {
			for _, t := range store.data.Tows {
				if t.ID == towID {
					for _, bid := range t.BargeIDs {
						bargesInTow[bid] = true
					}
					break
				}
			}
			if len(bargesInTow) == 0 {
				// tow_id specified but not found — return empty collection
				writeJSON(w, http.StatusOK, collectionResponse[domain.Barge]{
					Items: []domain.Barge{}, Total: 0, Limit: p.limit, Offset: p.offset,
				})
				return
			}
		}

		barges := make([]domain.Barge, 0, len(store.data.Barges))
		for _, b := range store.data.Barges {
			if capability != "" && b.Capability != capability {
				continue
			}
			if towID != "" && !bargesInTow[b.ID] {
				continue
			}
			barges = append(barges, b)
		}

		sort.Slice(barges, func(i, j int) bool { return barges[i].ID < barges[j].ID })
		page := paginate(barges, p)
		writeJSON(w, http.StatusOK, collectionResponse[domain.Barge]{
			Items: page, Total: len(barges), Limit: p.limit, Offset: p.offset,
		})
	}
}

func makeBargeByIDHandler(store *Store, _ *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		for _, b := range store.data.Barges {
			if b.ID == id {
				writeJSON(w, http.StatusOK, b)
				return
			}
		}
		writeProblem(w, http.StatusNotFound, "barge not found")
	}
}
