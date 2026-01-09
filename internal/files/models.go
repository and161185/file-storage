package files

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
	ID        string
	FileSize  int
	IsImage   bool
	Format    imgproc.ImgFormat
	Width     int
	Height    int
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
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
