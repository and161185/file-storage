package handlers

import (
	"context"
	"net/http"
	"testing"
)

func TestContentHandler(t *testing.T) {

	table := []struct {
		name       string
		service    *mockService
		ctx        context.Context
		request    *http.Request
		wantStatus int
	}{}
}
