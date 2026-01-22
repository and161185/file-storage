package files_test

import (
	"bytes"
	"context"
	"errors"
	"file-storage/internal/config"
	"file-storage/internal/errs"
	"file-storage/internal/filedata"
	"file-storage/internal/files"
	"file-storage/internal/imgproc"
	"fmt"
	"image/color"
	"io"
	"reflect"
	"testing"

	"github.com/disintegration/imaging"
)

type mockStorage struct {
	fnUpsert  func(ctx context.Context, fd *filedata.FileData) (string, error)
	fnInfo    func(ctx context.Context, ID string) (*filedata.FileInfo, error)
	fnContent func(ctx context.Context, ID string) (*filedata.ContentData, error)
	fnDelete  func(ctx context.Context, ID string) error
}

func (m *mockStorage) Upsert(ctx context.Context, fd *filedata.FileData) (string, error) {
	return m.fnUpsert(ctx, fd)
}
func (m *mockStorage) Info(ctx context.Context, ID string) (*filedata.FileInfo, error) {
	return m.fnInfo(ctx, ID)
}
func (m *mockStorage) Content(ctx context.Context, ID string) (*filedata.ContentData, error) {
	return m.fnContent(ctx, ID)
}
func (m *mockStorage) Delete(ctx context.Context, ID string) error {
	return m.fnDelete(ctx, ID)
}

func TestUpdate(t *testing.T) {

	var callUpsert bool
	ctx := context.Background()
	cfg := &config.Image{Ext: "jpeg", MaxDimention: 1000}

	storageError := fmt.Errorf("storage error")

	img := imaging.New(cfg.MaxDimention, cfg.MaxDimention, color.Black)
	imagingFormat, err := imaging.FormatFromExtension(cfg.Ext)
	if err != nil {
		t.Fatalf("test image format definition: %v", err)
	}
	b, err := imgproc.Encode(img, imagingFormat)
	if err != nil {
		t.Fatalf("test image encode error: %v", err)
	}

	table := []struct {
		name           string
		storage        *mockStorage
		uploadCommand  *filedata.UploadCommand
		wantErr        error
		wantID         string
		wantCallUpsert bool
	}{
		{
			name: "image error",
			storage: &mockStorage{
				fnUpsert: func(ctx context.Context, fd *filedata.FileData) (string, error) {
					callUpsert = true
					return "", nil
				},
			},
			uploadCommand: &filedata.UploadCommand{
				ID:      "1",
				IsImage: true,
				Data:    []byte("not an image")},
			wantErr:        errs.ErrInvalidImage,
			wantID:         "",
			wantCallUpsert: false,
		},
		{
			name: "storage error",
			storage: &mockStorage{
				fnUpsert: func(ctx context.Context, fd *filedata.FileData) (string, error) {
					callUpsert = true
					return "", storageError
				},
			},
			uploadCommand: &filedata.UploadCommand{
				ID:      "",
				IsImage: false,
				Data:    []byte("not an image")},
			wantErr:        storageError,
			wantID:         "",
			wantCallUpsert: true,
		},
		{
			name: "ok",
			storage: &mockStorage{
				fnUpsert: func(ctx context.Context, fd *filedata.FileData) (string, error) {
					callUpsert = true
					return "1", nil
				},
			},
			uploadCommand: &filedata.UploadCommand{
				ID:      "1",
				IsImage: true,
				Data:    b},
			wantErr:        nil,
			wantID:         "1",
			wantCallUpsert: true,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {

			callUpsert = false
			s := files.NewService(cfg, tt.storage)
			id, err := s.Update(ctx, tt.uploadCommand)

			if id != tt.wantID {
				t.Errorf("id mismatch got %s want %s", id, tt.wantID)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("error mismatch got %v want %v", err, tt.wantErr)
			}
			if callUpsert != tt.wantCallUpsert {
				t.Errorf("call upsert mismatch got %v want %v", callUpsert, tt.wantCallUpsert)
			}

		})
	}
}

func TestContent(t *testing.T) {

	var call bool
	zero := 0
	storageError := fmt.Errorf("storage error")

	ctx := context.Background()
	cfg := &config.Image{Ext: "jpeg", MaxDimention: 1000}

	img := imaging.New(cfg.MaxDimention, cfg.MaxDimention, color.Black)
	imagingFormat, err := imaging.FormatFromExtension(cfg.Ext)
	if err != nil {
		t.Fatalf("test image format definition: %v", err)
	}
	imgBytes, err := imgproc.Encode(img, imagingFormat)
	if err != nil {
		t.Fatalf("test image encode error: %v", err)
	}

	table := []struct {
		name           string
		storage        *mockStorage
		contentCommand *filedata.ContentCommand
		wantErr        error
		wantCall       bool
		wantBytes      []byte
	}{
		{
			name: "width error",
			storage: &mockStorage{
				fnContent: func(ctx context.Context, ID string) (*filedata.ContentData, error) {
					call = true
					return nil, nil
				},
			},
			contentCommand: &filedata.ContentCommand{
				ID:    "1",
				Width: &zero,
			},
			wantErr:   errs.ErrWrongUrlParameter,
			wantCall:  false,
			wantBytes: nil,
		},
		{
			name: "height error",
			storage: &mockStorage{
				fnContent: func(ctx context.Context, ID string) (*filedata.ContentData, error) {
					call = true
					return nil, nil
				},
			},
			contentCommand: &filedata.ContentCommand{
				ID:     "1",
				Height: &zero,
			},
			wantErr:   errs.ErrWrongUrlParameter,
			wantCall:  false,
			wantBytes: nil,
		},
		{
			name: "storage error",
			storage: &mockStorage{
				fnContent: func(ctx context.Context, ID string) (*filedata.ContentData, error) {
					call = true
					return nil, storageError
				},
			},
			contentCommand: &filedata.ContentCommand{
				ID: "1",
			},
			wantErr:   storageError,
			wantCall:  true,
			wantBytes: nil,
		},
		{
			name: "processing image error",
			storage: &mockStorage{
				fnContent: func(ctx context.Context, ID string) (*filedata.ContentData, error) {
					call = true
					b := []byte("not an image")
					data := io.NopCloser(bytes.NewReader(b))
					return &filedata.ContentData{Data: data, IsImage: true}, nil
				},
			},
			contentCommand: &filedata.ContentCommand{
				ID: "1",
			},
			wantErr:   errs.ErrInvalidImage,
			wantCall:  true,
			wantBytes: nil,
		},
		{
			name: "ok",
			storage: &mockStorage{
				fnContent: func(ctx context.Context, ID string) (*filedata.ContentData, error) {
					call = true
					data := io.NopCloser(bytes.NewReader(imgBytes))
					return &filedata.ContentData{Data: data, IsImage: true}, nil
				},
			},
			contentCommand: &filedata.ContentCommand{
				ID: "1",
			},
			wantErr:   nil,
			wantCall:  true,
			wantBytes: imgBytes,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			call = false
			s := files.NewService(cfg, tt.storage)
			b, err := s.Content(ctx, tt.contentCommand)

			if call != tt.wantCall {
				t.Errorf("call mismatch got %v want %v", call, tt.wantCall)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("error mismatch got %v want %v", err, tt.wantErr)
			}
			if tt.wantBytes != nil {
				if b == nil {
					t.Errorf("bytes mismatch got bytes want nil")
				}
			} else {
				if !bytes.Equal(tt.wantBytes, b) {
					t.Errorf("bytes mismatch")
				}
			}
		})
	}
}

func TestInfo(t *testing.T) {
	cfg := config.Image{Ext: "jpeg", MaxDimention: 1000}
	ctx := context.Background()

	storageError := fmt.Errorf("storage error")

	table := []struct {
		name         string
		storage      *mockStorage
		id           string
		wantError    error
		wantFileInfo *filedata.FileInfo
	}{
		{
			name: "storage error",
			storage: &mockStorage{fnInfo: func(ctx context.Context, ID string) (*filedata.FileInfo, error) {
				return nil, storageError
			}},
			id:           "1",
			wantError:    storageError,
			wantFileInfo: nil,
		},
		{
			name: "ok",
			storage: &mockStorage{fnInfo: func(ctx context.Context, ID string) (*filedata.FileInfo, error) {
				return &filedata.FileInfo{ID: "1"}, nil
			}},
			id:           "1",
			wantError:    nil,
			wantFileInfo: &filedata.FileInfo{ID: "1"},
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {

			s := files.NewService(&cfg, tt.storage)

			fi, err := s.Info(ctx, tt.id)

			if !reflect.DeepEqual(fi, tt.wantFileInfo) {
				t.Errorf("info result mismatch got %v want %v", fi, tt.wantFileInfo)
			}

			if !errors.Is(err, tt.wantError) {
				t.Errorf("errors mismatch got %v want %v", err, tt.wantError)
			}
		})
	}
}

func TestDelete(t *testing.T) {

	ctx := context.Background()
	cfg := config.Image{Ext: "jpeg", MaxDimention: 1000}
	storageError := fmt.Errorf("storage error")

	table := []struct {
		name    string
		storage *mockStorage
		id      string
		wantErr error
	}{
		{
			name: "storage error",
			storage: &mockStorage{fnDelete: func(ctx context.Context, ID string) error {
				return storageError
			}},
			id:      "1",
			wantErr: storageError,
		},
		{
			name: "ok",
			storage: &mockStorage{fnDelete: func(ctx context.Context, ID string) error {
				return nil
			}},
			id:      "1",
			wantErr: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			s := files.NewService(&cfg, tt.storage)
			err := s.Delete(ctx, tt.id)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("errors mismatch got %v want %v", err, tt.wantErr)
			}
		})
	}
}
