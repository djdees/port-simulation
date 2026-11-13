package handler

import (
	"log/slog"
	"net/http"
	"sort"

	"github.com/djdees/port-simulation/internal/domain"
)

func makeTowsHandler(store *Store, _ *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := parsePagination(r)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, err.Error())
			return
		}
		q := r.URL.Query()
		towboatID := q.Get("towboat_id")
		bargeID := q.Get("barge_id")

		tows := make([]domain.Tow, 0, len(store.data.Tows))
		for _, t := range store.data.Tows {
			if towboatID != "" && t.TowboatID != towboatID {
				continue
			}
			if bargeID != "" {
				found := false
				for _, bid := range t.BargeIDs {
					if bid == bargeID {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			tows = append(tows, t)
		}

		sort.Slice(tows, func(i, j int) bool { return tows[i].ID < tows[j].ID })
		page := paginate(tows, p)
		writeJSON(w, http.StatusOK, collectionResponse[domain.Tow]{
			Items: page, Total: len(tows), Limit: p.limit, Offset: p.offset,
		})
	}
}

func makeTowByIDHandler(store *Store, _ *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		for _, t := range store.data.Tows {
			if t.ID == id {
				writeJSON(w, http.StatusOK, t)
				return
			}
		}
		writeProblem(w, http.StatusNotFound, "tow not found")
	}
}
