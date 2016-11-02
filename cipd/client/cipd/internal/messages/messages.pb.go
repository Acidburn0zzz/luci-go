// Code generated by protoc-gen-go.
// source: github.com/luci/luci-go/cipd/client/cipd/internal/messages/messages.proto
// DO NOT EDIT!

/*
Package messages is a generated protocol buffer package.

It is generated from these files:
	github.com/luci/luci-go/cipd/client/cipd/internal/messages/messages.proto

It has these top-level messages:
	BlobWithSHA1
	TagCache
	InstanceCache
*/
package messages

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/luci/luci-go/common/proto/google"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// BlobWithSHA1 is a wrapper around a binary blob with SHA1 hash to verify
// its integrity.
type BlobWithSHA1 struct {
	Blob []byte `protobuf:"bytes,1,opt,name=blob,proto3" json:"blob,omitempty"`
	Sha1 []byte `protobuf:"bytes,2,opt,name=sha1,proto3" json:"sha1,omitempty"`
}

func (m *BlobWithSHA1) Reset()                    { *m = BlobWithSHA1{} }
func (m *BlobWithSHA1) String() string            { return proto.CompactTextString(m) }
func (*BlobWithSHA1) ProtoMessage()               {}
func (*BlobWithSHA1) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// TagCache stores a mapping (package name, tag) -> instance ID to speed up
// subsequence ResolveVersion calls when tags are used.
type TagCache struct {
	// Capped list of entries, most recently resolved is last.
	Entries []*TagCache_Entry `protobuf:"bytes,1,rep,name=entries" json:"entries,omitempty"`
}

func (m *TagCache) Reset()                    { *m = TagCache{} }
func (m *TagCache) String() string            { return proto.CompactTextString(m) }
func (*TagCache) ProtoMessage()               {}
func (*TagCache) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *TagCache) GetEntries() []*TagCache_Entry {
	if m != nil {
		return m.Entries
	}
	return nil
}

type TagCache_Entry struct {
	Package    string `protobuf:"bytes,1,opt,name=package" json:"package,omitempty"`
	Tag        string `protobuf:"bytes,2,opt,name=tag" json:"tag,omitempty"`
	InstanceId string `protobuf:"bytes,3,opt,name=instance_id,json=instanceId" json:"instance_id,omitempty"`
}

func (m *TagCache_Entry) Reset()                    { *m = TagCache_Entry{} }
func (m *TagCache_Entry) String() string            { return proto.CompactTextString(m) }
func (*TagCache_Entry) ProtoMessage()               {}
func (*TagCache_Entry) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1, 0} }

// InstanceCache stores a list of instances in cache
// and their last access time.
type InstanceCache struct {
	// Entries is a map of {instance id -> information about instance}.
	Entries map[string]*InstanceCache_Entry `protobuf:"bytes,1,rep,name=entries" json:"entries,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	// LastSynced is timestamp when we synchronized Entries with actual
	// instance files.
	LastSynced *google_protobuf.Timestamp `protobuf:"bytes,2,opt,name=last_synced,json=lastSynced" json:"last_synced,omitempty"`
}

func (m *InstanceCache) Reset()                    { *m = InstanceCache{} }
func (m *InstanceCache) String() string            { return proto.CompactTextString(m) }
func (*InstanceCache) ProtoMessage()               {}
func (*InstanceCache) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *InstanceCache) GetEntries() map[string]*InstanceCache_Entry {
	if m != nil {
		return m.Entries
	}
	return nil
}

func (m *InstanceCache) GetLastSynced() *google_protobuf.Timestamp {
	if m != nil {
		return m.LastSynced
	}
	return nil
}

// Entry stores info about an instance.
type InstanceCache_Entry struct {
	// LastAccess is last time this instance was retrieved from or put to the
	// cache.
	LastAccess *google_protobuf.Timestamp `protobuf:"bytes,2,opt,name=last_access,json=lastAccess" json:"last_access,omitempty"`
}

func (m *InstanceCache_Entry) Reset()                    { *m = InstanceCache_Entry{} }
func (m *InstanceCache_Entry) String() string            { return proto.CompactTextString(m) }
func (*InstanceCache_Entry) ProtoMessage()               {}
func (*InstanceCache_Entry) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2, 0} }

func (m *InstanceCache_Entry) GetLastAccess() *google_protobuf.Timestamp {
	if m != nil {
		return m.LastAccess
	}
	return nil
}

func init() {
	proto.RegisterType((*BlobWithSHA1)(nil), "messages.BlobWithSHA1")
	proto.RegisterType((*TagCache)(nil), "messages.TagCache")
	proto.RegisterType((*TagCache_Entry)(nil), "messages.TagCache.Entry")
	proto.RegisterType((*InstanceCache)(nil), "messages.InstanceCache")
	proto.RegisterType((*InstanceCache_Entry)(nil), "messages.InstanceCache.Entry")
}

func init() {
	proto.RegisterFile("github.com/luci/luci-go/cipd/client/cipd/internal/messages/messages.proto", fileDescriptor0)
}

var fileDescriptor0 = []byte{
	// 360 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x8c, 0x90, 0xdf, 0x4a, 0xc3, 0x30,
	0x14, 0x87, 0xe9, 0xe6, 0xdc, 0x76, 0x3a, 0x41, 0x72, 0x55, 0x0a, 0x32, 0x19, 0x5e, 0xec, 0xc6,
	0x96, 0x6d, 0x20, 0xa2, 0x20, 0xcc, 0x3f, 0xe0, 0x6e, 0xbb, 0x81, 0x78, 0x35, 0xd2, 0x34, 0x76,
	0x61, 0x59, 0x53, 0x96, 0x54, 0xe8, 0x7b, 0xf8, 0x1a, 0xbe, 0xa3, 0x69, 0xba, 0x4c, 0xe7, 0x85,
	0x78, 0x53, 0x7e, 0xe7, 0xe4, 0x6b, 0xce, 0x97, 0x03, 0xb3, 0x94, 0xa9, 0x55, 0x11, 0x07, 0x44,
	0x6c, 0x42, 0x5e, 0x10, 0x66, 0x3e, 0x97, 0xa9, 0x08, 0x09, 0xcb, 0x93, 0x90, 0x70, 0x46, 0x33,
	0x55, 0x67, 0x96, 0x29, 0xba, 0xcd, 0x30, 0x0f, 0x37, 0x54, 0x4a, 0x9c, 0x52, 0xb9, 0x0f, 0x41,
	0xbe, 0x15, 0x4a, 0xa0, 0x8e, 0xad, 0xfd, 0x7e, 0x2a, 0x44, 0xca, 0x69, 0x68, 0xfa, 0x71, 0xf1,
	0x16, 0x2a, 0xa6, 0xcf, 0x14, 0xde, 0xe4, 0x35, 0x3a, 0xb8, 0x82, 0xde, 0x3d, 0x17, 0xf1, 0x8b,
	0x9e, 0x3d, 0x7f, 0x9e, 0x8e, 0x10, 0x82, 0xa3, 0x58, 0xd7, 0x9e, 0x73, 0xee, 0x0c, 0x7b, 0x91,
	0xc9, 0x55, 0x4f, 0xae, 0xf0, 0xc8, 0x6b, 0xd4, 0xbd, 0x2a, 0x0f, 0x3e, 0x1c, 0xe8, 0x2c, 0x70,
	0xfa, 0x80, 0xc9, 0x8a, 0xa2, 0x31, 0xb4, 0xb5, 0xdc, 0x96, 0x51, 0xa9, 0xff, 0x6b, 0x0e, 0xdd,
	0xb1, 0x17, 0xec, 0x8d, 0x2c, 0x14, 0x3c, 0x69, 0xa2, 0x8c, 0x2c, 0xe8, 0x2f, 0xa0, 0x65, 0x3a,
	0xc8, 0x83, 0x76, 0x8e, 0xc9, 0x5a, 0xc3, 0x66, 0x68, 0x37, 0xb2, 0x25, 0x3a, 0x85, 0xa6, 0xc2,
	0xa9, 0x19, 0xdb, 0x8d, 0xaa, 0x88, 0xfa, 0xe0, 0xb2, 0x4c, 0xeb, 0x67, 0x84, 0x2e, 0x59, 0xe2,
	0x35, 0xcd, 0x09, 0xd8, 0xd6, 0x2c, 0x19, 0x7c, 0x36, 0xe0, 0x64, 0xb6, 0x2b, 0x6b, 0xb7, 0xbb,
	0xdf, 0x6e, 0x17, 0xdf, 0x6e, 0x07, 0xa4, 0x11, 0xd4, 0xd8, 0xa1, 0x27, 0xba, 0x05, 0x97, 0x63,
	0xa9, 0x96, 0xb2, 0xd4, 0x60, 0x62, 0x64, 0xdc, 0xb1, 0x1f, 0xd4, 0x7b, 0x0d, 0xec, 0x5e, 0x83,
	0x85, 0xdd, 0x6b, 0x04, 0x15, 0x3e, 0x37, 0xb4, 0xff, 0x68, 0x1f, 0x69, 0x6f, 0xc1, 0x84, 0xe8,
	0xe1, 0xff, 0xbd, 0x65, 0x6a, 0x68, 0xff, 0x15, 0x7a, 0x3f, 0xdd, 0xaa, 0xbd, 0xac, 0x69, 0xb9,
	0xdb, 0x56, 0x15, 0xd1, 0x04, 0x5a, 0xef, 0x98, 0x17, 0x74, 0x77, 0xf1, 0xd9, 0x5f, 0x4f, 0x2c,
	0xa3, 0x9a, 0xbd, 0x69, 0x5c, 0x3b, 0xf1, 0xb1, 0x19, 0x3d, 0xf9, 0x0a, 0x00, 0x00, 0xff, 0xff,
	0xe8, 0xbb, 0x67, 0xc8, 0x7d, 0x02, 0x00, 0x00,
}