package codec

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/rpc"

	"code.google.com/p/goprotobuf/proto"
	"github.com/kylelemons/go-rpcgen/plugin/wire"
)

// ServerCodec implements the rpc.ServerCodec interface for generic protobufs.
// The same implementation works for all protobufs because it defers the
// decoding of a protocol buffer to the proto package and it uses a set header
// that is the same regardless of the protobuf being used for the RPC.
type ServerCodec struct {
	r *bufio.Reader
	w io.WriteCloser
}

type ProtoReader interface {
	io.Reader
	io.ByteReader
}

// ReadProto reads a uvarint size and then a protobuf from r.
// If the size read is zero, nothing more is read.
func ReadProto(r ProtoReader, pb interface{}) error {
	size, err := binary.ReadUvarint(r)
	if err != nil {
		return err
	}
	// TODO max size?
	buf := make([]byte, size)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}
	return proto.Unmarshal(buf, pb)
}

// WriteProto writes a uvarint size and then a protobuf to w.
// If the data takes no space (like rpc.InvalidRequest),
// only a zero size is written.
func WriteProto(w io.Writer, pb interface{}) error {
	// Allocate enough space for the biggest uvarint
	var size [binary.MaxVarintLen64]byte

	// Marshal the protobuf
	data, err := proto.Marshal(pb)
	if err != nil {
		return err
	}

	// Write the size and data
	n := binary.PutUvarint(size[:], uint64(len(data)))
	if _, err = w.Write(size[:n]); err != nil {
		return err
	}
	if _, err = w.Write(data); err != nil {
		return err
	}
	return nil
}

// NewServerCodec returns a ServerCodec that communicates with the ClientCodec
// on the other end of the given conn.
func NewServerCodec(conn net.Conn) *ServerCodec {
	return &ServerCodec{bufio.NewReader(conn), conn}
}

// ReadRequestHeader reads the header protobuf (which is prefixed by a uvarint
// indicating its size) from the connection, decodes it, and stores the fields
// in the given request.
func (s *ServerCodec) ReadRequestHeader(req *rpc.Request) error {
	var header wire.Header
	if err := ReadProto(s.r, &header); err != nil {
		return err
	}
	if header.Method == nil {
		return fmt.Errorf("header missing method: %s", header)
	}
	if header.Seq == nil {
		return fmt.Errorf("header missing seq: %s", header)
	}
	req.ServiceMethod = *header.Method
	req.Seq = *header.Seq
	return nil
}

// ReadRequestBody reads a uvarint from the connection and decodes that many
// subsequent bytes into the given protobuf (which should be a pointer to a
// struct that is generated by the proto package).
func (s *ServerCodec) ReadRequestBody(pb interface{}) error {
	return ReadProto(s.r, pb)
}

// WriteResponse writes the appropriate header protobuf and the given protobuf
// to the connection (each prefixed with a uvarint indicating its size).  If
// the response was invalid, the size of the body of the resp is reported as
// having size zero and is not sent.
func (s *ServerCodec) WriteResponse(resp *rpc.Response, pb interface{}) error {
	// Write the header
	header := wire.Header{
		Method: &resp.ServiceMethod,
		Seq:    &resp.Seq,
	}
	if resp.Error != "" {
		header.Error = &resp.Error
	}
	if err := WriteProto(s.w, &header); err != nil {
		return nil
	}

	// Write the proto
	return WriteProto(s.w, pb)
}

// Close closes the underlying conneciton.
func (s *ServerCodec) Close() error {
	return s.w.Close()
}

// ClientCodec implements the rpc.ClientCodec interface for generic protobufs.
// The same implementation works for all protobufs because it defers the
// encoding of a protocol buffer to the proto package and it uses a set header
// that is the same regardless of the protobuf being used for the RPC.
type ClientCodec struct {
	r *bufio.Reader
	w io.WriteCloser
}

// NewClientCodec returns a ClientCodec for communicating with the ServerCodec
// on the other end of the conn.
func NewClientCodec(conn net.Conn) *ClientCodec {
	return &ClientCodec{bufio.NewReader(conn), conn}
}

// WriteRequest writes the appropriate header protobuf and the given protobuf
// to the connection (each prefixed with a uvarint indicating its size).
func (c *ClientCodec) WriteRequest(req *rpc.Request, pb interface{}) error {
	// Write the header
	header := wire.Header{
		Method: &req.ServiceMethod,
		Seq:    &req.Seq,
	}
	if err := WriteProto(c.w, &header); err != nil {
		return err
	}

	return WriteProto(c.w, pb)
}

// ReadResponseHeader reads the header protobuf (which is prefixed by a uvarint
// indicating its size) from the connection, decodes it, and stores the fields
// in the given request.
func (c *ClientCodec) ReadResponseHeader(resp *rpc.Response) error {
	var header wire.Header
	if err := ReadProto(c.r, &header); err != nil {
		return err
	}
	if header.Method == nil {
		return fmt.Errorf("header missing method: %s", header)
	}
	if header.Seq == nil {
		return fmt.Errorf("header missing seq: %s", header)
	}
	resp.ServiceMethod = *header.Method
	resp.Seq = *header.Seq
	if header.Error != nil {
		resp.Error = *header.Error
	}
	return nil
}

// ReadResponseBody reads a uvarint from the connection and decodes that many
// subsequent bytes into the given protobuf (which should be a pointer to a
// struct that is generated by the proto package).  If the uvarint size read
// is zero, nothing is done (this indicates an error condition, which was
// encapsulated in the header)
func (c *ClientCodec) ReadResponseBody(pb interface{}) error {
	return ReadProto(c.r, pb)
}

// Close closes the underlying connection.
func (c *ClientCodec) Close() error {
	return c.w.Close()
}

// BUG: The server/client don't do a sanity check on the size of the proto
// before reading it, so it's possible to maliciously instruct the
// client/server to allocate too much memory.
