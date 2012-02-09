// Code generated by protoc-gen-go from "protoc-gen-go/testdata/service.proto"
// DO NOT EDIT!

package svc

import proto "code.google.com/p/goprotobuf/proto"
import "math"

import "net"
import "net/rpc"
import "github.com/kylelemons/go-rpcgen/codec"

// Reference proto and math imports to suppress error if they are not otherwise used.
var _ = proto.GetString
var _ = math.Inf

type Args struct {
	A                *string `protobuf:"bytes,1,req,name=a" json:"a,omitempty"`
	B                *string `protobuf:"bytes,2,req,name=b" json:"b,omitempty"`
	XXX_unrecognized []byte  `json:",omitempty"`
}

func (this *Args) Reset()         { *this = Args{} }
func (this *Args) String() string { return proto.CompactTextString(this) }

type Return struct {
	C                *string `protobuf:"bytes,1,req,name=c" json:"c,omitempty"`
	XXX_unrecognized []byte  `json:",omitempty"`
}

func (this *Return) Reset()         { *this = Return{} }
func (this *Return) String() string { return proto.CompactTextString(this) }

func init() {
}

// ConcatService is an interface satisfied by the generated client and
// which must be implemented by the object wrapped by the server.
type ConcatService interface {
	Concat(in *Args, out *Return) error
}

// internal wrapper for type-safe RPC calling
type rpcConcatServiceClient struct {
	*rpc.Client
}

func (this rpcConcatServiceClient) Concat(in *Args, out *Return) error {
	return this.Call("ConcatService.Concat", in, out)
}

// NewConcatServiceClient returns an *rpc.Client wrapper for calling the methods of
// ConcatService remotely.
func NewConcatServiceClient(conn net.Conn) ConcatService {
	return rpcConcatServiceClient{rpc.NewClientWithCodec(services.NewClientCodec(conn))}
}

// ServeConcatService serves the given ConcatService backend implementation on conn.
func ServeConcatService(conn net.Conn, backend ConcatService) error {
	srv := rpc.NewServer()
	if err := srv.RegisterName("ConcatService", backend); err != nil {
		return err
	}
	srv.ServeCodec(services.NewServerCodec(conn))
	return nil
}

// DialConcatService returns a ConcatService for calling the ConcatService servince at addr (TCP).
func DialConcatService(addr string) (ConcatService, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return NewConcatServiceClient(conn), nil
}

// ListenAndServeConcatService serves the given ConcatService backend implementation
// on all connections accepted as a result of listening on addr (TCP).
func ListenAndServeConcatService(addr string, backend ConcatService) error {
	clients, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	srv := rpc.NewServer()
	if err := srv.RegisterName("ConcatService", backend); err != nil {
		return err
	}
	for {
		conn, err := clients.Accept()
		if err != nil {
			return err
		}
		go srv.ServeCodec(services.NewServerCodec(conn))
	}
	panic("unreachable")
}
