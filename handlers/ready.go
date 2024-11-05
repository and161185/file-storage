package handlers

import (
	"context"
	"encoding/json"
	"net/http"
)

func ReadyHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	readyStruct := storageService.Ready(context.Background())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(readyStruct)
}
