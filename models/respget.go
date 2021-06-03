package models

type RespGet struct {
	RespMessage  string            `json:"message"`
	RespData     []byte            `json:"data"`
	RespMetadata map[string]string `json:"metadata,omitempty"`
}

func (r RespGet) Message() string {
	return r.RespMessage
}

func (r RespGet) Data() []byte {
	return r.RespData
}

func (r RespGet) Metadata() map[string]string {
	return r.RespMetadata
}

func NewFileGet(msg string, data []byte, metadata map[string]string) RespGet {
	return RespGet{
		RespMessage:  msg,
		RespData:     data,
		RespMetadata: metadata,
	}
}
