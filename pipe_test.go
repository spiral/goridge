package goridge

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPipeReceive(t *testing.T) {
	// ATTENTION, SHOULD BE INITIALIZED
	initLookupTable()

	var err error
	pr, pw := io.Pipe()

	relay := NewPipeRelay(pr, pw)

	nf := NewFrame()
	nf.WriteVersion(VERSION_1)
	nf.WriteFlags(CONTEXT_SEPARATOR, PAYLOAD_CONTROL, PAYLOAD_ERROR)
	nf.WritePayloadLen(uint32(len([]byte(TestPayload))))
	nf.WritePayload([]byte(TestPayload))
	nf.WriteCRC()
	assert.Equal(t, true, nf.VerifyCRC())

	go func(frame *Frame) {
		defer func() {
			_ = pw.Close()
		}()
		err := relay.Send(nf)
		if err != nil {
			t.Fatal(err)
		}
		_ = pw.Close()
	}(nf)

	frame := &Frame{}
	err = relay.Receive(frame)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, frame.ReadVersion(), nf.ReadVersion())
	assert.Equal(t, frame.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, frame.ReadPayloadLen(), nf.ReadPayloadLen())
	assert.Equal(t, true, frame.VerifyCRC())
	assert.Equal(t, []byte(TestPayload), frame.Payload())
}

func TestPipeCRC_Failed(t *testing.T) {
	// ATTENTION, SHOULD BE INITIALIZED
	initLookupTable()

	var err error
	pr, pw := io.Pipe()

	relay := NewPipeRelay(pr, pw)

	nf := NewFrame()
	nf.WriteVersion(VERSION_1)
	nf.WriteFlags(CONTEXT_SEPARATOR)
	nf.WritePayloadLen(uint32(len([]byte(TestPayload))))

	assert.Equal(t, false, nf.VerifyCRC())

	nf.WritePayload([]byte(TestPayload))

	go func(frame *Frame) {
		defer func() {
			_ = pw.Close()
		}()
		err := relay.Send(nf)
		if err != nil {
			t.Fatal(err)
		}
		_ = pw.Close()
	}(nf)

	frame := &Frame{}
	err = relay.Receive(frame)
	assert.Error(t, err)
	assert.Nil(t, frame.header)
	assert.Nil(t, frame.payload)
}
