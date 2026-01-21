package filesystemstorage

import "time"

type FileSystemStorage struct {
	path         string
	lockLifetime time.Time
}

func New()
