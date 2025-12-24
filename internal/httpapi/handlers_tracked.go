package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"tarkovhelper-api/internal/models"
	"tarkovhelper-api/internal/repo"
)

func parseMode(r *http.Request) (models.Mode, bool) {
	mode := models.Mode(strings.TrimSpace(r.URL.Query().Get("mode")))
	return mode, mode.Valid()
}

func validateTracked(items []models.TrackedItem) bool {
	for _, it := range items {
		if strings.TrimSpace(it.ID) == "" {
			return false
		}
		if it.UpdatedAt < 0 {
			return false
		}
	}
	return true
}

func handleGetTracked(tr *repo.TrackedRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := UserIDFromCtx(r.Context())
		if !ok {
			writeError(w, 401, "UNAUTHORIZED", "Invalid or missing token")
			return
		}

		mode, ok := parseMode(r)
		if !ok {
			writeError(w, 400, "VALIDATION_ERROR", "Invalid request")
			return
		}

		items, err := tr.Get(r.Context(), userID, mode)
		if err != nil {
			writeError(w, 500, "INTERNAL_ERROR", "Something went wrong")
			return
		}

		writeJSON(w, 200, models.GetTrackedResponse{Items: items})
	}
}

func handlePutTracked(tr *repo.TrackedRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := UserIDFromCtx(r.Context())
		if !ok {
			writeError(w, 401, "UNAUTHORIZED", "Invalid or missing token")
			return
		}

		mode, ok := parseMode(r)
		if !ok {
			writeError(w, 400, "VALIDATION_ERROR", "Invalid request")
			return
		}

		var req models.PutTrackedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "VALIDATION_ERROR", "Invalid request")
			return
		}

		if req.Items == nil {
			req.Items = []models.TrackedItem{}
		}
		if !validateTracked(req.Items) {
			writeError(w, 400, "VALIDATION_ERROR", "Invalid request")
			return
		}

		items, err := tr.Put(r.Context(), userID, mode, req.Items)
		if err != nil {
			writeError(w, 500, "INTERNAL_ERROR", "Something went wrong")
			return
		}

		writeJSON(w, 200, models.PutTrackedResponse{Items: items})
	}
}
