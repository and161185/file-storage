package files

import (
	"file-storage/internal/imgproc"
	"time"
)

type UploadCommand struct {
	ID       string
	Data     []byte
	Hash     string
	IsImage  bool
	Metadata map[string]any
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

type ImageInfo struct {
	Format imgproc.ImgFormat
	Width  int
	Height int
}
