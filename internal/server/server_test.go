package server

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"file-storage/internal/config"
	"file-storage/internal/files"
	"file-storage/internal/handlers/httpdto"
	"file-storage/internal/imgproc"
	"file-storage/internal/logger"
	"file-storage/internal/storage/inmemory"
	"fmt"
	"image/color"
	"io"
	"log"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/disintegration/imaging"
)

func TestServer_Authorization(t *testing.T) {

	cfg := config.Config{
		App: config.App{
			Host:      "127.0.0.1",
			Port:      8081,
			Timeout:   5 * time.Second,
			SizeLimit: 1024 * 1024,
			Security: config.Security{
				ReadToken:  "1",
				WriteToken: "2",
			},
		},
		Log:   config.Log{Level: config.LogLevelError, Type: config.LogTypeText},
		Image: config.Image{Ext: "jpeg", MaxDimension: 2000},
	}

	err := runServer(cfg, t)
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
			request:    newRequest("GET", "http://"+serverUrl+"/files/1/info", nil, t),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "delete",
			request:    newRequest("DELETE", "http://"+serverUrl+"/files/1/delete", nil, t),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "content",
			request:    newRequest("GET", "http://"+serverUrl+"/files/1/content", nil, t),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "upload",
			request:    newRequest("POST", "http://"+serverUrl+"/files/upload", nil, t),
			wantStatus: http.StatusForbidden,
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

	w := 100
	h := 100
	format := "jpeg"
	img := imaging.New(w, h, color.Black)
	imagingFormat, err := imaging.FormatFromExtension(format)
	if err != nil {
		t.Fatalf("test image format definition: %v", err)
	}

	uploadData, err := imgproc.Encode(img, imagingFormat)
	if err != nil {
		t.Fatalf("test image creation error: %v", err)
	}

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
		Image: config.Image{Ext: "jpeg", MaxDimension: 2000},
	}

	err = runServer(cfg, t)
	if err != nil {
		t.Fatalf("run server error: %v", err)
	}

	serverUrl := net.JoinHostPort(cfg.App.Host, strconv.Itoa(cfg.App.Port))

	client := http.DefaultClient

	//upload
	ur := httpdto.UploadRequest{
		ID:       "",
		Data:     uploadData,
		Metadata: map[string]any{"field1": 0, "field2": "0"},
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
		t.Errorf("got upload status %v want %v", response.StatusCode, http.StatusOK)
	}

	var uploadAnswer map[string]string
	decoder := json.NewDecoder(response.Body)
	defer response.Body.Close()

	err = decoder.Decode(&uploadAnswer)
	if err != nil {
		t.Fatalf("upload response decode error: %v", err)
	}

	id := uploadAnswer["id"]
	if len(id) != 36 {
		t.Fatalf("got empty id from upload want 36 symbols length")
	}

	//content
	request = newRequest("GET", "http://"+serverUrl+"/files/"+id+"/content", nil, t)
	request.Header.Add("Authorization", "Bearer 1")

	responseContent, err := client.Do(request)
	if err != nil {
		t.Errorf("content request error: %v", err)
	}

	if responseContent.StatusCode != http.StatusOK {
		t.Errorf("got content status %v want %v", responseContent.StatusCode, http.StatusOK)
	}

	b, err := io.ReadAll(responseContent.Body)
	defer responseContent.Body.Close()
	if err != nil {
		t.Errorf("content response body read error: %v", err)
	}

	if !bytes.Equal(b, uploadData) {
		t.Errorf("bytes are not equal")
	}

	//upload 2
	uploadData2 := []byte("updated data")
	ur2 := httpdto.UploadRequest{
		ID:       id,
		Data:     uploadData2,
		Metadata: map[string]any{"field1": 3.0, "field2": "1"},
	}
	sum = sha256.Sum256(ur2.Data)
	ur2.Hash = hex.EncodeToString(sum[:])
	dataJson, err = json.Marshal(ur2)
	if err != nil {
		t.Fatalf("marshal upload 2 error: %v", err)
	}

	request = newRequest("POST", "http://"+serverUrl+"/files/upload", dataJson, t)
	request.Header.Add("Authorization", "Bearer 2")

	response2, err := client.Do(request)
	if err != nil {
		t.Errorf("upload 2 request error: %v", err)
	}

	if response2.StatusCode != http.StatusOK {
		t.Errorf("got upload 2 status %v want %v", response2.StatusCode, http.StatusOK)
	}

	decoder = json.NewDecoder(response2.Body)
	defer response2.Body.Close()

	err = decoder.Decode(&uploadAnswer)
	if err != nil {
		t.Fatalf("upload response decode error: %v", err)
	}

	id2 := uploadAnswer["id"]
	if id != id2 {
		t.Fatalf("id mismatch, got %s want %s", id2, id)
	}

	//content 2
	requestC2 := newRequest("GET", "http://"+serverUrl+"/files/"+id+"/content", nil, t)
	requestC2.Header.Add("Authorization", "Bearer 2")

	responseContent2, err := client.Do(requestC2)
	if err != nil {
		t.Errorf("content 2 request error: %v", err)
	}

	if responseContent2.StatusCode != http.StatusOK {
		t.Errorf("got content 2 status %v want %v", responseContent2.StatusCode, http.StatusOK)
	}

	b, err = io.ReadAll(responseContent2.Body)
	defer responseContent2.Body.Close()
	if err != nil {
		t.Errorf("content 2 response body read error: %v", err)
	}

	if !bytes.Equal(b, uploadData2) {
		t.Errorf("bytes are not equal")
	}

	//info
	requestInfo := newRequest("GET", "http://"+serverUrl+"/files/"+id+"/info", nil, t)
	requestInfo.Header.Add("Authorization", "Bearer 1")

	responseInfo, err := client.Do(requestInfo)
	if err != nil {
		t.Errorf("info request error: %v", err)
	}

	if responseInfo.StatusCode != http.StatusOK {
		t.Errorf("info status %v want %v", responseInfo.StatusCode, http.StatusOK)
	}

	var infoAnswer map[string]any
	decoder = json.NewDecoder(responseInfo.Body)
	defer responseInfo.Body.Close()

	err = decoder.Decode(&infoAnswer)
	if err != nil {
		t.Errorf("info response body read error: %v", err)
	}

	if !reflect.DeepEqual(ur2.Metadata, infoAnswer["metadata"]) {
		t.Errorf("info metadata mismatch got %v want %v", infoAnswer["metadata"], ur2.Metadata)
	}

	//delete
	requestDelete := newRequest("DELETE", "http://"+serverUrl+"/files/"+id+"/delete", nil, t)
	requestDelete.Header.Add("Authorization", "Bearer 2")

	responseDelete, err := client.Do(requestDelete)
	if err != nil {
		t.Errorf("delete request error: %v", err)
	}

	if responseDelete.StatusCode != http.StatusOK {
		t.Errorf("delete status %v want %v", responseDelete.StatusCode, http.StatusOK)
	}

	//content 3
	requestC3 := newRequest("GET", "http://"+serverUrl+"/files/"+id+"/content", nil, t)
	requestC3.Header.Add("Authorization", "Bearer 2")

	responseContent3, err := client.Do(requestC3)
	if err != nil {
		t.Errorf("content 3 request error: %v", err)
	}

	if responseContent3.StatusCode != http.StatusNotFound {
		t.Errorf("got content 3 status %v want %v", responseContent3.StatusCode, http.StatusNotFound)
	}

}

func runServer(cfg config.Config, t *testing.T) error {

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

	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	})

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
