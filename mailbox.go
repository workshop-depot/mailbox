// Package mailbox provides unbound mailboxes
package mailbox

import (
	"time"
)

// Storage interface must be implemented for any back storage for a mailbox.
// The storage can be bound.
type Storage interface {
	Len() int
	Peek() interface{}
	Drop()
	Append(interface{})
}

// SliceStorage is a slice which implements the Storage interface
type SliceStorage []interface{}

// Len len of the storage
func (store SliceStorage) Len() int { return len(store) }

// Append appends value to the end of the storage
func (store *SliceStorage) Append(v interface{}) { *store = append(*store, v) }

// Peek peeks one element from the head of the storage without removing it
func (store SliceStorage) Peek() interface{} { return store[0] }

// Drop drops the head element of the storage
func (store *SliceStorage) Drop() { *store = (*store)[1:] }

// Mailbox interface is an unbound mailbox
type Mailbox interface {
	Send(interface{}, ...time.Duration) bool
	Receive(...time.Duration) (interface{}, bool)
	Close() error
}

type mailbox struct {
	close   chan struct{}
	send    chan interface{}
	receive chan interface{}
	mails   Storage
}

func (mb *mailbox) Send(v interface{}, timeout ...time.Duration) bool {
	var toch <-chan time.Time
	if len(timeout) > 0 && timeout[0] > 0 {
		toch = time.After(timeout[0])
	}
	select {
	case <-toch:
		return false
	case mb.send <- v:
	}
	return true
}

func (mb *mailbox) Close() error { close(mb.close); return nil }

func (mb *mailbox) Receive(timeout ...time.Duration) (interface{}, bool) {
	var toch <-chan time.Time
	if len(timeout) > 0 && timeout[0] > 0 {
		toch = time.After(timeout[0])
	}
	select {
	case <-toch:
		return nil, false
	case v, ok := <-mb.receive:
		return v, ok
	}
}

func (mb *mailbox) loop() {
	defer close(mb.receive)
	var actualReceive chan interface{}
	var actualSend = mb.send
	first := func() interface{} {
		// this check is needed because (f* semantics)
		// select statement first evaluates the arguments of it's case clauses
		// then will ignore them if the channel is nil.
		// I've got some "not constructive" responses on golang-nuts group, here:
		// https://groups.google.com/forum/#!topic/golang-nuts/uKllRM89qb0
		// :\ whatever ...
		if mb.mails.Len() == 0 {
			return nil
		}
		return mb.mails.Peek()
	}
	for {
		select {
		case <-mb.close:
			actualSend = nil
			if mb.mails.Len() > 0 {
				continue
			}
			return
		case v := <-actualSend:
			mb.mails.Append(v)
			actualReceive = mb.receive
		case actualReceive <- first():
			mb.mails.Drop()
			if mb.mails.Len() == 0 {
				actualReceive = nil
			}
		}
	}
}

// New .
func New(store Storage) Mailbox {
	res := &mailbox{
		close:   make(chan struct{}),
		send:    make(chan interface{}),
		receive: make(chan interface{}),
		mails:   store,
	}
	go res.loop()
	return res
}
