package messaging

import "context"

type Message struct {
	Body          []byte
	MessageID     string
	ReceiptHandle string
}

type Queue interface {
	Send(ctx context.Context, body []byte) error
	Receive(ctx context.Context, maxMessages int) ([]Message, error)
	Delete(ctx context.Context, receiptHandle string) error
}
