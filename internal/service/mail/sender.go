package mail

import "context"

type Message struct {
	ToEmail string
	ToName  string
	Subject string
	Text    string
}

type Sender interface {
	Ready() bool
	Send(ctx context.Context, msg Message) error
}

type noopSender struct {
	ready bool
}

func NewNoopSender(ready bool) Sender {
	return noopSender{ready: ready}
}

func (s noopSender) Ready() bool {
	return s.ready
}

func (s noopSender) Send(ctx context.Context, msg Message) error {
	return nil
}
