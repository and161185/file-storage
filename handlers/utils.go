package handlers

import (
	"file-storage/models"
	"file-storage/storage"
)

type FileMetadata = models.FileMetadata

var storageService storage.StorageService

// Функция для установки реализации хранилища
func SetStorageService(service storage.StorageService) {
	storageService = service
}
