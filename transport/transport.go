package transport

import (
	"context"
	"errors"
)

var (
	ErrPeerNotFound = errors.New("peer not found")
	ErrTimeout      = errors.New("timeout")
)

type Transport interface {
	Start(context.Context)
	Send(*Message) error
	Consume() <-chan any
	AddPeer(int, chan *Message)
	Chan() chan *Message
}

type Message struct {
	From    int
	To      int
	Payload any
}
