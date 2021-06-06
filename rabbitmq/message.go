package rabbitmq

type Publishing struct {
	MessageId   string
	ContentType string
	Body        []byte
}

type Delivery struct {
	MessageId   string
	ContentType string
	Body        []byte
}
