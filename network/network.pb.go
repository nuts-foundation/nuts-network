// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.6.1
// source: network.proto

package network

import (
	context "context"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

// We multiplex all of our messages over a single super-message type, because we're using streams. If we did not do that,
// we'd need to open a stream per operation, which would make the number of required streams explode (since we're
// building a full mesh network).
type NetworkMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Header        *Header        `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	AdvertHash    *AdvertHash    `protobuf:"bytes,100,opt,name=advertHash,proto3" json:"advertHash,omitempty"`
	HashListQuery *HashListQuery `protobuf:"bytes,101,opt,name=hashListQuery,proto3" json:"hashListQuery,omitempty"`
	HashList      *HashList      `protobuf:"bytes,102,opt,name=hashList,proto3" json:"hashList,omitempty"`
	DocumentQuery *DocumentQuery `protobuf:"bytes,103,opt,name=documentQuery,proto3" json:"documentQuery,omitempty"`
	Document      *Document      `protobuf:"bytes,104,opt,name=document,proto3" json:"document,omitempty"`
}

func (x *NetworkMessage) Reset() {
	*x = NetworkMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_network_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NetworkMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NetworkMessage) ProtoMessage() {}

func (x *NetworkMessage) ProtoReflect() protoreflect.Message {
	mi := &file_network_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NetworkMessage.ProtoReflect.Descriptor instead.
func (*NetworkMessage) Descriptor() ([]byte, []int) {
	return file_network_proto_rawDescGZIP(), []int{0}
}

func (x *NetworkMessage) GetHeader() *Header {
	if x != nil {
		return x.Header
	}
	return nil
}

func (x *NetworkMessage) GetAdvertHash() *AdvertHash {
	if x != nil {
		return x.AdvertHash
	}
	return nil
}

func (x *NetworkMessage) GetHashListQuery() *HashListQuery {
	if x != nil {
		return x.HashListQuery
	}
	return nil
}

func (x *NetworkMessage) GetHashList() *HashList {
	if x != nil {
		return x.HashList
	}
	return nil
}

func (x *NetworkMessage) GetDocumentQuery() *DocumentQuery {
	if x != nil {
		return x.DocumentQuery
	}
	return nil
}

func (x *NetworkMessage) GetDocument() *Document {
	if x != nil {
		return x.Document
	}
	return nil
}

// Metadata types
type Header struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Version uint32 `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
}

func (x *Header) Reset() {
	*x = Header{}
	if protoimpl.UnsafeEnabled {
		mi := &file_network_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Header) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Header) ProtoMessage() {}

func (x *Header) ProtoReflect() protoreflect.Message {
	mi := &file_network_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Header.ProtoReflect.Descriptor instead.
func (*Header) Descriptor() ([]byte, []int) {
	return file_network_proto_rawDescGZIP(), []int{1}
}

func (x *Header) GetVersion() uint32 {
	if x != nil {
		return x.Version
	}
	return 0
}

// Actual messages go here
type AdvertHash struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Hash []byte `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
}

func (x *AdvertHash) Reset() {
	*x = AdvertHash{}
	if protoimpl.UnsafeEnabled {
		mi := &file_network_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AdvertHash) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AdvertHash) ProtoMessage() {}

func (x *AdvertHash) ProtoReflect() protoreflect.Message {
	mi := &file_network_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AdvertHash.ProtoReflect.Descriptor instead.
func (*AdvertHash) Descriptor() ([]byte, []int) {
	return file_network_proto_rawDescGZIP(), []int{2}
}

func (x *AdvertHash) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

// Message to ask for a peer's hash list
type HashListQuery struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *HashListQuery) Reset() {
	*x = HashListQuery{}
	if protoimpl.UnsafeEnabled {
		mi := &file_network_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HashListQuery) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HashListQuery) ProtoMessage() {}

func (x *HashListQuery) ProtoReflect() protoreflect.Message {
	mi := &file_network_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HashListQuery.ProtoReflect.Descriptor instead.
func (*HashListQuery) Descriptor() ([]byte, []int) {
	return file_network_proto_rawDescGZIP(), []int{3}
}

// Message to inform a peer of our hash list
type HashList struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Hashes []*HashInTime `protobuf:"bytes,1,rep,name=hashes,proto3" json:"hashes,omitempty"`
}

func (x *HashList) Reset() {
	*x = HashList{}
	if protoimpl.UnsafeEnabled {
		mi := &file_network_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HashList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HashList) ProtoMessage() {}

func (x *HashList) ProtoReflect() protoreflect.Message {
	mi := &file_network_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HashList.ProtoReflect.Descriptor instead.
func (*HashList) Descriptor() ([]byte, []int) {
	return file_network_proto_rawDescGZIP(), []int{4}
}

func (x *HashList) GetHashes() []*HashInTime {
	if x != nil {
		return x.Hashes
	}
	return nil
}

type HashInTime struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Time int64  `protobuf:"varint,1,opt,name=time,proto3" json:"time,omitempty"`
	Hash []byte `protobuf:"bytes,2,opt,name=hash,proto3" json:"hash,omitempty"`
}

func (x *HashInTime) Reset() {
	*x = HashInTime{}
	if protoimpl.UnsafeEnabled {
		mi := &file_network_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HashInTime) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HashInTime) ProtoMessage() {}

func (x *HashInTime) ProtoReflect() protoreflect.Message {
	mi := &file_network_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HashInTime.ProtoReflect.Descriptor instead.
func (*HashInTime) Descriptor() ([]byte, []int) {
	return file_network_proto_rawDescGZIP(), []int{5}
}

func (x *HashInTime) GetTime() int64 {
	if x != nil {
		return x.Time
	}
	return 0
}

func (x *HashInTime) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

// Message to ask for a peer for a specific document
type DocumentQuery struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Hash []byte `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
}

func (x *DocumentQuery) Reset() {
	*x = DocumentQuery{}
	if protoimpl.UnsafeEnabled {
		mi := &file_network_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DocumentQuery) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DocumentQuery) ProtoMessage() {}

func (x *DocumentQuery) ProtoReflect() protoreflect.Message {
	mi := &file_network_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DocumentQuery.ProtoReflect.Descriptor instead.
func (*DocumentQuery) Descriptor() ([]byte, []int) {
	return file_network_proto_rawDescGZIP(), []int{6}
}

func (x *DocumentQuery) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

type Document struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Time     int64  `protobuf:"varint,1,opt,name=time,proto3" json:"time,omitempty"`
	Contents []byte `protobuf:"bytes,2,opt,name=contents,proto3" json:"contents,omitempty"`
}

func (x *Document) Reset() {
	*x = Document{}
	if protoimpl.UnsafeEnabled {
		mi := &file_network_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Document) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Document) ProtoMessage() {}

func (x *Document) ProtoReflect() protoreflect.Message {
	mi := &file_network_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Document.ProtoReflect.Descriptor instead.
func (*Document) Descriptor() ([]byte, []int) {
	return file_network_proto_rawDescGZIP(), []int{7}
}

func (x *Document) GetTime() int64 {
	if x != nil {
		return x.Time
	}
	return 0
}

func (x *Document) GetContents() []byte {
	if x != nil {
		return x.Contents
	}
	return nil
}

var File_network_proto protoreflect.FileDescriptor

var file_network_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x07, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x22, 0xc8, 0x02, 0x0a, 0x0e, 0x4e, 0x65, 0x74,
	0x77, 0x6f, 0x72, 0x6b, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x27, 0x0a, 0x06, 0x68,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x6e, 0x65,
	0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2e, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x52, 0x06, 0x68, 0x65,
	0x61, 0x64, 0x65, 0x72, 0x12, 0x33, 0x0a, 0x0a, 0x61, 0x64, 0x76, 0x65, 0x72, 0x74, 0x48, 0x61,
	0x73, 0x68, 0x18, 0x64, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f,
	0x72, 0x6b, 0x2e, 0x41, 0x64, 0x76, 0x65, 0x72, 0x74, 0x48, 0x61, 0x73, 0x68, 0x52, 0x0a, 0x61,
	0x64, 0x76, 0x65, 0x72, 0x74, 0x48, 0x61, 0x73, 0x68, 0x12, 0x3c, 0x0a, 0x0d, 0x68, 0x61, 0x73,
	0x68, 0x4c, 0x69, 0x73, 0x74, 0x51, 0x75, 0x65, 0x72, 0x79, 0x18, 0x65, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x16, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2e, 0x48, 0x61, 0x73, 0x68, 0x4c,
	0x69, 0x73, 0x74, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x0d, 0x68, 0x61, 0x73, 0x68, 0x4c, 0x69,
	0x73, 0x74, 0x51, 0x75, 0x65, 0x72, 0x79, 0x12, 0x2d, 0x0a, 0x08, 0x68, 0x61, 0x73, 0x68, 0x4c,
	0x69, 0x73, 0x74, 0x18, 0x66, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x6e, 0x65, 0x74, 0x77,
	0x6f, 0x72, 0x6b, 0x2e, 0x48, 0x61, 0x73, 0x68, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x08, 0x68, 0x61,
	0x73, 0x68, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x3c, 0x0a, 0x0d, 0x64, 0x6f, 0x63, 0x75, 0x6d, 0x65,
	0x6e, 0x74, 0x51, 0x75, 0x65, 0x72, 0x79, 0x18, 0x67, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e,
	0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2e, 0x44, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74,
	0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x0d, 0x64, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x51,
	0x75, 0x65, 0x72, 0x79, 0x12, 0x2d, 0x0a, 0x08, 0x64, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74,
	0x18, 0x68, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b,
	0x2e, 0x44, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x08, 0x64, 0x6f, 0x63, 0x75, 0x6d,
	0x65, 0x6e, 0x74, 0x22, 0x22, 0x0a, 0x06, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x18, 0x0a,
	0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07,
	0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x20, 0x0a, 0x0a, 0x41, 0x64, 0x76, 0x65, 0x72,
	0x74, 0x48, 0x61, 0x73, 0x68, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x04, 0x68, 0x61, 0x73, 0x68, 0x22, 0x0f, 0x0a, 0x0d, 0x48, 0x61, 0x73,
	0x68, 0x4c, 0x69, 0x73, 0x74, 0x51, 0x75, 0x65, 0x72, 0x79, 0x22, 0x37, 0x0a, 0x08, 0x48, 0x61,
	0x73, 0x68, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x2b, 0x0a, 0x06, 0x68, 0x61, 0x73, 0x68, 0x65, 0x73,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b,
	0x2e, 0x48, 0x61, 0x73, 0x68, 0x49, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x52, 0x06, 0x68, 0x61, 0x73,
	0x68, 0x65, 0x73, 0x22, 0x34, 0x0a, 0x0a, 0x48, 0x61, 0x73, 0x68, 0x49, 0x6e, 0x54, 0x69, 0x6d,
	0x65, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x04, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x04, 0x68, 0x61, 0x73, 0x68, 0x22, 0x23, 0x0a, 0x0d, 0x44, 0x6f, 0x63,
	0x75, 0x6d, 0x65, 0x6e, 0x74, 0x51, 0x75, 0x65, 0x72, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x61,
	0x73, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x68, 0x61, 0x73, 0x68, 0x22, 0x3a,
	0x0a, 0x08, 0x44, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x69,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x1a,
	0x0a, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x32, 0x4c, 0x0a, 0x07, 0x4e, 0x65,
	0x74, 0x77, 0x6f, 0x72, 0x6b, 0x12, 0x41, 0x0a, 0x07, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74,
	0x12, 0x17, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2e, 0x4e, 0x65, 0x74, 0x77, 0x6f,
	0x72, 0x6b, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x1a, 0x17, 0x2e, 0x6e, 0x65, 0x74, 0x77,
	0x6f, 0x72, 0x6b, 0x2e, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x4d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x22, 0x00, 0x28, 0x01, 0x30, 0x01, 0x42, 0x31, 0x5a, 0x2f, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6e, 0x75, 0x74, 0x73, 0x2d, 0x66, 0x6f, 0x75, 0x6e,
	0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6e, 0x75, 0x74, 0x73, 0x2d, 0x6e, 0x65, 0x74, 0x77,
	0x6f, 0x72, 0x6b, 0x2f, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_network_proto_rawDescOnce sync.Once
	file_network_proto_rawDescData = file_network_proto_rawDesc
)

func file_network_proto_rawDescGZIP() []byte {
	file_network_proto_rawDescOnce.Do(func() {
		file_network_proto_rawDescData = protoimpl.X.CompressGZIP(file_network_proto_rawDescData)
	})
	return file_network_proto_rawDescData
}

var file_network_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_network_proto_goTypes = []interface{}{
	(*NetworkMessage)(nil), // 0: network.NetworkMessage
	(*Header)(nil),         // 1: network.Header
	(*AdvertHash)(nil),     // 2: network.AdvertHash
	(*HashListQuery)(nil),  // 3: network.HashListQuery
	(*HashList)(nil),       // 4: network.HashList
	(*HashInTime)(nil),     // 5: network.HashInTime
	(*DocumentQuery)(nil),  // 6: network.DocumentQuery
	(*Document)(nil),       // 7: network.Document
}
var file_network_proto_depIdxs = []int32{
	1, // 0: network.NetworkMessage.header:type_name -> network.Header
	2, // 1: network.NetworkMessage.advertHash:type_name -> network.AdvertHash
	3, // 2: network.NetworkMessage.hashListQuery:type_name -> network.HashListQuery
	4, // 3: network.NetworkMessage.hashList:type_name -> network.HashList
	6, // 4: network.NetworkMessage.documentQuery:type_name -> network.DocumentQuery
	7, // 5: network.NetworkMessage.document:type_name -> network.Document
	5, // 6: network.HashList.hashes:type_name -> network.HashInTime
	0, // 7: network.Network.Connect:input_type -> network.NetworkMessage
	0, // 8: network.Network.Connect:output_type -> network.NetworkMessage
	8, // [8:9] is the sub-list for method output_type
	7, // [7:8] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_network_proto_init() }
func file_network_proto_init() {
	if File_network_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_network_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NetworkMessage); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_network_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Header); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_network_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AdvertHash); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_network_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HashListQuery); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_network_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HashList); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_network_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HashInTime); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_network_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DocumentQuery); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_network_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Document); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_network_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_network_proto_goTypes,
		DependencyIndexes: file_network_proto_depIdxs,
		MessageInfos:      file_network_proto_msgTypes,
	}.Build()
	File_network_proto = out.File
	file_network_proto_rawDesc = nil
	file_network_proto_goTypes = nil
	file_network_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// NetworkClient is the client API for Network service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type NetworkClient interface {
	Connect(ctx context.Context, opts ...grpc.CallOption) (Network_ConnectClient, error)
}

type networkClient struct {
	cc grpc.ClientConnInterface
}

func NewNetworkClient(cc grpc.ClientConnInterface) NetworkClient {
	return &networkClient{cc}
}

func (c *networkClient) Connect(ctx context.Context, opts ...grpc.CallOption) (Network_ConnectClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Network_serviceDesc.Streams[0], "/network.Network/Connect", opts...)
	if err != nil {
		return nil, err
	}
	x := &networkConnectClient{stream}
	return x, nil
}

type Network_ConnectClient interface {
	Send(*NetworkMessage) error
	Recv() (*NetworkMessage, error)
	grpc.ClientStream
}

type networkConnectClient struct {
	grpc.ClientStream
}

func (x *networkConnectClient) Send(m *NetworkMessage) error {
	return x.ClientStream.SendMsg(m)
}

func (x *networkConnectClient) Recv() (*NetworkMessage, error) {
	m := new(NetworkMessage)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// NetworkServer is the server API for Network service.
type NetworkServer interface {
	Connect(Network_ConnectServer) error
}

// UnimplementedNetworkServer can be embedded to have forward compatible implementations.
type UnimplementedNetworkServer struct {
}

func (*UnimplementedNetworkServer) Connect(Network_ConnectServer) error {
	return status.Errorf(codes.Unimplemented, "method Connect not implemented")
}

func RegisterNetworkServer(s *grpc.Server, srv NetworkServer) {
	s.RegisterService(&_Network_serviceDesc, srv)
}

func _Network_Connect_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(NetworkServer).Connect(&networkConnectServer{stream})
}

type Network_ConnectServer interface {
	Send(*NetworkMessage) error
	Recv() (*NetworkMessage, error)
	grpc.ServerStream
}

type networkConnectServer struct {
	grpc.ServerStream
}

func (x *networkConnectServer) Send(m *NetworkMessage) error {
	return x.ServerStream.SendMsg(m)
}

func (x *networkConnectServer) Recv() (*NetworkMessage, error) {
	m := new(NetworkMessage)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _Network_serviceDesc = grpc.ServiceDesc{
	ServiceName: "network.Network",
	HandlerType: (*NetworkServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Connect",
			Handler:       _Network_Connect_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "network.proto",
}
