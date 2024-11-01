// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.2
// source: game.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	GameService_Play_FullMethodName = "/game.GameService/Play"
)

// GameServiceClient is the client API for GameService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GameServiceClient interface {
	Play(ctx context.Context, opts ...grpc.CallOption) (GameService_PlayClient, error)
}

type gameServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewGameServiceClient(cc grpc.ClientConnInterface) GameServiceClient {
	return &gameServiceClient{cc}
}

func (c *gameServiceClient) Play(ctx context.Context, opts ...grpc.CallOption) (GameService_PlayClient, error) {
	stream, err := c.cc.NewStream(ctx, &GameService_ServiceDesc.Streams[0], GameService_Play_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &gameServicePlayClient{stream}
	return x, nil
}

type GameService_PlayClient interface {
	Send(*PlayRequest) error
	Recv() (*PlayResponse, error)
	grpc.ClientStream
}

type gameServicePlayClient struct {
	grpc.ClientStream
}

func (x *gameServicePlayClient) Send(m *PlayRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *gameServicePlayClient) Recv() (*PlayResponse, error) {
	m := new(PlayResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// GameServiceServer is the server API for GameService service.
// All implementations must embed UnimplementedGameServiceServer
// for forward compatibility
type GameServiceServer interface {
	Play(GameService_PlayServer) error
	mustEmbedUnimplementedGameServiceServer()
}

// UnimplementedGameServiceServer must be embedded to have forward compatible implementations.
type UnimplementedGameServiceServer struct {
}

func (UnimplementedGameServiceServer) Play(GameService_PlayServer) error {
	return status.Errorf(codes.Unimplemented, "method Play not implemented")
}
func (UnimplementedGameServiceServer) mustEmbedUnimplementedGameServiceServer() {}

// UnsafeGameServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GameServiceServer will
// result in compilation errors.
type UnsafeGameServiceServer interface {
	mustEmbedUnimplementedGameServiceServer()
}

func RegisterGameServiceServer(s grpc.ServiceRegistrar, srv GameServiceServer) {
	s.RegisterService(&GameService_ServiceDesc, srv)
}

func _GameService_Play_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(GameServiceServer).Play(&gameServicePlayServer{stream})
}

type GameService_PlayServer interface {
	Send(*PlayResponse) error
	Recv() (*PlayRequest, error)
	grpc.ServerStream
}

type gameServicePlayServer struct {
	grpc.ServerStream
}

func (x *gameServicePlayServer) Send(m *PlayResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *gameServicePlayServer) Recv() (*PlayRequest, error) {
	m := new(PlayRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// GameService_ServiceDesc is the grpc.ServiceDesc for GameService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GameService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "game.GameService",
	HandlerType: (*GameServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Play",
			Handler:       _GameService_Play_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "game.proto",
}
