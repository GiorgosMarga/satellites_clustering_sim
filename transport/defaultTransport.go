package transport

import (
	"context"
	"fmt"
	"time"
)

type Default struct {
	internalChan chan *Message
	externalChan chan any
	peers        map[int]chan *Message
}

func NewDefault() *Default {
	return &Default{
		internalChan: make(chan *Message, 30),
		externalChan: make(chan any, 30),
		peers:        make(map[int]chan *Message),
	}
}

func (d *Default) Start(ctx context.Context) {
	for {
		select {
		case msg := <-d.internalChan:
			d.externalChan <- msg.Payload
		case <-ctx.Done():
			// fmt.Println("Shutting down transport...")
			return
		}
	}
}

func (d *Default) Send(msg *Message) error {
	if _, ok := d.peers[msg.To]; !ok {
		return fmt.Errorf("peer %d not found: %w", msg.To, ErrPeerNotFound)
	}
	select {
	case d.peers[msg.To] <- msg:
	case <-time.Tick(500 * time.Millisecond):
		return fmt.Errorf("peer %d timeout: %w", msg.To, ErrTimeout)
	}
	return nil
}

func (d *Default) Consume() <-chan any {
	return d.externalChan
}
func (d *Default) AddPeer(id int, ch chan *Message) {
	d.peers[id] = ch
}
func (d *Default) Chan() chan *Message {
	return d.internalChan
}
