package httpapi

import (
	"net/http"
	"strings"
)

type CORSOptions struct {
	AllowedOrigins []string
}

func corsMiddleware(opt CORSOptions) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(opt.AllowedOrigins))
	for _, o := range opt.AllowedOrigins {
		o = strings.TrimSpace(o)
		if o != "" {
			allowed[o] = struct{}{}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Если запрос не из браузера (нет Origin) — пропускаем как есть
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Разрешаем только из списка
			if _, ok := allowed[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
			}

			// preflight
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
