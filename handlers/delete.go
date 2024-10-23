package handlers

import (
	"context"
	"encoding/json"
	"file-storage/models"
	"net/http"

	"github.com/gorilla/mux"
)

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	fileID := vars["file_id"]
	if fileID == "" {
		http.Error(w, "Параметр 'file_id' не указан", http.StatusBadRequest)
		return
	}

	err := storageService.DeleteFile(context.Background(), fileID)
	if err != nil {
		http.Error(w, "Ошибка при сохранении файла", http.StatusInternalServerError)
		return
	}

	// Возвращаем file_id в ответе
	response := models.UploadResponse{FileID: fileID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
