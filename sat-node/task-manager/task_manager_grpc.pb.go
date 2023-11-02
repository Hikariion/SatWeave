// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.12.4
// source: task_manager.proto

package task_manager

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	common "satweave/messenger/common"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// TaskManagerServiceClient is the client API for TaskManagerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TaskManagerServiceClient interface {
	// Deploy a task to the task manager.
	DeployTask(ctx context.Context, in *DeployTaskRequest, opts ...grpc.CallOption) (*common.NilResponse, error)
	StartTask(ctx context.Context, in *StartTaskRequest, opts ...grpc.CallOption) (*common.NilResponse, error)
}

type taskManagerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTaskManagerServiceClient(cc grpc.ClientConnInterface) TaskManagerServiceClient {
	return &taskManagerServiceClient{cc}
}

func (c *taskManagerServiceClient) DeployTask(ctx context.Context, in *DeployTaskRequest, opts ...grpc.CallOption) (*common.NilResponse, error) {
	out := new(common.NilResponse)
	err := c.cc.Invoke(ctx, "/messenger.TaskManagerService/DeployTask", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskManagerServiceClient) StartTask(ctx context.Context, in *StartTaskRequest, opts ...grpc.CallOption) (*common.NilResponse, error) {
	out := new(common.NilResponse)
	err := c.cc.Invoke(ctx, "/messenger.TaskManagerService/StartTask", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TaskManagerServiceServer is the server API for TaskManagerService service.
// All implementations must embed UnimplementedTaskManagerServiceServer
// for forward compatibility
type TaskManagerServiceServer interface {
	// Deploy a task to the task manager.
	DeployTask(context.Context, *DeployTaskRequest) (*common.NilResponse, error)
	StartTask(context.Context, *StartTaskRequest) (*common.NilResponse, error)
	mustEmbedUnimplementedTaskManagerServiceServer()
}

// UnimplementedTaskManagerServiceServer must be embedded to have forward compatible implementations.
type UnimplementedTaskManagerServiceServer struct {
}

func (UnimplementedTaskManagerServiceServer) DeployTask(context.Context, *DeployTaskRequest) (*common.NilResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeployTask not implemented")
}
func (UnimplementedTaskManagerServiceServer) StartTask(context.Context, *StartTaskRequest) (*common.NilResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartTask not implemented")
}
func (UnimplementedTaskManagerServiceServer) mustEmbedUnimplementedTaskManagerServiceServer() {}

// UnsafeTaskManagerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TaskManagerServiceServer will
// result in compilation errors.
type UnsafeTaskManagerServiceServer interface {
	mustEmbedUnimplementedTaskManagerServiceServer()
}

func RegisterTaskManagerServiceServer(s grpc.ServiceRegistrar, srv TaskManagerServiceServer) {
	s.RegisterService(&TaskManagerService_ServiceDesc, srv)
}

func _TaskManagerService_DeployTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeployTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskManagerServiceServer).DeployTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/messenger.TaskManagerService/DeployTask",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskManagerServiceServer).DeployTask(ctx, req.(*DeployTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaskManagerService_StartTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskManagerServiceServer).StartTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/messenger.TaskManagerService/StartTask",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskManagerServiceServer).StartTask(ctx, req.(*StartTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// TaskManagerService_ServiceDesc is the grpc.ServiceDesc for TaskManagerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TaskManagerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "messenger.TaskManagerService",
	HandlerType: (*TaskManagerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "DeployTask",
			Handler:    _TaskManagerService_DeployTask_Handler,
		},
		{
			MethodName: "StartTask",
			Handler:    _TaskManagerService_StartTask_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "task_manager.proto",
}
