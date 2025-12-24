package httpapi

import (
	"net/http"

	"tarkovhelper-api/internal/models"
	"tarkovhelper-api/internal/repo"
)

func handleMe(users *repo.UserRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := UserIDFromCtx(r.Context())
		if !ok {
			writeError(w, 401, "UNAUTHORIZED", "Invalid or missing token")
			return
		}

		u, err := users.FindByID(r.Context(), userID)
		if err != nil {
			writeError(w, 401, "UNAUTHORIZED", "Invalid or missing token")
			return
		}

		writeJSON(w, 200, models.MeResponse{
			User: models.User{
				ID:        u.ID,
				Email:     u.Email,
				Name:      u.Name,
				AvatarURL: u.AvatarURL,
			},
		})
	}
}
