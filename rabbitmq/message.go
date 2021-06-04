package rabbitmq

import "github.com/streadway/amqp"

type Publishing struct {
	MessageId   string
	ContentType string
	Body        []byte
}

type Delivery struct {
	tag         uint64
	ack         amqp.Acknowledger
	MessageId   string
	ContentType string
	Body        []byte
}

func (d Delivery) Ack(multiple bool) error {
	return d.ack.Ack(d.tag, multiple)
}

func (d Delivery) Reject(requeue bool) error {
	return d.ack.Reject(d.tag, requeue)
}

func (d Delivery) Nack(multiple, requeue bool) error {
	return d.ack.Nack(d.tag, multiple, requeue)
}
