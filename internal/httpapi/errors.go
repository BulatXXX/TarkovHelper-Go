package httpapi

import (
	"encoding/json"
	"net/http"
	"tarkovhelper-api/internal/models"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	var env models.ErrorEnvelope
	env.Error.Code = code
	env.Error.Message = message
	writeJSON(w, status, env)
}
