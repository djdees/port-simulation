package handler

import (
	"log/slog"
	"net/http"
	"sort"

	"github.com/djdees/port-simulation/internal/domain"
)

func makeTerminalsHandler(store *Store, _ *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := parsePagination(r)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, err.Error())
			return
		}
		facilityID := r.URL.Query().Get("facility_id")
		terminals := make([]domain.Terminal, 0, len(store.data.Terminals))
		for _, t := range store.data.Terminals {
			if facilityID != "" && t.FacilityID != facilityID {
				continue
			}
			terminals = append(terminals, t)
		}
		sort.Slice(terminals, func(i, j int) bool { return terminals[i].ID < terminals[j].ID })
		page := paginate(terminals, p)
		writeJSON(w, http.StatusOK, collectionResponse[domain.Terminal]{
			Items: page, Total: len(terminals), Limit: p.limit, Offset: p.offset,
		})
	}
}

func makeTerminalByIDHandler(store *Store, _ *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		for _, t := range store.data.Terminals {
			if t.ID == id {
				writeJSON(w, http.StatusOK, t)
				return
			}
		}
		writeProblem(w, http.StatusNotFound, "terminal not found")
	}
}
