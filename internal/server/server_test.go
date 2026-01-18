package server

import (
	"context"
	"file-storage/internal/config"
	"file-storage/internal/files"
	"file-storage/internal/logger"
	"file-storage/internal/storage/inmemory"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	port := 8081

	cfg := config.Config{
		App: config.App{
			Port:      port,
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

	serverUrl := fmt.Sprintf("http://localhost:%d", port)

	log := logger.NewBootstrap()
	store := inmemory.New()
	svc := files.NewService(&cfg.Image, store)
	srv := NewServer(&cfg.App, svc, log)

	ctx := context.Background()
	srv.Run(ctx, cfg.App.Security)

	err := waitServerUp(5*time.Second, serverUrl)
	if err != nil {
		t.Fatalf("wait server up error: %v", err)
	}

}

func waitServerUp(timeout time.Duration, serverUrl string) error {

	deadline := time.Now().Add(timeout)

	for {

		r, err := http.NewRequest("GET", serverUrl+"/files/1/info", nil)
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

		time.Sleep(100 * time.Millisecond)
	}
}
