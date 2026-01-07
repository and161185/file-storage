package files

import "time"

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
	MimeType  string
	FileSize  int
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FileInfo struct {
	ID        string
	MimeType  string
	FileSize  int
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}
