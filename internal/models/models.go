package models

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

type UploadRequest struct {
	ID       string         `json:"id"`
	Data     []byte         `json:"data"`
	Hash     string         `json:"hash"`
	IsImage  *bool          `json:"is_image"`
	Metadata map[string]any `json:"metadata"`
}
