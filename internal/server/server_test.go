package server

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"file-storage/internal/config"
	"file-storage/internal/files"
	"file-storage/internal/handlers/models"
	"file-storage/internal/logger"
	"file-storage/internal/storage/inmemory"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestServer_Authorization(t *testing.T) {
	cfg := config.Config{
		App: config.App{
			Host:      "127.0.0.1",
			Port:      8080,
			Timeout:   5 * time.Second,
			SizeLimit: 1024 * 1024,
			Security: config.Security{
				ReadToken:  "1",
				WriteToken: "2",
			},
		},
		Log:   config.Log{Level: config.LogLevelError, Type: config.LogTypeText},
		Image: config.Image{Ext: "jpeg", MaxDimention: 2000},
	}

	err := runServer(cfg)
	if err != nil {
		t.Fatalf("run server error: %v", err)
	}

	serverUrl := net.JoinHostPort(cfg.App.Host, strconv.Itoa(cfg.App.Port))

	table := []struct {
		name       string
		request    *http.Request
		wantStatus int
	}{
		{
			name:       "info",
			request:    newRequest("Get", "http://"+serverUrl+"/files/1/info", nil, t),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "delete",
			request:    newRequest("Get", "http://"+serverUrl+"/files/1/ideletefo", nil, t),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "content",
			request:    newRequest("Get", "http://"+serverUrl+"/files/1/content", nil, t),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "upload",
			request:    newRequest("Get", "http://"+serverUrl+"/files/upload", nil, t),
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			client := http.DefaultClient
			response, err := client.Do(tt.request)
			if err != nil {
				t.Errorf("info request error: %v", err)
			}

			if response.StatusCode != tt.wantStatus {
				t.Errorf("got status %v want %v", response.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestServer_Lifecycle(t *testing.T) {
	cfg := config.Config{
		App: config.App{
			Host:      "127.0.0.1",
			Port:      8080,
			Timeout:   5 * time.Second,
			SizeLimit: 1024 * 1024,
			Security: config.Security{
				ReadToken:  "1",
				WriteToken: "2",
			},
		},
		Log:   config.Log{Level: config.LogLevelError, Type: config.LogTypeText},
		Image: config.Image{Ext: "jpeg", MaxDimention: 2000},
	}

	err := runServer(cfg)
	if err != nil {
		t.Fatalf("run server error: %v", err)
	}

	serverUrl := net.JoinHostPort(cfg.App.Host, strconv.Itoa(cfg.App.Port))

	client := http.DefaultClient

	ur := models.UploadRequest{
		ID:       "",
		Data:     []byte("some data"),
		Metadata: map[string]any{"field1": 1, "field2": "2"},
	}
	sum := sha256.Sum256(ur.Data)
	ur.Hash = hex.EncodeToString(sum[:])
	dataJson, err := json.Marshal(ur)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	request := newRequest("POST", "http://"+serverUrl+"/files/upload", dataJson, t)
	request.Header.Add("Authorization", "Bearer 2")

	response, err := client.Do(request)
	if err != nil {
		t.Errorf("upload request error: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("got status %v want %v", response.StatusCode, http.StatusOK)
	}
}

func runServer(cfg config.Config) error {

	serverUrl := net.JoinHostPort(cfg.App.Host, strconv.Itoa(cfg.App.Port))

	log := logger.NewBootstrap()
	store := inmemory.New()
	svc := files.NewService(&cfg.Image, store)
	srv := NewServer(&cfg.App, svc, log)

	ctx := context.Background()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Run(ctx, cfg.App.Security)
	}()

	err := waitServerUp(5*time.Second, serverUrl)
	if err != nil {
		return fmt.Errorf("wait server up error: %v", err)
	}

	return nil
}

func newRequest(method string, url string, body []byte, t *testing.T) *http.Request {

	reader := bytes.NewReader(body)
	r, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("new request error: %s %s %v", method, url, err)
	}
	return r
}

func waitServerUp(timeout time.Duration, serverUrl string) error {

	deadline := time.Now().Add(timeout)

	for {

		log.Print("wait for server")

		r, err := http.NewRequest("GET", "http://"+serverUrl+"/files/1/info", nil)
		if err != nil {
			return fmt.Errorf("check online request creation error: %v", err)
		}

		client := http.Client{Timeout: 1 * time.Second}
		resp, err := client.Do(r)
		if err != nil {
			err = fmt.Errorf("check online request exec error: %v", err)
		}

		if err == nil {
			resp.Body.Close()
			return nil
		}

		if time.Now().After(deadline) {
			return err
		}

		time.Sleep(50 * time.Millisecond)
	}
}
