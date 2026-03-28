package filesystemstorage

import (
	"bytes"
	"context"
	"errors"
	"file-storage/internal/config"
	"file-storage/internal/contextkeys"
	"file-storage/internal/errs"
	"file-storage/internal/filedata"
	"file-storage/internal/logger"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestUpsert(t *testing.T) {
	id := "123456789012345678901234567890123456"
	data := []byte("some data")

	log := logger.NewBootstrap()
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.ContextKeyLogger, log)

	table := []struct {
		name    string
		path    string
		fd      *filedata.FileData
		wantErr bool
		wantID  string
	}{
		{
			name:    "invalid filedata",
			path:    t.TempDir(),
			fd:      nil,
			wantErr: true,
			wantID:  "",
		},
		{
			name:    "invalid filedata 2",
			path:    t.TempDir(),
			fd:      &filedata.FileData{},
			wantErr: true,
			wantID:  "",
		},
		{
			name:    "ok",
			path:    t.TempDir(),
			fd:      &filedata.FileData{ID: id, Data: data, HashSource: "123", IsImage: false},
			wantErr: false,
			wantID:  id,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.FileSystem{Path: tt.path}
			f := New(&cfg, log)
			id, err := f.Upsert(ctx, tt.fd)

			if tt.wantErr && err == nil {
				t.Errorf("got nil want error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("got error want nil")
			}
			if tt.wantID != id {
				t.Errorf("got id %s want %s", id, tt.wantID)
			}

			if tt.name == "ok" {
				dirPath, err := fileCatalog(tt.path, id)
				if err != nil {
					t.Errorf("catalog name error: %v", err)
				}

				basePath := filepath.Join(dirPath, id)
				v, _, err := slotInfo(dirPath, id)
				if err != nil {
					t.Errorf("read versions error: %v", err)
				}

				dataFile := dataFileName(basePath, v)
				_, err = os.Stat(dataFile)
				if err != nil {
					t.Errorf("data file %s not created", dataFile)
				}

				metadataFile := metadataFileName(basePath, v)
				_, err = os.Stat(metadataFile)
				if err != nil {
					t.Errorf("meta file %s not created", metadataFile)
				}
			}
		})
	}
}

func TestDelete(t *testing.T) {
	log := logger.NewBootstrap()
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.ContextKeyLogger, log)

	path := t.TempDir()
	cfg := config.FileSystem{Path: path}
	fUpsert := New(&cfg, log)

	id := "123456789012345678901234567890123456"
	data := []byte("some data")
	fd := &filedata.FileData{ID: id, Data: data, HashSource: "123", IsImage: false}
	id, err := fUpsert.Upsert(ctx, fd)
	if err != nil {
		t.Fatalf("upsert error %v", err)
	}

	table := []struct {
		name      string
		id        string
		path      string
		wantError error
	}{
		{
			name:      "delete err",
			id:        "",
			path:      path,
			wantError: errs.ErrInvalidID,
		},
		{
			name:      "delete from not exesting dir",
			id:        id,
			path:      "/not/existing/derectory",
			wantError: nil,
		},
		{
			name:      "delete ok",
			id:        id,
			path:      path,
			wantError: nil,
		},
		{
			name:      "delete idempotent",
			id:        id,
			path:      path,
			wantError: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.FileSystem{Path: tt.path}
			f := New(&cfg, log)

			err := f.Delete(ctx, tt.id)
			if !errors.Is(err, tt.wantError) {
				t.Errorf("got %v want %v", err, tt.wantError)
			}

			if tt.name == "delete ok" || tt.name == "delete idempotent" {
				dirPath, err := fileCatalog(tt.path, id)
				if err != nil {
					t.Errorf("catalog name error: %v", err)
				}

				filesToRemove, err := filenamesByID(dirPath, id)
				if err != nil {
					t.Errorf("files to remove search error: %v", err)
				}

				for _, filename := range filesToRemove {
					t.Errorf("file not deleted: %s", filename)
				}
			}
		})
	}
}

func TestInfo(t *testing.T) {
	log := logger.NewBootstrap()
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.ContextKeyLogger, log)

	path := t.TempDir()
	cfg := config.FileSystem{Path: path}
	f := New(&cfg, log)

	id := "123456789012345678901234567890123456"
	data := []byte("some data")
	fd := &filedata.FileData{ID: id, Data: data, HashSource: "123", IsImage: false}
	wantFi := &filedata.FileInfo{ID: id, HashSource: "123", IsImage: false}
	id, err := f.Upsert(ctx, fd)
	if err != nil {
		t.Fatalf("upsert error %v", err)
	}

	table := []struct {
		name    string
		id      string
		wantErr error
	}{
		{
			name:    "invalid id",
			id:      "",
			wantErr: errs.ErrInvalidID,
		},
		{
			name:    "not found",
			id:      "023456789012345678901234567890123456",
			wantErr: errs.ErrNotFound,
		},
		{
			name:    "ok",
			id:      id,
			wantErr: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			fi, err := f.Info(ctx, tt.id)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got %v want %v", err, tt.wantErr)
			}

			if tt.wantErr == nil {
				if !reflect.DeepEqual(fi, wantFi) {
					t.Errorf("file information mismatch got %v want %v", fi, wantFi)
				}
			}
		})
	}
}

func TestContent(t *testing.T) {
	log := logger.NewBootstrap()
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.ContextKeyLogger, log)

	path := t.TempDir()
	cfg := config.FileSystem{Path: path}
	f := New(&cfg, log)

	id := "123456789012345678901234567890123456"
	data := []byte("some data")
	fd := &filedata.FileData{ID: id, Data: data, HashSource: "123", IsImage: false}

	id, err := f.Upsert(ctx, fd)
	if err != nil {
		t.Fatalf("upsert error %v", err)
	}

	table := []struct {
		name    string
		id      string
		wantErr error
	}{
		{
			name:    "invalid id",
			id:      "",
			wantErr: errs.ErrInvalidID,
		},
		{
			name:    "not found",
			id:      "023456789012345678901234567890123456",
			wantErr: errs.ErrNotFound,
		},
		{
			name:    "ok",
			id:      id,
			wantErr: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			cd, err := f.Content(ctx, tt.id)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got %v want %v", err, tt.wantErr)
			}

			if tt.wantErr == nil {

				if cd.IsImage != false {
					t.Errorf("content data IsImage mismatch got %v want %v", cd.IsImage, false)
				}

				defer cd.Data.Close()
				b, err := io.ReadAll(cd.Data)
				if err != nil {
					t.Errorf("content data Data read error: %v", err)
				}

				if !bytes.Equal(b, data) {
					t.Errorf("content data bytes mismatch")
				}

			}
		})
	}
}
