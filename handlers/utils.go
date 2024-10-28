package handlers

import (
	"file-storage/storage"
)

var storageService storage.StorageService

// Функция для установки реализации хранилища
func SetStorageService(service storage.StorageService) {
	storageService = service
}
