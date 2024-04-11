// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.6.1
// source: watcher.proto

package watcher

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	common "satweave/messenger/common"
	infos "satweave/sat-node/infos"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Watcher_AddNewNodeToCluster_FullMethodName      = "/messenger.Watcher/AddNewNodeToCluster"
	Watcher_GetClusterInfo_FullMethodName           = "/messenger.Watcher/GetClusterInfo"
	Watcher_SubmitGeoUnSensitiveTask_FullMethodName = "/messenger.Watcher/SubmitGeoUnSensitiveTask"
)

// WatcherClient is the client API for Watcher service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type WatcherClient interface {
	// AddNewNodeToCluster will propose a new NodeInfo in moon,
	// if success, it will propose a ConfChang, to add the raftNode into moon group
	AddNewNodeToCluster(ctx context.Context, in *infos.NodeInfo, opts ...grpc.CallOption) (*AddNodeReply, error)
	// GetClusterInfo return requested cluster info to rpc client,
	// if GetClusterInfoRequest.Term == 0, it will return current cluster info.
	GetClusterInfo(ctx context.Context, in *GetClusterInfoRequest, opts ...grpc.CallOption) (*GetClusterInfoReply, error)
	SubmitGeoUnSensitiveTask(ctx context.Context, in *GeoUnSensitiveTaskRequest, opts ...grpc.CallOption) (*common.Result, error)
}

type watcherClient struct {
	cc grpc.ClientConnInterface
}

func NewWatcherClient(cc grpc.ClientConnInterface) WatcherClient {
	return &watcherClient{cc}
}

func (c *watcherClient) AddNewNodeToCluster(ctx context.Context, in *infos.NodeInfo, opts ...grpc.CallOption) (*AddNodeReply, error) {
	out := new(AddNodeReply)
	err := c.cc.Invoke(ctx, Watcher_AddNewNodeToCluster_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *watcherClient) GetClusterInfo(ctx context.Context, in *GetClusterInfoRequest, opts ...grpc.CallOption) (*GetClusterInfoReply, error) {
	out := new(GetClusterInfoReply)
	err := c.cc.Invoke(ctx, Watcher_GetClusterInfo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *watcherClient) SubmitGeoUnSensitiveTask(ctx context.Context, in *GeoUnSensitiveTaskRequest, opts ...grpc.CallOption) (*common.Result, error) {
	out := new(common.Result)
	err := c.cc.Invoke(ctx, Watcher_SubmitGeoUnSensitiveTask_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WatcherServer is the server API for Watcher service.
// All implementations must embed UnimplementedWatcherServer
// for forward compatibility
type WatcherServer interface {
	// AddNewNodeToCluster will propose a new NodeInfo in moon,
	// if success, it will propose a ConfChang, to add the raftNode into moon group
	AddNewNodeToCluster(context.Context, *infos.NodeInfo) (*AddNodeReply, error)
	// GetClusterInfo return requested cluster info to rpc client,
	// if GetClusterInfoRequest.Term == 0, it will return current cluster info.
	GetClusterInfo(context.Context, *GetClusterInfoRequest) (*GetClusterInfoReply, error)
	SubmitGeoUnSensitiveTask(context.Context, *GeoUnSensitiveTaskRequest) (*common.Result, error)
	mustEmbedUnimplementedWatcherServer()
}

// UnimplementedWatcherServer must be embedded to have forward compatible implementations.
type UnimplementedWatcherServer struct {
}

func (UnimplementedWatcherServer) AddNewNodeToCluster(context.Context, *infos.NodeInfo) (*AddNodeReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddNewNodeToCluster not implemented")
}
func (UnimplementedWatcherServer) GetClusterInfo(context.Context, *GetClusterInfoRequest) (*GetClusterInfoReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetClusterInfo not implemented")
}
func (UnimplementedWatcherServer) SubmitGeoUnSensitiveTask(context.Context, *GeoUnSensitiveTaskRequest) (*common.Result, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitGeoUnSensitiveTask not implemented")
}
func (UnimplementedWatcherServer) mustEmbedUnimplementedWatcherServer() {}

// UnsafeWatcherServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WatcherServer will
// result in compilation errors.
type UnsafeWatcherServer interface {
	mustEmbedUnimplementedWatcherServer()
}

func RegisterWatcherServer(s grpc.ServiceRegistrar, srv WatcherServer) {
	s.RegisterService(&Watcher_ServiceDesc, srv)
}

func _Watcher_AddNewNodeToCluster_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(infos.NodeInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WatcherServer).AddNewNodeToCluster(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Watcher_AddNewNodeToCluster_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WatcherServer).AddNewNodeToCluster(ctx, req.(*infos.NodeInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _Watcher_GetClusterInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetClusterInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WatcherServer).GetClusterInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Watcher_GetClusterInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WatcherServer).GetClusterInfo(ctx, req.(*GetClusterInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Watcher_SubmitGeoUnSensitiveTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GeoUnSensitiveTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WatcherServer).SubmitGeoUnSensitiveTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Watcher_SubmitGeoUnSensitiveTask_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WatcherServer).SubmitGeoUnSensitiveTask(ctx, req.(*GeoUnSensitiveTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Watcher_ServiceDesc is the grpc.ServiceDesc for Watcher service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Watcher_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "messenger.Watcher",
	HandlerType: (*WatcherServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddNewNodeToCluster",
			Handler:    _Watcher_AddNewNodeToCluster_Handler,
		},
		{
			MethodName: "GetClusterInfo",
			Handler:    _Watcher_GetClusterInfo_Handler,
		},
		{
			MethodName: "SubmitGeoUnSensitiveTask",
			Handler:    _Watcher_SubmitGeoUnSensitiveTask_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "watcher.proto",
}
