package handler

import (
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"github.com/djdees/port-simulation/internal/domain"
)

func makeCargoHandler(store *Store, _ *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := parsePagination(r)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, err.Error())
			return
		}
		q := r.URL.Query()
		commodity := domain.CommodityType(q.Get("commodity"))
		origin := q.Get("origin")
		destination := q.Get("destination")
		vesselID := q.Get("vessel_id")
		bargeID := q.Get("barge_id")

		if commodity != "" && !commodity.IsValid() {
			writeProblem(w, http.StatusBadRequest, "invalid commodity value")
			return
		}

		cargo := make([]domain.Cargo, 0, len(store.data.Cargo))
		for _, c := range store.data.Cargo {
			if commodity != "" && c.Commodity != commodity {
				continue
			}
			if origin != "" && !strings.EqualFold(c.Origin, origin) {
				continue
			}
			if destination != "" && !strings.EqualFold(c.Destination, destination) {
				continue
			}
			if vesselID != "" && c.VesselID != vesselID {
				continue
			}
			if bargeID != "" && c.BargeID != bargeID {
				continue
			}
			cargo = append(cargo, c)
		}

		sort.Slice(cargo, func(i, j int) bool { return cargo[i].ID < cargo[j].ID })
		page := paginate(cargo, p)
		writeJSON(w, http.StatusOK, collectionResponse[domain.Cargo]{
			Items: page, Total: len(cargo), Limit: p.limit, Offset: p.offset,
		})
	}
}

func makeCargoByIDHandler(store *Store, _ *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		for _, c := range store.data.Cargo {
			if c.ID == id {
				writeJSON(w, http.StatusOK, c)
				return
			}
		}
		writeProblem(w, http.StatusNotFound, "cargo record not found")
	}
}
