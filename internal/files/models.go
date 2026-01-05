package files

import "time"

type FileData struct {
	ID        string
	Data      []byte
	Hash      string
	IsImage   bool
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FileInfo struct {
	ID        string
	Hash      string
	IsImage   bool
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}
