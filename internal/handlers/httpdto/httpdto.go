package httpdto

type UploadRequest struct {
	ID       string         `json:"id"`
	Data     []byte         `json:"data"`
	Hash     string         `json:"hash"`
	Public   bool           `json:"bool"`
	IsImage  *bool          `json:"is_image"`
	Metadata map[string]any `json:"metadata"`
}

type ContentRequest struct {
	ID     string
	Width  *int
	Height *int
	Format *string
}
