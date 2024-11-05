package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"file-storage/handlers"
	"file-storage/middleware"
	"file-storage/models"
	"file-storage/storage"

	"github.com/gorilla/mux"
)

var version = "undefined"
var storageService storage.StorageService

type Config = models.Config

func LoadConfig(filename string) (*Config, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл настроек: %v", err)
	}

	var config Config
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, fmt.Errorf("не удалось преобразовать данные настроек в JSON: %v", err)
	}

	return &config, nil
}

func main() {

	config, err := LoadConfig("config.json")
	if err != nil {
		fmt.Println("ошибка загрузки конфигурации:", err)
		return
	}

	// Выбираем реализацию в зависимости от конфигурации
	if config.Features.Test {
		storageService = storage.NewTestStorage("uploads")
	} else {
		configMongo := config.Mongo
		storageService, _ = storage.NewMongoStorage(&configMongo)
	}

	// Используем storageService в хендлерах
	handlers.SetStorageService(storageService)

	router := mux.NewRouter()
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.AuthMiddleware(config.Tokens.GeneralToken, config.Tokens.DownloadToken))
	router.Use(middleware.GCMiddleware)

	// Определяем маршруты
	router.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
			return
		}

		fmt.Fprintf(w, "Version: %s", version)
	}).Methods("GET")
	router.HandleFunc("/upload", handlers.UploadHandler).Methods("POST")
	router.HandleFunc("/update/{file_id}", handlers.UpdateHandler).Methods("POST")
	router.HandleFunc("/download/{file_id}", handlers.DownloadHandler).Methods("GET")
	router.HandleFunc("/delete/{file_id}", handlers.DeleteHandler).Methods("DELETE")

	strPort := strconv.Itoa(config.Application.Port)
	srv := &http.Server{
		Addr:           ":" + strPort,
		Handler:        router,
		MaxHeaderBytes: 500 << 20,
	}

	// Запускаем сервер в горутине
	go func() {
		fmt.Println("Сервер запущен на порту :" + strPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v\n", err)
		}
	}()

	// Канал для прослушивания сигналов завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Блокируем выполнение до получения сигнала
	<-stop

	fmt.Println("\nПолучен сигнал завершения, выключение сервера...")

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Останавливаем сервер корректно
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка при завершении работы сервера: %v\n", err)
	}

	fmt.Println("Сервер завершил работу.")
}
