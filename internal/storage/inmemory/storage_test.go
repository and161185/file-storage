package inmemory

import (
	"bytes"
	"context"
	"file-storage/internal/errs"
	"file-storage/internal/files"
	"io"
	"reflect"
	"testing"
)

func TestUpsert(t *testing.T) {
	s := New()
	ctx := context.Background()

	table := []struct {
		name    string
		fd      *files.FileData
		wantID  string
		wantErr error
	}{
		{
			name:    "invalid file data",
			fd:      nil,
			wantID:  "",
			wantErr: errs.ErrInvalidFileData,
		},
		{
			name:    "invalid ID",
			fd:      &files.FileData{ID: " "},
			wantID:  "",
			wantErr: errs.ErrInvalidID,
		},
		{
			name:    "upsert",
			fd:      &files.FileData{ID: "1"},
			wantID:  "1",
			wantErr: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			id, err := s.Upsert(ctx, tt.fd)
			if id != tt.wantID {
				t.Errorf("upsert id mismatch got %s want %s", id, tt.wantID)
			}
			if err != tt.wantErr {
				t.Errorf("upsert error mismatch got %v want %v", err, tt.wantErr)
			}
		})
	}

}

func TestInfo(t *testing.T) {

	s := New()
	ctx := context.Background()

	id := "1"
	fd := &files.FileData{ID: id, IsImage: true}

	_, err := s.Upsert(ctx, fd)
	if err != nil {
		t.Fatalf("upsert error: %v", err)
	}

	table := []struct {
		name         string
		id           string
		wantFileInfo *files.FileInfo
		wantErr      error
	}{
		{
			name:         "not found",
			id:           "2",
			wantFileInfo: nil,
			wantErr:      errs.ErrNotFound,
		},
		{
			name:         "ok",
			id:           id,
			wantFileInfo: fileInfo(fd),
			wantErr:      nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			fi, err := s.Info(ctx, tt.id)

			if err != tt.wantErr {
				t.Errorf("got error %v want %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(fi, tt.wantFileInfo) {
				t.Errorf("got file info %v want %v", fi, tt.wantFileInfo)
			}
		})
	}
}

func TestContent(t *testing.T) {

	ctx := context.Background()
	s := New()
	b := []byte("bytes")

	id := "1"
	fd := &files.FileData{ID: id, Data: b}

	s.Upsert(ctx, fd)

	table := []struct {
		name            string
		id              string
		wantContentData *files.ContentData
		wantErr         error
	}{
		{
			name:            "not found",
			id:              "2",
			wantContentData: nil,
			wantErr:         errs.ErrNotFound,
		},
		{
			name:            "ok",
			id:              id,
			wantContentData: &files.ContentData{Data: io.NopCloser(bytes.NewReader(b)), IsImage: false},
			wantErr:         nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {

			cd, err := s.Content(ctx, tt.id)
			if err != tt.wantErr {
				t.Errorf("got error %v want %v", err, tt.wantErr)
			}
			if !equalContentData(cd, tt.wantContentData) {
				t.Errorf("got content data %v want %v", cd, tt.wantContentData)
			}
		})
	}
}

func equalContentData(a *files.ContentData, b *files.ContentData) bool {
	if a == b {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if a.IsImage != b.IsImage {
		return false
	}

	aData, err := io.ReadAll(a.Data)
	if err != nil {
		return false
	}

	bData, err := io.ReadAll(b.Data)
	if err != nil {
		return false
	}

	return bytes.Equal(aData, bData)
}

func TestDelete(t *testing.T) {

	ctx := context.Background()
	s := New()

	id := "1"
	fd := &files.FileData{ID: id}
	s.Upsert(ctx, fd)

	table := []struct {
		name        string
		id          string
		wantInfoErr error
	}{
		{
			name:        "delete existing",
			id:          id,
			wantInfoErr: errs.ErrNotFound,
		},
		{
			name:        "delete not existing",
			id:          "2",
			wantInfoErr: errs.ErrNotFound,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			s.Delete(ctx, tt.id)

			_, err := s.Info(ctx, tt.id)
			if err != tt.wantInfoErr {
				t.Errorf("got info err %v want %v", err, tt.wantInfoErr)
			}

		})
	}

}
