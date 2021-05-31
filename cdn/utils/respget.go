package utils

type RespGet interface {
	Message() string
	Data() []byte
	Metadata() map[string]string
}

type respGet struct {
	RespMessage  string            `json:"message"`
	RespData     []byte            `json:"data"`
	RespMetadata map[string]string `json:"metadata,omitempty"`
}

func (r respGet) Message() string {
	return r.RespMessage
}

func (r respGet) Data() []byte {
	return r.RespData
}

func (r respGet) Metadata() map[string]string {
	return r.RespMetadata
}

func NewFileGet(msg string, data []byte, metadata map[string]string) RespGet {
	return respGet{
		RespMessage:  msg,
		RespData:     data,
		RespMetadata: metadata,
	}
}
