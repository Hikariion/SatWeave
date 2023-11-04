// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: task_manager.proto

package task_manager

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
	common "satweave/messenger/common"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type RequiredSlotRequest struct {
	SlotDescs            []*common.RequiredSlotDescription `protobuf:"bytes,1,rep,name=slot_descs,json=slotDescs,proto3" json:"slot_descs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                          `json:"-"`
	XXX_unrecognized     []byte                            `json:"-"`
	XXX_sizecache        int32                             `json:"-"`
}

func (m *RequiredSlotRequest) Reset()         { *m = RequiredSlotRequest{} }
func (m *RequiredSlotRequest) String() string { return proto.CompactTextString(m) }
func (*RequiredSlotRequest) ProtoMessage()    {}
func (*RequiredSlotRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_370859c40635d1e2, []int{0}
}
func (m *RequiredSlotRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *RequiredSlotRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RequiredSlotRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *RequiredSlotRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RequiredSlotRequest.Merge(m, src)
}
func (m *RequiredSlotRequest) XXX_Size() int {
	return m.Size()
}
func (m *RequiredSlotRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_RequiredSlotRequest.DiscardUnknown(m)
}

var xxx_messageInfo_RequiredSlotRequest proto.InternalMessageInfo

func (m *RequiredSlotRequest) GetSlotDescs() []*common.RequiredSlotDescription {
	if m != nil {
		return m.SlotDescs
	}
	return nil
}

type RequiredSlotResponse struct {
	Status               *common.Status `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	AvailablePorts       []int64        `protobuf:"varint,2,rep,packed,name=available_ports,json=availablePorts,proto3" json:"available_ports,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *RequiredSlotResponse) Reset()         { *m = RequiredSlotResponse{} }
func (m *RequiredSlotResponse) String() string { return proto.CompactTextString(m) }
func (*RequiredSlotResponse) ProtoMessage()    {}
func (*RequiredSlotResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_370859c40635d1e2, []int{1}
}
func (m *RequiredSlotResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *RequiredSlotResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RequiredSlotResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *RequiredSlotResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RequiredSlotResponse.Merge(m, src)
}
func (m *RequiredSlotResponse) XXX_Size() int {
	return m.Size()
}
func (m *RequiredSlotResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_RequiredSlotResponse.DiscardUnknown(m)
}

var xxx_messageInfo_RequiredSlotResponse proto.InternalMessageInfo

func (m *RequiredSlotResponse) GetStatus() *common.Status {
	if m != nil {
		return m.Status
	}
	return nil
}

func (m *RequiredSlotResponse) GetAvailablePorts() []int64 {
	if m != nil {
		return m.AvailablePorts
	}
	return nil
}

type AvailableWorkersResponse struct {
	Workers              []uint64 `protobuf:"varint,1,rep,packed,name=workers,proto3" json:"workers,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AvailableWorkersResponse) Reset()         { *m = AvailableWorkersResponse{} }
func (m *AvailableWorkersResponse) String() string { return proto.CompactTextString(m) }
func (*AvailableWorkersResponse) ProtoMessage()    {}
func (*AvailableWorkersResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_370859c40635d1e2, []int{2}
}
func (m *AvailableWorkersResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *AvailableWorkersResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_AvailableWorkersResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *AvailableWorkersResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AvailableWorkersResponse.Merge(m, src)
}
func (m *AvailableWorkersResponse) XXX_Size() int {
	return m.Size()
}
func (m *AvailableWorkersResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_AvailableWorkersResponse.DiscardUnknown(m)
}

var xxx_messageInfo_AvailableWorkersResponse proto.InternalMessageInfo

func (m *AvailableWorkersResponse) GetWorkers() []uint64 {
	if m != nil {
		return m.Workers
	}
	return nil
}

type OperatorRequest struct {
	// 对应的算子
	ClsName string `protobuf:"bytes,1,opt,name=cls_name,json=clsName,proto3" json:"cls_name,omitempty"`
	// 流式计算的数据
	Data string `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	// 计算图的 id
	GraphId              int64    `protobuf:"varint,3,opt,name=graph_id,json=graphId,proto3" json:"graph_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *OperatorRequest) Reset()         { *m = OperatorRequest{} }
func (m *OperatorRequest) String() string { return proto.CompactTextString(m) }
func (*OperatorRequest) ProtoMessage()    {}
func (*OperatorRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_370859c40635d1e2, []int{3}
}
func (m *OperatorRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *OperatorRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_OperatorRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *OperatorRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OperatorRequest.Merge(m, src)
}
func (m *OperatorRequest) XXX_Size() int {
	return m.Size()
}
func (m *OperatorRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_OperatorRequest.DiscardUnknown(m)
}

var xxx_messageInfo_OperatorRequest proto.InternalMessageInfo

func (m *OperatorRequest) GetClsName() string {
	if m != nil {
		return m.ClsName
	}
	return ""
}

func (m *OperatorRequest) GetData() string {
	if m != nil {
		return m.Data
	}
	return ""
}

func (m *OperatorRequest) GetGraphId() int64 {
	if m != nil {
		return m.GraphId
	}
	return 0
}

type DeployTaskRequest struct {
	ExecTask             *common.ExecuteTask `protobuf:"bytes,1,opt,name=exec_task,json=execTask,proto3" json:"exec_task,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *DeployTaskRequest) Reset()         { *m = DeployTaskRequest{} }
func (m *DeployTaskRequest) String() string { return proto.CompactTextString(m) }
func (*DeployTaskRequest) ProtoMessage()    {}
func (*DeployTaskRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_370859c40635d1e2, []int{4}
}
func (m *DeployTaskRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *DeployTaskRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_DeployTaskRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *DeployTaskRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeployTaskRequest.Merge(m, src)
}
func (m *DeployTaskRequest) XXX_Size() int {
	return m.Size()
}
func (m *DeployTaskRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_DeployTaskRequest.DiscardUnknown(m)
}

var xxx_messageInfo_DeployTaskRequest proto.InternalMessageInfo

func (m *DeployTaskRequest) GetExecTask() *common.ExecuteTask {
	if m != nil {
		return m.ExecTask
	}
	return nil
}

type StartTaskRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StartTaskRequest) Reset()         { *m = StartTaskRequest{} }
func (m *StartTaskRequest) String() string { return proto.CompactTextString(m) }
func (*StartTaskRequest) ProtoMessage()    {}
func (*StartTaskRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_370859c40635d1e2, []int{5}
}
func (m *StartTaskRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *StartTaskRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_StartTaskRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *StartTaskRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StartTaskRequest.Merge(m, src)
}
func (m *StartTaskRequest) XXX_Size() int {
	return m.Size()
}
func (m *StartTaskRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_StartTaskRequest.DiscardUnknown(m)
}

var xxx_messageInfo_StartTaskRequest proto.InternalMessageInfo

func init() {
	proto.RegisterType((*RequiredSlotRequest)(nil), "messenger.RequiredSlotRequest")
	proto.RegisterType((*RequiredSlotResponse)(nil), "messenger.RequiredSlotResponse")
	proto.RegisterType((*AvailableWorkersResponse)(nil), "messenger.AvailableWorkersResponse")
	proto.RegisterType((*OperatorRequest)(nil), "messenger.OperatorRequest")
	proto.RegisterType((*DeployTaskRequest)(nil), "messenger.DeployTaskRequest")
	proto.RegisterType((*StartTaskRequest)(nil), "messenger.StartTaskRequest")
}

func init() { proto.RegisterFile("task_manager.proto", fileDescriptor_370859c40635d1e2) }

var fileDescriptor_370859c40635d1e2 = []byte{
	// 487 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x93, 0xcd, 0x6e, 0xd3, 0x40,
	0x10, 0xc7, 0xe3, 0xba, 0x6a, 0xea, 0x09, 0x22, 0x74, 0x0b, 0xc8, 0x04, 0x64, 0x22, 0x73, 0x20,
	0x1c, 0x9a, 0xa2, 0x94, 0x17, 0x48, 0x15, 0x04, 0x1c, 0x28, 0x91, 0x8d, 0x04, 0x82, 0x83, 0xb5,
	0xb5, 0x47, 0xc1, 0xc4, 0xf6, 0x9a, 0x9d, 0x4d, 0x5a, 0x5e, 0x82, 0x33, 0x8f, 0xc4, 0x91, 0x47,
	0x40, 0xe1, 0x45, 0xd0, 0x3a, 0x76, 0x70, 0x52, 0x25, 0xb7, 0xdd, 0xff, 0xce, 0xfc, 0xe6, 0x73,
	0x81, 0x29, 0x4e, 0xd3, 0x20, 0xe5, 0x19, 0x9f, 0xa0, 0xec, 0xe7, 0x52, 0x28, 0xc1, 0xac, 0x14,
	0x89, 0x30, 0x9b, 0xa0, 0xec, 0xdc, 0x0a, 0x45, 0x9a, 0x8a, 0x6c, 0xf9, 0xe0, 0x7e, 0x84, 0x63,
	0x0f, 0xbf, 0xcd, 0x62, 0x89, 0x91, 0x9f, 0x08, 0xa5, 0xcf, 0x48, 0x8a, 0x0d, 0x01, 0x28, 0x11,
	0x2a, 0x88, 0x90, 0x42, 0xb2, 0x8d, 0xae, 0xd9, 0x6b, 0x0d, 0xdc, 0xfe, 0x0a, 0xd2, 0xaf, 0xfb,
	0x8c, 0x90, 0x42, 0x19, 0xe7, 0x2a, 0x16, 0x99, 0x67, 0x51, 0x29, 0x90, 0xfb, 0x15, 0xee, 0xae,
	0x93, 0x29, 0x17, 0x19, 0x21, 0x7b, 0x06, 0x07, 0xa4, 0xb8, 0x9a, 0x69, 0xac, 0xd1, 0x6b, 0x0d,
	0x8e, 0x6a, 0x58, 0xbf, 0x78, 0xf0, 0x4a, 0x03, 0xf6, 0x14, 0xda, 0x7c, 0xce, 0xe3, 0x84, 0x5f,
	0x26, 0x18, 0xe4, 0x42, 0x2a, 0xb2, 0xf7, 0xba, 0x66, 0xcf, 0xf4, 0x6e, 0xaf, 0xe4, 0xb1, 0x56,
	0xdd, 0x17, 0x60, 0x0f, 0x2b, 0xe5, 0x83, 0x90, 0x53, 0x94, 0xb4, 0x8a, 0x67, 0x43, 0xf3, 0x6a,
	0x29, 0x15, 0x75, 0xec, 0x7b, 0xd5, 0xd5, 0xfd, 0x0c, 0xed, 0x77, 0x39, 0x4a, 0xae, 0x84, 0xac,
	0xea, 0x7e, 0x00, 0x87, 0x61, 0x42, 0x41, 0xc6, 0x53, 0x2c, 0xd2, 0xb3, 0xbc, 0x66, 0x98, 0xd0,
	0x05, 0x4f, 0x91, 0x31, 0xd8, 0x8f, 0xb8, 0xe2, 0xf6, 0x5e, 0x21, 0x17, 0x67, 0x6d, 0x3e, 0x91,
	0x3c, 0xff, 0x12, 0xc4, 0x91, 0x6d, 0x76, 0x8d, 0x9e, 0xe9, 0x35, 0x8b, 0xfb, 0x9b, 0xc8, 0x7d,
	0x0d, 0x47, 0x23, 0xcc, 0x13, 0xf1, 0xfd, 0x3d, 0xa7, 0x69, 0x85, 0x3f, 0x03, 0x0b, 0xaf, 0x31,
	0x0c, 0xf4, 0x84, 0xca, 0xf2, 0xef, 0xd7, 0xca, 0x7f, 0x79, 0x8d, 0xe1, 0x4c, 0x61, 0xe1, 0x71,
	0xa8, 0x0d, 0xf5, 0xc9, 0x65, 0x70, 0xc7, 0x57, 0x5c, 0xaa, 0x1a, 0x68, 0xf0, 0xc3, 0x04, 0xa6,
	0xef, 0x6f, 0x97, 0x53, 0xf6, 0x51, 0xce, 0xe3, 0x10, 0xd9, 0x08, 0xe0, 0x7f, 0x50, 0xf6, 0xa8,
	0x86, 0xbe, 0x91, 0x4b, 0xa7, 0x1e, 0xf8, 0x22, 0x4e, 0xaa, 0x7e, 0xb9, 0x0d, 0x76, 0x0e, 0xd6,
	0x2a, 0x20, 0x7b, 0xb8, 0x3e, 0x9e, 0xb5, 0x34, 0x76, 0x30, 0x5e, 0x41, 0x7b, 0x2c, 0x45, 0x88,
	0x44, 0x55, 0x8b, 0x59, 0xa7, 0x66, 0xbc, 0xd1, 0xf7, 0x1d, 0x20, 0x1f, 0x8e, 0x87, 0x34, 0xdd,
	0x9c, 0x2e, 0xbb, 0xb7, 0xe9, 0xb0, 0xe4, 0x3c, 0xa9, 0xc9, 0xdb, 0x36, 0xc2, 0x6d, 0xb0, 0x31,
	0xb4, 0x4a, 0x0f, 0xbd, 0x9a, 0xcc, 0xd9, 0xb2, 0xd9, 0x15, 0xf5, 0xf1, 0xd6, 0xf7, 0x8a, 0x78,
	0xfe, 0xfc, 0xd7, 0xc2, 0x31, 0x7e, 0x2f, 0x1c, 0xe3, 0xcf, 0xc2, 0x31, 0x7e, 0xfe, 0x75, 0x1a,
	0x9f, 0x1c, 0xe2, 0xea, 0x0a, 0xf9, 0x1c, 0x4f, 0x89, 0xab, 0x93, 0x4c, 0x44, 0x78, 0xaa, 0xc7,
	0x7e, 0x52, 0x7e, 0xcc, 0xcb, 0x83, 0xe2, 0x03, 0x9e, 0xfd, 0x0b, 0x00, 0x00, 0xff, 0xff, 0x8c,
	0xb7, 0x73, 0x65, 0xaf, 0x03, 0x00, 0x00,
}

func (m *RequiredSlotRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RequiredSlotRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *RequiredSlotRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.SlotDescs) > 0 {
		for iNdEx := len(m.SlotDescs) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.SlotDescs[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintTaskManager(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *RequiredSlotResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RequiredSlotResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *RequiredSlotResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.AvailablePorts) > 0 {
		dAtA2 := make([]byte, len(m.AvailablePorts)*10)
		var j1 int
		for _, num1 := range m.AvailablePorts {
			num := uint64(num1)
			for num >= 1<<7 {
				dAtA2[j1] = uint8(uint64(num)&0x7f | 0x80)
				num >>= 7
				j1++
			}
			dAtA2[j1] = uint8(num)
			j1++
		}
		i -= j1
		copy(dAtA[i:], dAtA2[:j1])
		i = encodeVarintTaskManager(dAtA, i, uint64(j1))
		i--
		dAtA[i] = 0x12
	}
	if m.Status != nil {
		{
			size, err := m.Status.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintTaskManager(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *AvailableWorkersResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *AvailableWorkersResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *AvailableWorkersResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.Workers) > 0 {
		dAtA5 := make([]byte, len(m.Workers)*10)
		var j4 int
		for _, num := range m.Workers {
			for num >= 1<<7 {
				dAtA5[j4] = uint8(uint64(num)&0x7f | 0x80)
				num >>= 7
				j4++
			}
			dAtA5[j4] = uint8(num)
			j4++
		}
		i -= j4
		copy(dAtA[i:], dAtA5[:j4])
		i = encodeVarintTaskManager(dAtA, i, uint64(j4))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *OperatorRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *OperatorRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *OperatorRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if m.GraphId != 0 {
		i = encodeVarintTaskManager(dAtA, i, uint64(m.GraphId))
		i--
		dAtA[i] = 0x18
	}
	if len(m.Data) > 0 {
		i -= len(m.Data)
		copy(dAtA[i:], m.Data)
		i = encodeVarintTaskManager(dAtA, i, uint64(len(m.Data)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.ClsName) > 0 {
		i -= len(m.ClsName)
		copy(dAtA[i:], m.ClsName)
		i = encodeVarintTaskManager(dAtA, i, uint64(len(m.ClsName)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *DeployTaskRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *DeployTaskRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *DeployTaskRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if m.ExecTask != nil {
		{
			size, err := m.ExecTask.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintTaskManager(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *StartTaskRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *StartTaskRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *StartTaskRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	return len(dAtA) - i, nil
}

func encodeVarintTaskManager(dAtA []byte, offset int, v uint64) int {
	offset -= sovTaskManager(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *RequiredSlotRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.SlotDescs) > 0 {
		for _, e := range m.SlotDescs {
			l = e.Size()
			n += 1 + l + sovTaskManager(uint64(l))
		}
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *RequiredSlotResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Status != nil {
		l = m.Status.Size()
		n += 1 + l + sovTaskManager(uint64(l))
	}
	if len(m.AvailablePorts) > 0 {
		l = 0
		for _, e := range m.AvailablePorts {
			l += sovTaskManager(uint64(e))
		}
		n += 1 + sovTaskManager(uint64(l)) + l
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *AvailableWorkersResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Workers) > 0 {
		l = 0
		for _, e := range m.Workers {
			l += sovTaskManager(uint64(e))
		}
		n += 1 + sovTaskManager(uint64(l)) + l
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *OperatorRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ClsName)
	if l > 0 {
		n += 1 + l + sovTaskManager(uint64(l))
	}
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + sovTaskManager(uint64(l))
	}
	if m.GraphId != 0 {
		n += 1 + sovTaskManager(uint64(m.GraphId))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *DeployTaskRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ExecTask != nil {
		l = m.ExecTask.Size()
		n += 1 + l + sovTaskManager(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *StartTaskRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovTaskManager(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTaskManager(x uint64) (n int) {
	return sovTaskManager(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *RequiredSlotRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTaskManager
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: RequiredSlotRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RequiredSlotRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SlotDescs", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskManager
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTaskManager
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTaskManager
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SlotDescs = append(m.SlotDescs, &common.RequiredSlotDescription{})
			if err := m.SlotDescs[len(m.SlotDescs)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTaskManager(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTaskManager
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *RequiredSlotResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTaskManager
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: RequiredSlotResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RequiredSlotResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Status", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskManager
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTaskManager
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTaskManager
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Status == nil {
				m.Status = &common.Status{}
			}
			if err := m.Status.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType == 0 {
				var v int64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowTaskManager
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= int64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.AvailablePorts = append(m.AvailablePorts, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowTaskManager
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					packedLen |= int(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthTaskManager
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthTaskManager
				}
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				var count int
				for _, integer := range dAtA[iNdEx:postIndex] {
					if integer < 128 {
						count++
					}
				}
				elementCount = count
				if elementCount != 0 && len(m.AvailablePorts) == 0 {
					m.AvailablePorts = make([]int64, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v int64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowTaskManager
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= int64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.AvailablePorts = append(m.AvailablePorts, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field AvailablePorts", wireType)
			}
		default:
			iNdEx = preIndex
			skippy, err := skipTaskManager(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTaskManager
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *AvailableWorkersResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTaskManager
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: AvailableWorkersResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: AvailableWorkersResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType == 0 {
				var v uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowTaskManager
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.Workers = append(m.Workers, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowTaskManager
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					packedLen |= int(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthTaskManager
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthTaskManager
				}
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				var count int
				for _, integer := range dAtA[iNdEx:postIndex] {
					if integer < 128 {
						count++
					}
				}
				elementCount = count
				if elementCount != 0 && len(m.Workers) == 0 {
					m.Workers = make([]uint64, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowTaskManager
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.Workers = append(m.Workers, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field Workers", wireType)
			}
		default:
			iNdEx = preIndex
			skippy, err := skipTaskManager(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTaskManager
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *OperatorRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTaskManager
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: OperatorRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: OperatorRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClsName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskManager
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTaskManager
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTaskManager
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ClsName = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskManager
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTaskManager
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTaskManager
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field GraphId", wireType)
			}
			m.GraphId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskManager
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.GraphId |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipTaskManager(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTaskManager
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *DeployTaskRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTaskManager
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: DeployTaskRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: DeployTaskRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ExecTask", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaskManager
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTaskManager
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTaskManager
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.ExecTask == nil {
				m.ExecTask = &common.ExecuteTask{}
			}
			if err := m.ExecTask.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTaskManager(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTaskManager
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *StartTaskRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTaskManager
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: StartTaskRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: StartTaskRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTaskManager(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTaskManager
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipTaskManager(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTaskManager
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTaskManager
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTaskManager
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthTaskManager
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTaskManager
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTaskManager
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTaskManager        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTaskManager          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTaskManager = fmt.Errorf("proto: unexpected end of group")
)
