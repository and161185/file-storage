package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"file-storage/imageutils"
	"file-storage/models"
	"net/http"

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

	var uploadReq models.UploadRequest

	// Ограничиваем размер тела запроса (например, 50 МБ)
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20)

	// Читаем и декодируем JSON
	err := json.NewDecoder(r.Body).Decode(&uploadReq)
	if err != nil {
		http.Error(w, "Некорректный JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if uploadReq.Metadata.FileName == "" {
		http.Error(w, "Поля 'Metadata.Filename' и 'Data' обязательны", http.StatusBadRequest)
		return
	}

	var fileData []byte
	if uploadReq.Data != "" {
		// Декодируем Base64 данные
		fileData, err = base64.StdEncoding.DecodeString(uploadReq.Data)
		if err != nil {
			http.Error(w, "Ошибка декодирования данных: "+err.Error(), http.StatusBadRequest)
			return
		}

		fileData, err = imageutils.ConvertToJPEG(fileData)
		if err != nil {
			http.Error(w, "Ошибка преобразования в JPEG: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Сохранение файла с помощью выбранного хранилища
	_, err = storageService.SaveFile(context.Background(), fileData, uploadReq.Metadata, fileID)
	if err != nil {
		http.Error(w, "Ошибка при сохранении файла", http.StatusInternalServerError)
		return
	}

	// Возвращаем file_id в ответе
	response := models.UploadResponse{FileID: fileID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
