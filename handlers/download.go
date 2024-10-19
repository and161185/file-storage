package handlers

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"

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

	filePath := filepath.Join(storagePath(), fileID)

	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Файл не найден", http.StatusNotFound)
		} else {
			http.Error(w, "Ошибка открытия файла: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	defer file.Close()

	// Получаем исходное имя файла
	metadata, err := getMetadata(fileID)
	if err != nil {
		http.Error(w, "Ошибка получения метаданных: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Читаем содержимое файла в память
	fileContent, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Ошибка чтения файла: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Кодируем содержимое файла в Base64
	encodedContent := base64.StdEncoding.EncodeToString(fileContent)

	// Отправляем файл
	response := DownloadResponse{Metadata: metadata, Data: encodedContent}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
