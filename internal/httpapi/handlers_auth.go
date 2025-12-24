package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"tarkovhelper-api/internal/config"
	"tarkovhelper-api/internal/models"
	"tarkovhelper-api/internal/repo"
	"tarkovhelper-api/internal/security"
)

func handleRegister(cfg config.Config, users *repo.UserRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "VALIDATION_ERROR", "Invalid request")
			return
		}

		req.Email = strings.TrimSpace(req.Email)
		req.Name = strings.TrimSpace(req.Name)

		if req.Email == "" || req.Name == "" || len(req.Password) < 6 {
			writeError(w, 400, "VALIDATION_ERROR", "Invalid request")
			return
		}

		hash, err := security.HashPassword(req.Password)
		if err != nil {
			writeError(w, 500, "INTERNAL_ERROR", "Something went wrong")
			return
		}

		u, err := users.Create(r.Context(), req.Email, req.Name, hash)
		if err != nil {
			if err == repo.ErrEmailTaken {
				writeError(w, 409, "EMAIL_TAKEN", "Email already registered")
				return
			}
			writeError(w, 500, "INTERNAL_ERROR", "Something went wrong")
			return
		}

		token, err := security.SignHS256(u.ID, cfg.JWTSecret, time.Duration(cfg.JWTTTLHours)*time.Hour)
		if err != nil {
			writeError(w, 500, "INTERNAL_ERROR", "Something went wrong")
			return
		}

		resp := models.AuthResponse{
			Token: token,
			User: models.User{
				ID:        u.ID,
				Email:     u.Email,
				Name:      u.Name,
				AvatarURL: u.AvatarURL,
			},
		}
		writeJSON(w, 200, resp)
	}
}

func handleLogin(cfg config.Config, users *repo.UserRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "VALIDATION_ERROR", "Invalid request")
			return
		}

		req.Email = strings.TrimSpace(req.Email)
		if req.Email == "" || len(req.Password) < 6 {
			writeError(w, 400, "VALIDATION_ERROR", "Invalid request")
			return
		}

		u, err := users.FindByEmail(r.Context(), req.Email)
		if err != nil {
			writeError(w, 401, "UNAUTHORIZED", "Invalid or missing token")
			return
		}

		if !security.CheckPassword(u.PasswordHash, req.Password) {
			writeError(w, 401, "UNAUTHORIZED", "Invalid or missing token")
			return
		}

		token, err := security.SignHS256(u.ID, cfg.JWTSecret, time.Duration(cfg.JWTTTLHours)*time.Hour)
		if err != nil {
			writeError(w, 500, "INTERNAL_ERROR", "Something went wrong")
			return
		}

		resp := models.AuthResponse{
			Token: token,
			User: models.User{
				ID:        u.ID,
				Email:     u.Email,
				Name:      u.Name,
				AvatarURL: u.AvatarURL,
			},
		}
		writeJSON(w, 200, resp)
	}
}
