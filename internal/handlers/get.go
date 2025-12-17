package handlers

import "net/http"

func Get(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("dummy"))
}
