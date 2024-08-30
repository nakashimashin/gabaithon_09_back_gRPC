// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.21.12
// source: game.proto

package grpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	MatchService_FindMatch_FullMethodName = "/sample_service.MatchService/FindMatch"
)

// MatchServiceClient is the client API for MatchService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MatchServiceClient interface {
	FindMatch(ctx context.Context, in *MatchRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[MatchResponse], error)
}

type matchServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMatchServiceClient(cc grpc.ClientConnInterface) MatchServiceClient {
	return &matchServiceClient{cc}
}

func (c *matchServiceClient) FindMatch(ctx context.Context, in *MatchRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[MatchResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &MatchService_ServiceDesc.Streams[0], MatchService_FindMatch_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[MatchRequest, MatchResponse]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type MatchService_FindMatchClient = grpc.ServerStreamingClient[MatchResponse]

// MatchServiceServer is the server API for MatchService service.
// All implementations must embed UnimplementedMatchServiceServer
// for forward compatibility.
type MatchServiceServer interface {
	FindMatch(*MatchRequest, grpc.ServerStreamingServer[MatchResponse]) error
	mustEmbedUnimplementedMatchServiceServer()
}

// UnimplementedMatchServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedMatchServiceServer struct{}

func (UnimplementedMatchServiceServer) FindMatch(*MatchRequest, grpc.ServerStreamingServer[MatchResponse]) error {
	return status.Errorf(codes.Unimplemented, "method FindMatch not implemented")
}
func (UnimplementedMatchServiceServer) mustEmbedUnimplementedMatchServiceServer() {}
func (UnimplementedMatchServiceServer) testEmbeddedByValue()                      {}

// UnsafeMatchServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MatchServiceServer will
// result in compilation errors.
type UnsafeMatchServiceServer interface {
	mustEmbedUnimplementedMatchServiceServer()
}

func RegisterMatchServiceServer(s grpc.ServiceRegistrar, srv MatchServiceServer) {
	// If the following call pancis, it indicates UnimplementedMatchServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&MatchService_ServiceDesc, srv)
}

func _MatchService_FindMatch_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(MatchRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(MatchServiceServer).FindMatch(m, &grpc.GenericServerStream[MatchRequest, MatchResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type MatchService_FindMatchServer = grpc.ServerStreamingServer[MatchResponse]

// MatchService_ServiceDesc is the grpc.ServiceDesc for MatchService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MatchService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "sample_service.MatchService",
	HandlerType: (*MatchServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "FindMatch",
			Handler:       _MatchService_FindMatch_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "game.proto",
}
