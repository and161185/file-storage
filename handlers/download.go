package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"file-storage/imageutils"
	"file-storage/models"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

type QueryParams = models.QueryParams

var decoder = schema.NewDecoder()

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
		http.Error(w, "Ошибка получения данных: "+err.Error(), http.StatusBadRequest)
		return
	}

	queryParams := QueryParams{}
	decoder.IgnoreUnknownKeys(true)
	if err := decoder.Decode(&queryParams, r.URL.Query()); err != nil {
		http.Error(w, "Некорректные параметры запроса: "+err.Error(), http.StatusBadRequest)
		return
	}

	isImage, ok := metadata["is_image"].(bool)
	if ok && isImage {
		fileContent, err = imageutils.ConvertImageFromJPEG(fileContent, queryParams)
		if err != nil {
			http.Error(w, "Ошибка преобразования изображения: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	if !queryParams.FileOnly {

		// Кодируем содержимое файла в Base64
		encodedContent := base64.StdEncoding.EncodeToString(fileContent)

		// Отправляем файл
		response := models.DownloadResponse{Metadata: metadata, Data: encodedContent}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		mimeType := http.DetectContentType(fileContent)
		w.Header().Set("Content-Type", mimeType)

		// Отправляем файл
		_, err := w.Write(fileContent)
		if err != nil {
			http.Error(w, "Ошибка при отправке файла: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

}
