package socket

import (
	"net"
	"testing"

	"github.com/spiral/goridge/v3/pkg/frame"
	"github.com/stretchr/testify/assert"
)

const TestPayload = `alsdjf;lskjdgljasg;lkjsalfkjaskldjflkasjdf;lkasjfdalksdjflkajsdf;lfasdgnslsnblna;sldjjfawlkejr;lwjenlksndlfjawl;ejr;lwjelkrjaldfjl;sdjf`

func TestSocketRelay(t *testing.T) {
	// configure and create tcp4 listener
	ls, err := net.Listen("tcp", "localhost:10002")
	assert.NoError(t, err)

	// TEST FRAME TO SEND
	nf := frame.NewFrame()
	nf.WriteVersion(frame.VERSION_1)
	nf.WriteFlags(frame.CONTROL, frame.CODEC_GOB, frame.CODEC_JSON)
	nf.WritePayloadLen(uint32(len([]byte(TestPayload))))
	nf.WritePayload([]byte(TestPayload))
	nf.WriteCRC()
	assert.Equal(t, true, nf.VerifyCRC())

	conn, err := net.Dial("tcp", "localhost:10002")
	assert.NoError(t, err)
	rsend := NewSocketRelay(conn)
	err = rsend.Send(nf)
	assert.NoError(t, err)

	accept, err := ls.Accept()
	assert.NoError(t, err)
	assert.NotNil(t, accept)

	r := NewSocketRelay(accept)

	fr := &frame.Frame{}
	err = r.Receive(fr)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, fr.ReadVersion(), nf.ReadVersion())
	assert.Equal(t, fr.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, fr.ReadPayloadLen(), nf.ReadPayloadLen())
	assert.Equal(t, true, fr.VerifyCRC())
	assert.Equal(t, []byte(TestPayload), fr.Payload())
}

func TestSocketRelayOptions(t *testing.T) {
	// configure and create tcp4 listener
	ls, err := net.Listen("tcp", "localhost:10001")
	assert.NoError(t, err)

	// TEST FRAME TO SEND
	nf := frame.NewFrame()
	nf.WriteVersion(frame.VERSION_1)
	nf.WriteFlags(frame.CONTROL, frame.CODEC_GOB, frame.CODEC_JSON)
	nf.WritePayloadLen(uint32(len([]byte(TestPayload))))
	nf.WritePayload([]byte(TestPayload))
	nf.WriteOptions(100, 10000, 100000)
	nf.WriteCRC()
	assert.Equal(t, true, nf.VerifyCRC())

	conn, err := net.Dial("tcp", "localhost:10001")
	assert.NoError(t, err)
	rsend := NewSocketRelay(conn)
	err = rsend.Send(nf)
	assert.NoError(t, err)

	accept, err := ls.Accept()
	assert.NoError(t, err)
	assert.NotNil(t, accept)

	r := NewSocketRelay(accept)

	fr := &frame.Frame{}
	err = r.Receive(fr)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, fr.ReadVersion(), nf.ReadVersion())
	assert.Equal(t, fr.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, fr.ReadPayloadLen(), nf.ReadPayloadLen())
	assert.Equal(t, true, fr.VerifyCRC())
	assert.Equal(t, []byte(TestPayload), fr.Payload())
	assert.Equal(t, []uint32{100, 10000, 100000}, fr.ReadOptions())
}

func TestSocketRelayNoPayload(t *testing.T) {
	// configure and create tcp4 listener
	ls, err := net.Listen("tcp", "localhost:12221")
	assert.NoError(t, err)

	// TEST FRAME TO SEND
	nf := frame.NewFrame()
	nf.WriteVersion(frame.VERSION_1)
	nf.WriteFlags(frame.CONTROL, frame.CODEC_GOB, frame.CODEC_JSON)
	nf.WriteOptions(100, 10000, 100000)
	nf.WriteCRC()
	assert.Equal(t, true, nf.VerifyCRC())

	conn, err := net.Dial("tcp", "localhost:12221")
	assert.NoError(t, err)
	rsend := NewSocketRelay(conn)
	err = rsend.Send(nf)
	assert.NoError(t, err)

	accept, err := ls.Accept()
	assert.NoError(t, err)
	assert.NotNil(t, accept)

	r := NewSocketRelay(accept)

	fr := &frame.Frame{}
	err = r.Receive(fr)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, fr.ReadVersion(), nf.ReadVersion())
	assert.Equal(t, fr.ReadFlags(), nf.ReadFlags())
	assert.Equal(t, fr.ReadPayloadLen(), nf.ReadPayloadLen()) // should be zero, without error
	assert.Equal(t, true, fr.VerifyCRC())
	assert.Equal(t, []byte{}, fr.Payload()) // empty
	assert.Equal(t, []uint32{100, 10000, 100000}, fr.ReadOptions())
}

func TestSocketRelayWrongCRC(t *testing.T) {
	// configure and create tcp4 listener
	ls, err := net.Listen("tcp", "localhost:13445")
	assert.NoError(t, err)

	// TEST FRAME TO SEND
	nf := frame.NewFrame()
	nf.WriteVersion(frame.VERSION_1)
	nf.WriteFlags(frame.CONTROL, frame.CODEC_GOB, frame.CODEC_JSON)
	nf.WriteOptions(100, 10000, 100000)
	nf.WriteCRC()
	nf.Header()[6] = 22 // just random wrong CRC directly

	conn, err := net.Dial("tcp", "localhost:13445")
	assert.NoError(t, err)
	_, err = conn.Write(nf.Bytes())
	assert.NoError(t, err)

	accept, err := ls.Accept()
	assert.NoError(t, err)
	assert.NotNil(t, accept)

	r := NewSocketRelay(accept)

	fr := &frame.Frame{}
	err = r.Receive(fr)
	assert.Error(t, err)
	assert.Nil(t, fr.Header())
	assert.Nil(t, fr.Payload())
}
