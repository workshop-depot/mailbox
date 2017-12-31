package mailbox

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var _ Storage = &SliceStorage{}
var _ Mailbox = &mailbox{}

func TestSmoke(t *testing.T) {
	assert := assert.New(t)
	mbox := New(&SliceStorage{})
	go func() { mbox.Send(1) }()
	v, _ := mbox.Receive()
	assert.Equal(1, v)
}

func TestCount(t *testing.T) {
	assert := assert.New(t)
	store := SliceStorage{}
	mbox := New(&store)
	N := 1000
	go func() {
		for i := 0; i < N; i++ {
			mbox.Send(1)
		}
		mbox.Close()
	}()
	total := 0
	for {
		v, ok := mbox.Receive()
		if !ok {
			break
		}
		assert.Equal(1, v)
		total++
	}
	assert.Equal(N, total)
}

func TestItems(t *testing.T) {
	assert := assert.New(t)
	mbox := New(&SliceStorage{})
	N := 1000
	go func() {
		for i := 1; i <= N; i++ {
			i := i
			mbox.Send(i)
		}
		mbox.Close()
	}()
	total := 0
	for {
		v, ok := mbox.Receive()
		if !ok {
			break
		}
		total += v.(int)
	}
	assert.Equal(500500, total)
}

func TestSendClose(t *testing.T) {
	assert := assert.New(t)
	mbox := New(&SliceStorage{})
	assert.True(mbox.Send(1))
	mbox.Close()
	<-time.After(time.Millisecond * 30)
	assert.False(mbox.Send(1, time.Millisecond*10))
}

func TestRcvdClose(t *testing.T) {
	assert := assert.New(t)
	mbox := New(&SliceStorage{})
	assert.True(mbox.Send(1))
	v, ok := mbox.Receive()
	assert.True(ok)
	assert.Equal(1, v)
	mbox.Close()
	<-time.After(time.Millisecond * 30)
	assert.False(mbox.Send(1, time.Millisecond*10))
	v, ok = mbox.Receive()
	assert.False(ok)
	assert.Nil(v)
	v, ok = mbox.Receive(time.Millisecond * 10)
	assert.False(ok)
	assert.Nil(v)
}

func TestRcvdCloseTimeout(t *testing.T) {
	assert := assert.New(t)
	mbox := New(&SliceStorage{})
	assert.True(mbox.Send(1))
	v, ok := mbox.Receive()
	assert.True(ok)
	assert.Equal(1, v)
	v, ok = mbox.Receive(time.Millisecond * 10)
	assert.False(ok)
	assert.Nil(v)
}

func Example() {
	mbox := New(&SliceStorage{})
	mbox.Send("VAL")
	v, _ := mbox.Receive()
	fmt.Println(v)
	// Output: VAL
}
