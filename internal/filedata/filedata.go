package filedata

import (
	"file-storage/internal/imgproc"
	"io"
	"maps"
	"time"
)

// UploadCommand contains input required to create a new file or update an existing one.
type UploadCommand struct {
	ID       string
	Data     []byte
	Hash     string
	Public   bool
	IsImage  bool
	Metadata map[string]any
}

// ContentCommand describes a content read request, including optional image transformation parameters.
type ContentCommand struct {
	ID     string
	Width  *int
	Height *int
	Format *string
}

// FileData contains file bytes together with system metadata used by business logic and storage.
type FileData struct {
	ID         string
	Data       []byte
	HashSource string
	HashStored string
	Public     bool
	FileSize   int
	IsImage    bool
	Format     imgproc.ImgFormat
	Width      int
	Height     int
	Metadata   map[string]any
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// FileInfo contains file metadata without file content.
type FileInfo struct {
	ID         string            `json:"id"`
	HashSource string            `json:"hash_source"`
	HashStored string            `json:"hash_stored"`
	Public     bool              `json:"public"`
	FileSize   int               `json:"file_size"`
	IsImage    bool              `json:"is_image"`
	Format     imgproc.ImgFormat `json:"format"`
	Width      int               `json:"width"`
	Height     int               `json:"height"`
	Metadata   map[string]any    `json:"metadata"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// ContentData contains a file content stream and metadata required to build an HTTP response.
type ContentData struct {
	Data    io.ReadCloser
	IsImage bool
}

// ImageInfo describes detected or stored image format and dimensions.
type ImageInfo struct {
	Format imgproc.ImgFormat
	Width  int
	Height int
}

// FileInfoFromFileData builds FileInfo from FileData by copying metadata fields and omitting file content.
func FileInfoFromFileData(fd *FileData) *FileInfo {
	fi := FileInfo{
		ID:         fd.ID,
		HashSource: fd.HashSource,
		HashStored: fd.HashStored,
		Public:     fd.Public,
		FileSize:   fd.FileSize,
		IsImage:    fd.IsImage,
		Format:     fd.Format,
		Width:      fd.Width,
		Height:     fd.Height,
		CreatedAt:  fd.CreatedAt,
		UpdatedAt:  fd.UpdatedAt,
	}

	if fd.Metadata != nil {
		metadata := make(map[string]any, len(fd.Metadata))
		maps.Copy(metadata, fd.Metadata)
		fi.Metadata = metadata
	}

	return &fi
}
