package utils

type RespFile interface {
	Name() string
	Size() int
}

type respFile struct {
	FileName string `json:"file_name"`
	FileSize int    `json:"file_size"`
}

func (f respFile) Name() string {
	return f.FileName
}

func (f respFile) Size() int {
	return f.FileSize
}

func NewFileResponse(name string, size int) RespFile {
	return &respFile{
		FileName: name,
		FileSize: size,
	}
}
