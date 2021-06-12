package models

type RespFile struct {
	FileID   string `json:"fileID"`
	FileSize int    `json:"fileSize"`
}

func (f RespFile) ID() string {
	return f.FileID
}

func (f RespFile) Size() int {
	return f.FileSize
}

func NewFileResponse(fileID string, size int) RespFile {
	return RespFile{
		FileID:   fileID,
		FileSize: size,
	}
}
