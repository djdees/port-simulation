package handler

import (
	"log/slog"
	"net/http"
	"sort"

	"github.com/djdees/port-simulation/internal/domain"
)

func makeBerthsHandler(store *Store, _ *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := parsePagination(r)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, err.Error())
			return
		}
		terminalID := r.URL.Query().Get("terminal_id")
		berths := make([]domain.Berth, 0, len(store.data.Berths))
		for _, b := range store.data.Berths {
			if terminalID != "" && b.TerminalID != terminalID {
				continue
			}
			berths = append(berths, b)
		}
		sort.Slice(berths, func(i, j int) bool { return berths[i].ID < berths[j].ID })
		page := paginate(berths, p)
		writeJSON(w, http.StatusOK, collectionResponse[domain.Berth]{
			Items: page, Total: len(berths), Limit: p.limit, Offset: p.offset,
		})
	}
}

func makeBerthByIDHandler(store *Store, _ *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		for _, b := range store.data.Berths {
			if b.ID == id {
				writeJSON(w, http.StatusOK, b)
				return
			}
		}
		writeProblem(w, http.StatusNotFound, "berth not found")
	}
}
