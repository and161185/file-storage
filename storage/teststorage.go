package storage

import (
	"context"
	"encoding/json"
	"file-storage/models"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
)

type FileSystemStorage struct {
	StoragePath   string
	MetadataStore string
}

var MetaMutex = &sync.Mutex{}

func NewTestStorage(path string) *FileSystemStorage {
	CreateUploadsDir(path)
	return &FileSystemStorage{StoragePath: path, MetadataStore: path + "/metadata.json"}
}

func CreateUploadsDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, 0755)
		if err != nil {
			log.Fatalf("не удалось создать директорию '%s': %v", path, err)
		}
	}
}

func (fs *FileSystemStorage) Ready(ctx context.Context) models.ReadyResponse {
	result := models.ReadyResponse{App: true, Database: true}

	return result
}

func (fs *FileSystemStorage) SaveFile(ctx context.Context, fileData []byte, metadata map[string]interface{}, fileID string) (string, error) {

	// Генерируем уникальный идентификатор файла
	if fileID == "" {
		fileID = uuid.New().String()
	}

	// Путь для сохранения файла
	filePath := filepath.Join(fs.StoragePath, fileID)

	// Сохраняем файл
	if len(fileData) != 0 {
		err := os.WriteFile(filePath, fileData, 0644)
		if err != nil {
			return "", err
		}
	}

	// Сохраняем метаданные
	err := fs.SaveMetadata(fileID, metadata)
	if err != nil {
		return "", err
	}

	return fileID, nil
}

func (fs *FileSystemStorage) GetFile(ctx context.Context, fileID string) ([]byte, map[string]interface{}, error) {

	filePath := filepath.Join(fs.StoragePath, fileID)

	var defaultMetadata map[string]interface{}

	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			err = fmt.Errorf("файл не найден")
		} else {
			err = fmt.Errorf("ошибка открытия файла: " + err.Error())
		}
		return nil, defaultMetadata, err
	}
	defer file.Close()

	// Получаем исходное имя файла
	metadata, err := fs.GetMetadata(fileID)
	if err != nil {
		return nil, defaultMetadata, err
	}

	// Читаем содержимое файла в память
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, defaultMetadata, err
	}

	return fileContent, metadata, nil
}

func (fs *FileSystemStorage) DeleteFile(ctx context.Context, fileID string) error {

	filePath := filepath.Join(fs.StoragePath, fileID)

	if !fileExists(filePath) {
		return fmt.Errorf("не найден файл для удаления %s", filePath)
	}

	err := os.Remove(filePath)
	if err != nil {
		return err
	}

	err = fs.DeleteMetadata(fileID)
	if err != nil {
		return err
	}

	return nil
}

func (fs *FileSystemStorage) GetMetadata(fileID string) (map[string]interface{}, error) {

	var defaultMetadata map[string]interface{}
	var metadata []map[string]interface{}

	data, err := os.ReadFile(fs.MetadataStore)
	if err != nil {
		return defaultMetadata, err
	}

	json.Unmarshal(data, &metadata)

	for _, m := range metadata {
		if m["fileID"] == fileID {
			return m, nil
		}
	}

	return defaultMetadata, fmt.Errorf("файл не найден")
}

func (fs *FileSystemStorage) SaveMetadata(fileID string, fileMetadata map[string]interface{}) error {
	MetaMutex.Lock()
	defer MetaMutex.Unlock()

	var metadata []map[string]interface{}

	fileMetadata["fileID"] = fileID

	data, err := os.ReadFile(fs.MetadataStore)
	if err != nil {
		// Если файл не найден, создаем пустой срез метаданных
		if os.IsNotExist(err) {
			metadata = []map[string]interface{}{}
		} else {
			return err // Возвращаем ошибку, если возникла другая ошибка
		}
	} else {
		// Парсим JSON
		err = json.Unmarshal(data, &metadata)
		if err != nil {
			return err
		}
	}

	idFound := false
	// Фильтруем метаданные, исключая запись с указанным fileID
	updatedMetadata := make([]map[string]interface{}, 0, len(metadata))
	for _, m := range metadata {
		if m["fileID"] == fileID {
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

	return os.WriteFile(fs.MetadataStore, data, 0644)
}

func (fs *FileSystemStorage) DeleteMetadata(fileID string) error {

	var metadata []map[string]interface{}

	// Читаем существующие метаданные
	data, err := os.ReadFile(fs.MetadataStore)
	if err != nil {
		return err
	}

	// Парсим JSON
	err = json.Unmarshal(data, &metadata)
	if err != nil {
		return err
	}

	// Фильтруем метаданные, исключая запись с указанным fileID
	updatedMetadata := make([]map[string]interface{}, 0, len(metadata))
	for _, m := range metadata {
		if m["fileID"] != fileID {
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
	err = os.WriteFile(fs.MetadataStore, data, 0644)
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
