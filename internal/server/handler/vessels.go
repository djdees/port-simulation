package handler

import (
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"github.com/djdees/port-simulation/internal/domain"
)

func makeVesselsHandler(store *Store, _ *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := parsePagination(r)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, err.Error())
			return
		}
		q := r.URL.Query()
		kind := domain.VesselKind(q.Get("kind"))
		flagCountry := strings.ToUpper(q.Get("flag_country"))

		if kind != "" && !kind.IsValid() {
			writeProblem(w, http.StatusBadRequest, "invalid kind value")
			return
		}

		vessels := make([]domain.Vessel, 0, len(store.data.Vessels))
		for _, v := range store.data.Vessels {
			if kind != "" && v.Kind != kind {
				continue
			}
			if flagCountry != "" && !strings.EqualFold(v.FlagCountry, flagCountry) {
				continue
			}
			vessels = append(vessels, v)
		}

		// Deterministic order by ID
		sort.Slice(vessels, func(i, j int) bool { return vessels[i].ID < vessels[j].ID })

		page := paginate(vessels, p)
		writeJSON(w, http.StatusOK, collectionResponse[domain.Vessel]{
			Items:  page,
			Total:  len(vessels),
			Limit:  p.limit,
			Offset: p.offset,
		})
	}
}

func makeVesselByIDHandler(store *Store, _ *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		for _, v := range store.data.Vessels {
			if v.ID == id {
				writeJSON(w, http.StatusOK, v)
				return
			}
		}
		writeProblem(w, http.StatusNotFound, "vessel not found")
	}
}
