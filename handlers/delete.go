package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

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

	filePath := filepath.Join(storagePath(), fileID)

	if !fileExists(filePath) {
		http.Error(w, "не найден файл для удаления", http.StatusBadRequest)
		return
	}

	err := os.Remove(filePath)
	if err != nil {
		http.Error(w, "ошибка при удалении файла: "+err.Error(), http.StatusBadRequest)
		return
	}

	DeleteMetadata(fileID)

	// Возвращаем file_id в ответе
	response := UploadResponse{FileID: fileID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
