package handler

import (
	"log/slog"
	"net/http"
	"sort"

	"github.com/djdees/port-simulation/internal/domain"
)

func makeFacilitiesHandler(store *Store, _ *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := parsePagination(r)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, err.Error())
			return
		}
		kind := domain.FacilityKind(r.URL.Query().Get("kind"))
		if kind != "" && !kind.IsValid() {
			writeProblem(w, http.StatusBadRequest, "invalid kind value")
			return
		}
		facilities := make([]domain.Facility, 0, len(store.data.Facilities))
		for _, f := range store.data.Facilities {
			if kind != "" && f.Kind != kind {
				continue
			}
			facilities = append(facilities, f)
		}
		sort.Slice(facilities, func(i, j int) bool { return facilities[i].ID < facilities[j].ID })
		page := paginate(facilities, p)
		writeJSON(w, http.StatusOK, collectionResponse[domain.Facility]{
			Items: page, Total: len(facilities), Limit: p.limit, Offset: p.offset,
		})
	}
}

func makeFacilityByIDHandler(store *Store, _ *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		for _, f := range store.data.Facilities {
			if f.ID == id {
				writeJSON(w, http.StatusOK, f)
				return
			}
		}
		writeProblem(w, http.StatusNotFound, "facility not found")
	}
}
