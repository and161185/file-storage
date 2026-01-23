package filesystemstorage

import (
	"context"
	"file-storage/internal/config"
	"file-storage/internal/contextkeys"
	"file-storage/internal/filedata"
	"file-storage/internal/logger"
	"testing"
	"time"
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
			path:    "./temp",
			fd:      nil,
			wantErr: true,
			wantID:  "",
		},
		{
			name:    "invalid filedata 2",
			path:    "./temp",
			fd:      &filedata.FileData{},
			wantErr: true,
			wantID:  "",
		},
		{
			name:    "invalid path",
			path:    "\x00",
			fd:      &filedata.FileData{ID: id, Data: data, Hash: "123"},
			wantErr: true,
			wantID:  "",
		},
		{
			name:    "ok",
			path:    "./temp",
			fd:      &filedata.FileData{ID: id, Data: data, Hash: "123", IsImage: false},
			wantErr: false,
			wantID:  id,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.FileSystem{Path: tt.path, LockLifetime: 1 * time.Second}
			f := New(&cfg)
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
		})
	}

}
