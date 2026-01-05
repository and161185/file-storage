package files

type Storage interface {
	Save(fd *FileData) (string, error)
	Info(ID string) (*FileInfo, error)
	Content(ID string) ([]byte, error)
	Delete(ID string) error
}
