package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	fileID := vars["file_id"]
	if fileID == "" {
		http.Error(w, "Параметр 'file_id' не указан", http.StatusBadRequest)
		return
	}

	var uploadReq UploadRequest

	// Ограничиваем размер тела запроса (например, 50 МБ)
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20)

	// Читаем и декодируем JSON
	err := json.NewDecoder(r.Body).Decode(&uploadReq)
	if err != nil {
		http.Error(w, "Некорректный JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if uploadReq.Metadata.Filename == "" {
		http.Error(w, "Поля 'Metadata.Filename' и 'Data' обязательны", http.StatusBadRequest)
		return
	}

	if uploadReq.Data != "" {
		// Декодируем Base64 данные
		fileData, err := base64.StdEncoding.DecodeString(uploadReq.Data)
		if err != nil {
			http.Error(w, "Ошибка декодирования данных: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Путь для сохранения файла
		filePath := filepath.Join(storagePath(), fileID)

		/*
			if !fileExists(filePath) {
				http.Error(w, "не найден файл для обновления", http.StatusBadRequest)
				return
			}
		*/

		// Сохраняем файл
		err = os.WriteFile(filePath, fileData, 0644)
		if err != nil {
			http.Error(w, "Ошибка сохранения файла: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = saveMetadata(fileID, uploadReq.Metadata)
	if err != nil {
		http.Error(w, "Ошибка сохранения метаданных: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Возвращаем file_id в ответе
	response := UploadResponse{FileID: fileID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
