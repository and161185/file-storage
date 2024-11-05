package middleware

import (
	"log"
	"net/http"
	"runtime"
	"strings"
)

type Middleware func(http.Handler) http.Handler

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

// Функция, создающая middleware с токенами из конфигурации
func AuthMiddleware(generalToken, downloadToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/version") ||
				strings.HasPrefix(r.URL.Path, "/health") ||
				strings.HasPrefix(r.URL.Path, "/ready") ||
				strings.HasPrefix(r.URL.Path, "/metrics") {
				next.ServeHTTP(w, r)
				return
			}

			token := r.URL.Query().Get("token")
			if strings.HasPrefix(r.URL.Path, "/download") && r.Method == http.MethodGet {
				// Проверка специального токена для маршрута загрузки
				if token != downloadToken && token != generalToken {
					http.Error(w, "Unauthorized - Invalid token", http.StatusUnauthorized)
					return
				}
			} else {
				// Проверка общего токена для всех остальных маршрутов
				if token != generalToken {
					http.Error(w, "Unauthorized - Invalid token", http.StatusUnauthorized)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func GCMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r) // Вызов обработчика
		if strings.HasPrefix(r.URL.Path, "/upload") || strings.HasPrefix(r.URL.Path, "/update") {

			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			if m.Alloc > 2*1024*1024*1024 {
				runtime.GC()
			}
		}
	})
}
