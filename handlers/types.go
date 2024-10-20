package handlers

type FileMetadata struct {
	FileID   string `json:"file_id"`
	Filename string `json:"filename"`
	IsImage  bool   `json:"is_image"`
}

type UploadRequest struct {
	Metadata FileMetadata `json:"metadata"`
	Data     string       `json:"data"`
}

type UploadResponse struct {
	FileID string `json:"file_id"`
}

type DownloadResponse struct {
	Metadata FileMetadata `json:"metadata"`
	Data     string       `json:"data"`
}
