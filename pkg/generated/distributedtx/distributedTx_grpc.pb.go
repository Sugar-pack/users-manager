// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: proto/distributedTx.proto

package distributedtx

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

// DistributedTxServiceClient is the client API for DistributedTxService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DistributedTxServiceClient interface {
	Commit(ctx context.Context, in *TxToCommit, opts ...grpc.CallOption) (*TxResponse, error)
}

type distributedTxServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewDistributedTxServiceClient(cc grpc.ClientConnInterface) DistributedTxServiceClient {
	return &distributedTxServiceClient{cc}
}

func (c *distributedTxServiceClient) Commit(ctx context.Context, in *TxToCommit, opts ...grpc.CallOption) (*TxResponse, error) {
	out := new(TxResponse)
	err := c.cc.Invoke(ctx, "/DistributedTxService/Commit", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DistributedTxServiceServer is the server API for DistributedTxService service.
// All implementations must embed UnimplementedDistributedTxServiceServer
// for forward compatibility
type DistributedTxServiceServer interface {
	Commit(context.Context, *TxToCommit) (*TxResponse, error)
	mustEmbedUnimplementedDistributedTxServiceServer()
}

// UnimplementedDistributedTxServiceServer must be embedded to have forward compatible implementations.
type UnimplementedDistributedTxServiceServer struct {
}

func (UnimplementedDistributedTxServiceServer) Commit(context.Context, *TxToCommit) (*TxResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Commit not implemented")
}
func (UnimplementedDistributedTxServiceServer) mustEmbedUnimplementedDistributedTxServiceServer() {}

// UnsafeDistributedTxServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DistributedTxServiceServer will
// result in compilation errors.
type UnsafeDistributedTxServiceServer interface {
	mustEmbedUnimplementedDistributedTxServiceServer()
}

func RegisterDistributedTxServiceServer(s grpc.ServiceRegistrar, srv DistributedTxServiceServer) {
	s.RegisterService(&DistributedTxService_ServiceDesc, srv)
}

func _DistributedTxService_Commit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TxToCommit)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DistributedTxServiceServer).Commit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/DistributedTxService/Commit",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DistributedTxServiceServer).Commit(ctx, req.(*TxToCommit))
	}
	return interceptor(ctx, in, info, handler)
}

// DistributedTxService_ServiceDesc is the grpc.ServiceDesc for DistributedTxService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DistributedTxService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "DistributedTxService",
	HandlerType: (*DistributedTxServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Commit",
			Handler:    _DistributedTxService_Commit_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/distributedTx.proto",
}