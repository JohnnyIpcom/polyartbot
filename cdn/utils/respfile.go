package utils

type RespFile interface {
	ID() string
	Size() int
}

type respFile struct {
	FileID   string `json:"fileID"`
	FileSize int    `json:"fileSize"`
}

func (f respFile) ID() string {
	return f.FileID
}

func (f respFile) Size() int {
	return f.FileSize
}

func NewFileResponse(fileID string, size int) RespFile {
	return &respFile{
		FileID:   fileID,
		FileSize: size,
	}
}
