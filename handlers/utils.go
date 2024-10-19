package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

var (
	MetadataStore = storagePath() + "/metadata.json"
	MetaMutex     = &sync.Mutex{}
)

func CreateUploadsDir() {
	storagePath := storagePath()
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		err := os.Mkdir(storagePath, 0755)
		if err != nil {
			log.Fatalf("не удалось создать директорию '%s': %v", storagePath, err)
		}
	}
}

func getMetadata(fileID string) (FileMetadata, error) {

	defaultMetadata := FileMetadata{}
	var metadata []FileMetadata

	data, err := os.ReadFile(MetadataStore)
	if err != nil {
		return defaultMetadata, err
	}

	json.Unmarshal(data, &metadata)

	for _, m := range metadata {
		if m.FileID == fileID {
			return m, nil
		}
	}

	return defaultMetadata, fmt.Errorf("файл не найден")
}

func saveMetadata(fileID string, fileMetadata FileMetadata) error {
	MetaMutex.Lock()
	defer MetaMutex.Unlock()

	var metadata []FileMetadata

	fileMetadata.FileID = fileID

	// Читаем существующие метаданные
	data, err := os.ReadFile(MetadataStore)
	if err == nil {
		json.Unmarshal(data, &metadata)
	}

	// Парсим JSON
	err = json.Unmarshal(data, &metadata)
	if err != nil {
		return err
	}

	idFound := false
	// Фильтруем метаданные, исключая запись с указанным fileID
	updatedMetadata := make([]FileMetadata, 0, len(metadata))
	for _, m := range metadata {
		if m.FileID == fileID {
			updatedMetadata = append(updatedMetadata, fileMetadata)
			idFound = true
		} else {
			updatedMetadata = append(updatedMetadata, m)
		}
	}

	// Добавляем новые метаданные
	if !idFound {
		updatedMetadata = append(updatedMetadata, fileMetadata)
	}

	// Сохраняем обратно в файл
	data, err = json.MarshalIndent(updatedMetadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(MetadataStore, data, 0644)
}

func DeleteMetadata(fileID string) error {

	var metadata []FileMetadata

	// Читаем существующие метаданные
	data, err := os.ReadFile(MetadataStore)
	if err != nil {
		return err
	}

	// Парсим JSON
	err = json.Unmarshal(data, &metadata)
	if err != nil {
		return err
	}

	// Фильтруем метаданные, исключая запись с указанным fileID
	updatedMetadata := make([]FileMetadata, 0, len(metadata))
	for _, m := range metadata {
		if m.FileID != fileID {
			updatedMetadata = append(updatedMetadata, m)
		}
	}

	// Проверяем, изменились ли метаданные (если нет - файл не изменяем)
	if len(updatedMetadata) == len(metadata) {
		return err
	}

	// Сериализуем обновлённые метаданные в JSON
	data, err = json.MarshalIndent(updatedMetadata, "", "  ")
	if err != nil {
		return err
	}

	// Сохраняем обновлённые метаданные в файл
	err = os.WriteFile(MetadataStore, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir() // Убедимся, что это файл, а не директория
}

func storagePath() string {
	return "storage"
}
