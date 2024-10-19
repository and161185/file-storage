package main

import (
	"fmt"
	"log"
	"net/http"

	"file-storage/handlers"
	"file-storage/middleware"

	"github.com/gorilla/mux"
)

func main() {
	handlers.CreateUploadsDir()

	router := mux.NewRouter()

	// Применяем middleware к роутеру
	router.Use(middleware.LoggingMiddleware)

	// Определяем маршруты
	router.HandleFunc("/upload", handlers.UploadHandler).Methods("POST")
	router.HandleFunc("/update/{file_id}", handlers.UpdateHandler).Methods("POST")
	router.HandleFunc("/download/{file_id}", handlers.DownloadHandler).Methods("GET")
	router.HandleFunc("/delete/{file_id}", handlers.DeleteHandler).Methods("DELETE")

	fmt.Println("Сервер запущен на порту :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
