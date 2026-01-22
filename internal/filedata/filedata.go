package filedata

import (
	"file-storage/internal/imgproc"
	"io"
	"time"
)

type UploadCommand struct {
	ID       string
	Data     []byte
	Hash     string
	IsImage  bool
	Metadata map[string]any
}

type ContentCommand struct {
	ID     string
	Width  *int
	Height *int
	Format *string
}

type FileData struct {
	ID        string
	Data      []byte
	Hash      string
	FileSize  int
	IsImage   bool
	Format    imgproc.ImgFormat
	Width     int
	Height    int
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FileInfo struct {
	ID        string            `json:"id"`
	Hash      string            `json:"hash"`
	FileSize  int               `json:"file_size"`
	IsImage   bool              `json:"is_image"`
	Format    imgproc.ImgFormat `json:"format"`
	Width     int               `json:"width"`
	Height    int               `json:"height"`
	Metadata  map[string]any    `json:"metadata"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type ContentData struct {
	Data    io.ReadCloser
	IsImage bool
}

type ImageInfo struct {
	Format imgproc.ImgFormat
	Width  int
	Height int
}

func FileInfoFromFileData(fd *FileData) *FileInfo {
	fi := FileInfo{
		ID:        fd.ID,
		FileSize:  fd.FileSize,
		IsImage:   fd.IsImage,
		Format:    fd.Format,
		Width:     fd.Width,
		Height:    fd.Height,
		CreatedAt: fd.CreatedAt,
		UpdatedAt: fd.UpdatedAt,
	}

	if fd.Metadata != nil {
		metadata := make(map[string]any, len(fd.Metadata))
		for k, v := range fd.Metadata {
			metadata[k] = v
		}
		fi.Metadata = metadata
	}

	return &fi
}
