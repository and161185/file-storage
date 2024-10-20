package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"file-storage/handlers"
	"file-storage/middleware"

	"github.com/gorilla/mux"
)

func main() {
	handlers.CreateUploadsDir()

	router := mux.NewRouter()
	router.Use(middleware.LoggingMiddleware)

	// Определяем маршруты
	router.HandleFunc("/upload", handlers.UploadHandler).Methods("POST")
	router.HandleFunc("/update/{file_id}", handlers.UpdateHandler).Methods("POST")
	router.HandleFunc("/download/{file_id}", handlers.DownloadHandler).Methods("GET")
	router.HandleFunc("/delete/{file_id}", handlers.DeleteHandler).Methods("DELETE")

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Запускаем сервер в горутине
	go func() {
		fmt.Println("Сервер запущен на порту :8080")
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
