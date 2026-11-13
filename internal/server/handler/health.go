package handler

import "net/http"

// handleLive always returns 200 OK — it confirms the process is running.
func handleLive(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// makeReadyHandler returns 200 when the data store is loaded, 503 otherwise.
func makeReadyHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if store == nil || store.data == nil {
			writeProblem(w, http.StatusServiceUnavailable, "data store not ready")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	}
}
