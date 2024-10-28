package models

// config file
type Mongo struct {
	URI                string `json:"uri"`
	Name               string `json:"name"`
	Collection         string `json:"collection"`
	MaxPoolSize        int    `json:"max_pool_size"`
	MinPoolSize        int    `json:"min_pool_size"`
	MaxConnIdleTimeSec int    `json:"max_conn_idle_time_sec"`
}

type Application struct {
	Port int `json:"port"`
}

type Features struct {
	Test bool `json:"test"`
}

type Tokens struct {
	GeneralToken  string `json:"general_token"`
	DownloadToken string `json:"download_token"`
}

type Config struct {
	Mongo       Mongo       `json:"database"`
	Application Application `json:"server"`
	Features    Features    `json:"features"`
	Tokens      Tokens      `json:"tokens"`
}

// requests
type UploadRequest struct {
	Metadata map[string]interface{} `json:"metadata"`
	Data     string                 `json:"data"`
}

type UploadResponse struct {
	FileID string `json:"file_id"`
}

type DownloadResponse struct {
	Metadata map[string]interface{} `json:"metadata"`
	Data     string                 `json:"data"`
}

type QueryParams struct {
	Ext      string `schema:"ext"`
	Width    int    `schema:"width"`
	Heigth   int    `schema:"height"`
	FileOnly bool   `schema:"fileOnly"`
}
