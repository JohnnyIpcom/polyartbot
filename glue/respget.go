package glue

type RespGet interface {
	Message() string
	Data() []byte
}

type respGet struct {
	RespMessage string `json:"message"`
	RespData    []byte `json:"data"`
}

func (r respGet) Message() string {
	return r.RespMessage
}

func (r respGet) Data() []byte {
	return r.RespData
}

func NewFileGet(msg string, data []byte) RespGet {
	return respGet{
		RespMessage: msg,
		RespData:    data,
	}
}
