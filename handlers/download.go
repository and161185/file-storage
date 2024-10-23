package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"file-storage/models"
	"net/http"

	"github.com/gorilla/mux"
)

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	fileID := vars["file_id"]
	if fileID == "" {
		http.Error(w, "Параметр 'file_id' не указан", http.StatusBadRequest)
		return
	}

	fileContent, metadata, err := storageService.GetFile(context.Background(), fileID)
	if err != nil {
		http.Error(w, "Ошибка поучения данных: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Кодируем содержимое файла в Base64
	encodedContent := base64.StdEncoding.EncodeToString(fileContent)

	// Отправляем файл
	response := models.DownloadResponse{Metadata: metadata, Data: encodedContent}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
