package handlers

import (
	"net/http"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Устанавливаем статус 200 и записываем ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
