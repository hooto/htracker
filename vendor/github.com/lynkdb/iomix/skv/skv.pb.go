// Code generated by protoc-gen-go. DO NOT EDIT.
// source: skv.proto

package skv

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type KvMeta struct {
	Type                 uint32   `protobuf:"varint,1,opt,name=type" json:"type,omitempty"`
	Version              uint64   `protobuf:"varint,2,opt,name=version" json:"version,omitempty"`
	Expired              uint64   `protobuf:"varint,3,opt,name=expired" json:"expired,omitempty"`
	Created              uint64   `protobuf:"varint,4,opt,name=created" json:"created,omitempty"`
	Updated              uint64   `protobuf:"varint,5,opt,name=updated" json:"updated,omitempty"`
	Size                 uint64   `protobuf:"varint,6,opt,name=size" json:"size,omitempty"`
	Sum                  uint32   `protobuf:"varint,7,opt,name=sum" json:"sum,omitempty"`
	Num                  uint64   `protobuf:"varint,8,opt,name=num" json:"num,omitempty"`
	Name                 string   `protobuf:"bytes,13,opt,name=name" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KvMeta) Reset()         { *m = KvMeta{} }
func (m *KvMeta) String() string { return proto.CompactTextString(m) }
func (*KvMeta) ProtoMessage()    {}
func (*KvMeta) Descriptor() ([]byte, []int) {
	return fileDescriptor_skv_f8d55fd67c11843a, []int{0}
}
func (m *KvMeta) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KvMeta.Unmarshal(m, b)
}
func (m *KvMeta) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KvMeta.Marshal(b, m, deterministic)
}
func (dst *KvMeta) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KvMeta.Merge(dst, src)
}
func (m *KvMeta) XXX_Size() int {
	return xxx_messageInfo_KvMeta.Size(m)
}
func (m *KvMeta) XXX_DiscardUnknown() {
	xxx_messageInfo_KvMeta.DiscardUnknown(m)
}

var xxx_messageInfo_KvMeta proto.InternalMessageInfo

func (m *KvMeta) GetType() uint32 {
	if m != nil {
		return m.Type
	}
	return 0
}

func (m *KvMeta) GetVersion() uint64 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *KvMeta) GetExpired() uint64 {
	if m != nil {
		return m.Expired
	}
	return 0
}

func (m *KvMeta) GetCreated() uint64 {
	if m != nil {
		return m.Created
	}
	return 0
}

func (m *KvMeta) GetUpdated() uint64 {
	if m != nil {
		return m.Updated
	}
	return 0
}

func (m *KvMeta) GetSize() uint64 {
	if m != nil {
		return m.Size
	}
	return 0
}

func (m *KvMeta) GetSum() uint32 {
	if m != nil {
		return m.Sum
	}
	return 0
}

func (m *KvMeta) GetNum() uint64 {
	if m != nil {
		return m.Num
	}
	return 0
}

func (m *KvMeta) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type KvEntry struct {
	Meta                 *KvMeta  `protobuf:"bytes,1,opt,name=meta" json:"meta,omitempty"`
	Value                []byte   `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KvEntry) Reset()         { *m = KvEntry{} }
func (m *KvEntry) String() string { return proto.CompactTextString(m) }
func (*KvEntry) ProtoMessage()    {}
func (*KvEntry) Descriptor() ([]byte, []int) {
	return fileDescriptor_skv_f8d55fd67c11843a, []int{1}
}
func (m *KvEntry) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KvEntry.Unmarshal(m, b)
}
func (m *KvEntry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KvEntry.Marshal(b, m, deterministic)
}
func (dst *KvEntry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KvEntry.Merge(dst, src)
}
func (m *KvEntry) XXX_Size() int {
	return xxx_messageInfo_KvEntry.Size(m)
}
func (m *KvEntry) XXX_DiscardUnknown() {
	xxx_messageInfo_KvEntry.DiscardUnknown(m)
}

var xxx_messageInfo_KvEntry proto.InternalMessageInfo

func (m *KvEntry) GetMeta() *KvMeta {
	if m != nil {
		return m.Meta
	}
	return nil
}

func (m *KvEntry) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

type KvProgKeyEntry struct {
	Type                 uint32   `protobuf:"varint,1,opt,name=type" json:"type,omitempty"`
	Data                 []byte   `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KvProgKeyEntry) Reset()         { *m = KvProgKeyEntry{} }
func (m *KvProgKeyEntry) String() string { return proto.CompactTextString(m) }
func (*KvProgKeyEntry) ProtoMessage()    {}
func (*KvProgKeyEntry) Descriptor() ([]byte, []int) {
	return fileDescriptor_skv_f8d55fd67c11843a, []int{2}
}
func (m *KvProgKeyEntry) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KvProgKeyEntry.Unmarshal(m, b)
}
func (m *KvProgKeyEntry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KvProgKeyEntry.Marshal(b, m, deterministic)
}
func (dst *KvProgKeyEntry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KvProgKeyEntry.Merge(dst, src)
}
func (m *KvProgKeyEntry) XXX_Size() int {
	return xxx_messageInfo_KvProgKeyEntry.Size(m)
}
func (m *KvProgKeyEntry) XXX_DiscardUnknown() {
	xxx_messageInfo_KvProgKeyEntry.DiscardUnknown(m)
}

var xxx_messageInfo_KvProgKeyEntry proto.InternalMessageInfo

func (m *KvProgKeyEntry) GetType() uint32 {
	if m != nil {
		return m.Type
	}
	return 0
}

func (m *KvProgKeyEntry) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type KvProgKey struct {
	Items                []*KvProgKeyEntry `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *KvProgKey) Reset()         { *m = KvProgKey{} }
func (m *KvProgKey) String() string { return proto.CompactTextString(m) }
func (*KvProgKey) ProtoMessage()    {}
func (*KvProgKey) Descriptor() ([]byte, []int) {
	return fileDescriptor_skv_f8d55fd67c11843a, []int{3}
}
func (m *KvProgKey) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KvProgKey.Unmarshal(m, b)
}
func (m *KvProgKey) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KvProgKey.Marshal(b, m, deterministic)
}
func (dst *KvProgKey) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KvProgKey.Merge(dst, src)
}
func (m *KvProgKey) XXX_Size() int {
	return xxx_messageInfo_KvProgKey.Size(m)
}
func (m *KvProgKey) XXX_DiscardUnknown() {
	xxx_messageInfo_KvProgKey.DiscardUnknown(m)
}

var xxx_messageInfo_KvProgKey proto.InternalMessageInfo

func (m *KvProgKey) GetItems() []*KvProgKeyEntry {
	if m != nil {
		return m.Items
	}
	return nil
}

type KvProgWriteOptions struct {
	Version              uint64   `protobuf:"varint,1,opt,name=version" json:"version,omitempty"`
	Actions              uint64   `protobuf:"varint,2,opt,name=actions" json:"actions,omitempty"`
	Expired              uint64   `protobuf:"varint,3,opt,name=expired" json:"expired,omitempty"`
	PrevSum              uint32   `protobuf:"varint,4,opt,name=prev_sum,json=prevSum" json:"prev_sum,omitempty"`
	PrevVersion          uint64   `protobuf:"varint,5,opt,name=prev_version,json=prevVersion" json:"prev_version,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KvProgWriteOptions) Reset()         { *m = KvProgWriteOptions{} }
func (m *KvProgWriteOptions) String() string { return proto.CompactTextString(m) }
func (*KvProgWriteOptions) ProtoMessage()    {}
func (*KvProgWriteOptions) Descriptor() ([]byte, []int) {
	return fileDescriptor_skv_f8d55fd67c11843a, []int{4}
}
func (m *KvProgWriteOptions) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KvProgWriteOptions.Unmarshal(m, b)
}
func (m *KvProgWriteOptions) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KvProgWriteOptions.Marshal(b, m, deterministic)
}
func (dst *KvProgWriteOptions) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KvProgWriteOptions.Merge(dst, src)
}
func (m *KvProgWriteOptions) XXX_Size() int {
	return xxx_messageInfo_KvProgWriteOptions.Size(m)
}
func (m *KvProgWriteOptions) XXX_DiscardUnknown() {
	xxx_messageInfo_KvProgWriteOptions.DiscardUnknown(m)
}

var xxx_messageInfo_KvProgWriteOptions proto.InternalMessageInfo

func (m *KvProgWriteOptions) GetVersion() uint64 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *KvProgWriteOptions) GetActions() uint64 {
	if m != nil {
		return m.Actions
	}
	return 0
}

func (m *KvProgWriteOptions) GetExpired() uint64 {
	if m != nil {
		return m.Expired
	}
	return 0
}

func (m *KvProgWriteOptions) GetPrevSum() uint32 {
	if m != nil {
		return m.PrevSum
	}
	return 0
}

func (m *KvProgWriteOptions) GetPrevVersion() uint64 {
	if m != nil {
		return m.PrevVersion
	}
	return 0
}

type KvProgKeyValueCommit struct {
	Meta                 *KvMeta             `protobuf:"bytes,1,opt,name=meta" json:"meta,omitempty"`
	Key                  *KvProgKey          `protobuf:"bytes,2,opt,name=key" json:"key,omitempty"`
	Value                []byte              `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
	Options              *KvProgWriteOptions `protobuf:"bytes,4,opt,name=options" json:"options,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *KvProgKeyValueCommit) Reset()         { *m = KvProgKeyValueCommit{} }
func (m *KvProgKeyValueCommit) String() string { return proto.CompactTextString(m) }
func (*KvProgKeyValueCommit) ProtoMessage()    {}
func (*KvProgKeyValueCommit) Descriptor() ([]byte, []int) {
	return fileDescriptor_skv_f8d55fd67c11843a, []int{5}
}
func (m *KvProgKeyValueCommit) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KvProgKeyValueCommit.Unmarshal(m, b)
}
func (m *KvProgKeyValueCommit) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KvProgKeyValueCommit.Marshal(b, m, deterministic)
}
func (dst *KvProgKeyValueCommit) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KvProgKeyValueCommit.Merge(dst, src)
}
func (m *KvProgKeyValueCommit) XXX_Size() int {
	return xxx_messageInfo_KvProgKeyValueCommit.Size(m)
}
func (m *KvProgKeyValueCommit) XXX_DiscardUnknown() {
	xxx_messageInfo_KvProgKeyValueCommit.DiscardUnknown(m)
}

var xxx_messageInfo_KvProgKeyValueCommit proto.InternalMessageInfo

func (m *KvProgKeyValueCommit) GetMeta() *KvMeta {
	if m != nil {
		return m.Meta
	}
	return nil
}

func (m *KvProgKeyValueCommit) GetKey() *KvProgKey {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *KvProgKeyValueCommit) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *KvProgKeyValueCommit) GetOptions() *KvProgWriteOptions {
	if m != nil {
		return m.Options
	}
	return nil
}

type FileObjectEntryInit struct {
	Path                 string   `protobuf:"bytes,3,opt,name=path" json:"path,omitempty"`
	Size                 uint64   `protobuf:"varint,4,opt,name=size" json:"size,omitempty"`
	Attrs                uint64   `protobuf:"varint,5,opt,name=attrs" json:"attrs,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FileObjectEntryInit) Reset()         { *m = FileObjectEntryInit{} }
func (m *FileObjectEntryInit) String() string { return proto.CompactTextString(m) }
func (*FileObjectEntryInit) ProtoMessage()    {}
func (*FileObjectEntryInit) Descriptor() ([]byte, []int) {
	return fileDescriptor_skv_f8d55fd67c11843a, []int{6}
}
func (m *FileObjectEntryInit) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FileObjectEntryInit.Unmarshal(m, b)
}
func (m *FileObjectEntryInit) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FileObjectEntryInit.Marshal(b, m, deterministic)
}
func (dst *FileObjectEntryInit) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FileObjectEntryInit.Merge(dst, src)
}
func (m *FileObjectEntryInit) XXX_Size() int {
	return xxx_messageInfo_FileObjectEntryInit.Size(m)
}
func (m *FileObjectEntryInit) XXX_DiscardUnknown() {
	xxx_messageInfo_FileObjectEntryInit.DiscardUnknown(m)
}

var xxx_messageInfo_FileObjectEntryInit proto.InternalMessageInfo

func (m *FileObjectEntryInit) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *FileObjectEntryInit) GetSize() uint64 {
	if m != nil {
		return m.Size
	}
	return 0
}

func (m *FileObjectEntryInit) GetAttrs() uint64 {
	if m != nil {
		return m.Attrs
	}
	return 0
}

type FileObjectEntryMeta struct {
	Meta                 *KvMeta  `protobuf:"bytes,1,opt,name=meta" json:"meta,omitempty"`
	Path                 string   `protobuf:"bytes,3,opt,name=path" json:"path,omitempty"`
	Size                 uint64   `protobuf:"varint,4,opt,name=size" json:"size,omitempty"`
	Attrs                uint64   `protobuf:"varint,5,opt,name=attrs" json:"attrs,omitempty"`
	Sn                   uint32   `protobuf:"varint,7,opt,name=sn" json:"sn,omitempty"`
	CommitKey            string   `protobuf:"bytes,8,opt,name=commit_key,json=commitKey" json:"commit_key,omitempty"`
	Blocks               []uint32 `protobuf:"varint,9,rep,packed,name=blocks" json:"blocks,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FileObjectEntryMeta) Reset()         { *m = FileObjectEntryMeta{} }
func (m *FileObjectEntryMeta) String() string { return proto.CompactTextString(m) }
func (*FileObjectEntryMeta) ProtoMessage()    {}
func (*FileObjectEntryMeta) Descriptor() ([]byte, []int) {
	return fileDescriptor_skv_f8d55fd67c11843a, []int{7}
}
func (m *FileObjectEntryMeta) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FileObjectEntryMeta.Unmarshal(m, b)
}
func (m *FileObjectEntryMeta) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FileObjectEntryMeta.Marshal(b, m, deterministic)
}
func (dst *FileObjectEntryMeta) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FileObjectEntryMeta.Merge(dst, src)
}
func (m *FileObjectEntryMeta) XXX_Size() int {
	return xxx_messageInfo_FileObjectEntryMeta.Size(m)
}
func (m *FileObjectEntryMeta) XXX_DiscardUnknown() {
	xxx_messageInfo_FileObjectEntryMeta.DiscardUnknown(m)
}

var xxx_messageInfo_FileObjectEntryMeta proto.InternalMessageInfo

func (m *FileObjectEntryMeta) GetMeta() *KvMeta {
	if m != nil {
		return m.Meta
	}
	return nil
}

func (m *FileObjectEntryMeta) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *FileObjectEntryMeta) GetSize() uint64 {
	if m != nil {
		return m.Size
	}
	return 0
}

func (m *FileObjectEntryMeta) GetAttrs() uint64 {
	if m != nil {
		return m.Attrs
	}
	return 0
}

func (m *FileObjectEntryMeta) GetSn() uint32 {
	if m != nil {
		return m.Sn
	}
	return 0
}

func (m *FileObjectEntryMeta) GetCommitKey() string {
	if m != nil {
		return m.CommitKey
	}
	return ""
}

func (m *FileObjectEntryMeta) GetBlocks() []uint32 {
	if m != nil {
		return m.Blocks
	}
	return nil
}

type FileObjectEntryBlock struct {
	Meta                 *KvMeta  `protobuf:"bytes,1,opt,name=meta" json:"meta,omitempty"`
	Path                 string   `protobuf:"bytes,3,opt,name=path" json:"path,omitempty"`
	Size                 uint64   `protobuf:"varint,4,opt,name=size" json:"size,omitempty"`
	Attrs                uint64   `protobuf:"varint,5,opt,name=attrs" json:"attrs,omitempty"`
	Num                  uint32   `protobuf:"varint,6,opt,name=num" json:"num,omitempty"`
	Sum                  uint64   `protobuf:"varint,7,opt,name=sum" json:"sum,omitempty"`
	Sn                   uint32   `protobuf:"varint,8,opt,name=sn" json:"sn,omitempty"`
	Data                 []byte   `protobuf:"bytes,9,opt,name=data,proto3" json:"data,omitempty"`
	CommitKey            string   `protobuf:"bytes,10,opt,name=commit_key,json=commitKey" json:"commit_key,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FileObjectEntryBlock) Reset()         { *m = FileObjectEntryBlock{} }
func (m *FileObjectEntryBlock) String() string { return proto.CompactTextString(m) }
func (*FileObjectEntryBlock) ProtoMessage()    {}
func (*FileObjectEntryBlock) Descriptor() ([]byte, []int) {
	return fileDescriptor_skv_f8d55fd67c11843a, []int{8}
}
func (m *FileObjectEntryBlock) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FileObjectEntryBlock.Unmarshal(m, b)
}
func (m *FileObjectEntryBlock) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FileObjectEntryBlock.Marshal(b, m, deterministic)
}
func (dst *FileObjectEntryBlock) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FileObjectEntryBlock.Merge(dst, src)
}
func (m *FileObjectEntryBlock) XXX_Size() int {
	return xxx_messageInfo_FileObjectEntryBlock.Size(m)
}
func (m *FileObjectEntryBlock) XXX_DiscardUnknown() {
	xxx_messageInfo_FileObjectEntryBlock.DiscardUnknown(m)
}

var xxx_messageInfo_FileObjectEntryBlock proto.InternalMessageInfo

func (m *FileObjectEntryBlock) GetMeta() *KvMeta {
	if m != nil {
		return m.Meta
	}
	return nil
}

func (m *FileObjectEntryBlock) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *FileObjectEntryBlock) GetSize() uint64 {
	if m != nil {
		return m.Size
	}
	return 0
}

func (m *FileObjectEntryBlock) GetAttrs() uint64 {
	if m != nil {
		return m.Attrs
	}
	return 0
}

func (m *FileObjectEntryBlock) GetNum() uint32 {
	if m != nil {
		return m.Num
	}
	return 0
}

func (m *FileObjectEntryBlock) GetSum() uint64 {
	if m != nil {
		return m.Sum
	}
	return 0
}

func (m *FileObjectEntryBlock) GetSn() uint32 {
	if m != nil {
		return m.Sn
	}
	return 0
}

func (m *FileObjectEntryBlock) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *FileObjectEntryBlock) GetCommitKey() string {
	if m != nil {
		return m.CommitKey
	}
	return ""
}

func init() {
	proto.RegisterType((*KvMeta)(nil), "skv.KvMeta")
	proto.RegisterType((*KvEntry)(nil), "skv.KvEntry")
	proto.RegisterType((*KvProgKeyEntry)(nil), "skv.KvProgKeyEntry")
	proto.RegisterType((*KvProgKey)(nil), "skv.KvProgKey")
	proto.RegisterType((*KvProgWriteOptions)(nil), "skv.KvProgWriteOptions")
	proto.RegisterType((*KvProgKeyValueCommit)(nil), "skv.KvProgKeyValueCommit")
	proto.RegisterType((*FileObjectEntryInit)(nil), "skv.FileObjectEntryInit")
	proto.RegisterType((*FileObjectEntryMeta)(nil), "skv.FileObjectEntryMeta")
	proto.RegisterType((*FileObjectEntryBlock)(nil), "skv.FileObjectEntryBlock")
}

func init() { proto.RegisterFile("skv.proto", fileDescriptor_skv_f8d55fd67c11843a) }

var fileDescriptor_skv_f8d55fd67c11843a = []byte{
	// 539 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x54, 0xcd, 0x8e, 0xd3, 0x30,
	0x10, 0x96, 0x9b, 0xf4, 0x27, 0x93, 0x6d, 0x85, 0xbc, 0x15, 0x98, 0x03, 0x22, 0xe4, 0x14, 0x2e,
	0x2b, 0x51, 0x24, 0xc4, 0x11, 0x2d, 0x02, 0x81, 0x2a, 0xb4, 0xc8, 0x2b, 0x2d, 0xc7, 0x95, 0x9b,
	0x5a, 0x10, 0xda, 0xfc, 0x28, 0x76, 0x22, 0xca, 0xd3, 0x20, 0x5e, 0x83, 0x97, 0xe0, 0x01, 0x78,
	0x18, 0xe4, 0xb1, 0xd3, 0x0d, 0x5d, 0xc4, 0x1e, 0xd0, 0xde, 0xe6, 0x9b, 0x2f, 0x1e, 0xcf, 0xf7,
	0x8d, 0x27, 0x10, 0xa8, 0x4d, 0x7b, 0x52, 0xd5, 0xa5, 0x2e, 0xa9, 0xa7, 0x36, 0x6d, 0xfc, 0x93,
	0xc0, 0x68, 0xd9, 0xbe, 0x93, 0x5a, 0x50, 0x0a, 0xbe, 0xde, 0x55, 0x92, 0x91, 0x88, 0x24, 0x53,
	0x8e, 0x31, 0x65, 0x30, 0x6e, 0x65, 0xad, 0xb2, 0xb2, 0x60, 0x83, 0x88, 0x24, 0x3e, 0xef, 0xa0,
	0x61, 0xe4, 0x97, 0x2a, 0xab, 0xe5, 0x9a, 0x79, 0x96, 0x71, 0xd0, 0x30, 0x69, 0x2d, 0x85, 0x96,
	0x6b, 0xe6, 0x5b, 0xc6, 0x41, 0xc3, 0x34, 0xd5, 0x1a, 0x99, 0xa1, 0x65, 0x1c, 0x34, 0x77, 0xab,
	0xec, 0xab, 0x64, 0x23, 0x4c, 0x63, 0x4c, 0xef, 0x80, 0xa7, 0x9a, 0x9c, 0x8d, 0xb1, 0x1d, 0x13,
	0x9a, 0x4c, 0xd1, 0xe4, 0x6c, 0x82, 0x1f, 0x99, 0xd0, 0x9c, 0x2b, 0x44, 0x2e, 0xd9, 0x34, 0x22,
	0x49, 0xc0, 0x31, 0x8e, 0x5f, 0xc0, 0x78, 0xd9, 0xbe, 0x2a, 0x74, 0xbd, 0xa3, 0x0f, 0xc1, 0xcf,
	0xa5, 0x16, 0x28, 0x29, 0x5c, 0x84, 0x27, 0x46, 0xbc, 0x55, 0xcb, 0x91, 0xa0, 0x73, 0x18, 0xb6,
	0x62, 0xdb, 0x48, 0x54, 0x77, 0xc4, 0x2d, 0x88, 0x9f, 0xc3, 0x6c, 0xd9, 0xbe, 0xaf, 0xcb, 0x8f,
	0x4b, 0xb9, 0xb3, 0x85, 0xfe, 0xe6, 0x0d, 0x05, 0x7f, 0x2d, 0xb4, 0x70, 0x47, 0x31, 0x8e, 0x9f,
	0x41, 0xb0, 0x3f, 0x49, 0x1f, 0xc3, 0x30, 0xd3, 0x32, 0x57, 0x8c, 0x44, 0x5e, 0x12, 0x2e, 0x8e,
	0xdd, 0xf5, 0xfd, 0xc2, 0xdc, 0x7e, 0x11, 0x7f, 0x23, 0x40, 0x2d, 0xf3, 0xa1, 0xce, 0xb4, 0x3c,
	0xab, 0x74, 0x56, 0x16, 0xaa, 0x6f, 0x3f, 0xb9, 0x66, 0xbf, 0x48, 0xf1, 0xa3, 0x6e, 0x30, 0x0e,
	0xfe, 0x63, 0x30, 0xf7, 0x61, 0x52, 0xd5, 0xb2, 0xbd, 0x34, 0xae, 0xfa, 0x28, 0x64, 0x6c, 0xf0,
	0x79, 0x93, 0xd3, 0x47, 0x70, 0x84, 0x54, 0x77, 0x9b, 0x1d, 0x4f, 0x68, 0x72, 0x17, 0x36, 0x15,
	0x7f, 0x27, 0x30, 0xdf, 0x37, 0x7f, 0x61, 0x7c, 0x7a, 0x59, 0xe6, 0x79, 0xa6, 0x6f, 0x36, 0x39,
	0x02, 0x6f, 0x23, 0x77, 0xd8, 0x67, 0xb8, 0x98, 0xfd, 0xe9, 0x02, 0x37, 0xd4, 0xd5, 0x18, 0xbc,
	0xde, 0x18, 0xe8, 0x13, 0x18, 0x97, 0xd6, 0x08, 0x6c, 0x37, 0x5c, 0xdc, 0xeb, 0x9d, 0xed, 0xfb,
	0xc4, 0xbb, 0xef, 0xe2, 0x73, 0x38, 0x7e, 0x9d, 0x6d, 0xe5, 0xd9, 0xea, 0xb3, 0x4c, 0x35, 0x3a,
	0xfc, 0xb6, 0xc8, 0xb4, 0x19, 0x55, 0x25, 0xf4, 0x27, 0x2c, 0x1f, 0x70, 0x8c, 0xf7, 0x4f, 0xce,
	0xef, 0x3d, 0xb9, 0x39, 0x0c, 0x85, 0xd6, 0xb5, 0x72, 0xfa, 0x2d, 0x88, 0x7f, 0x90, 0x6b, 0x55,
	0x71, 0x61, 0x6e, 0x14, 0xfe, 0x5f, 0xd7, 0xd2, 0x19, 0x0c, 0x54, 0xe1, 0x9e, 0xff, 0x40, 0x15,
	0xf4, 0x01, 0x40, 0x8a, 0x8e, 0x5f, 0x1a, 0x37, 0x27, 0x58, 0x33, 0xb0, 0x19, 0xf3, 0xda, 0xee,
	0xc2, 0x68, 0xb5, 0x2d, 0xd3, 0x8d, 0x62, 0x41, 0xe4, 0x25, 0x53, 0xee, 0x50, 0xfc, 0x8b, 0xc0,
	0xfc, 0xa0, 0xfb, 0x53, 0xc3, 0xdc, 0x76, 0xfb, 0x6e, 0x59, 0x47, 0x76, 0x7d, 0x0b, 0xbb, 0xbe,
	0xdd, 0x42, 0xfb, 0x76, 0xa1, 0xad, 0xc4, 0xc9, 0x5e, 0x62, 0xb7, 0x52, 0xc1, 0xd5, 0x4a, 0x1d,
	0xc8, 0x86, 0x03, 0xd9, 0xa7, 0x83, 0x37, 0xde, 0x6a, 0x84, 0x3f, 0xb4, 0xa7, 0xbf, 0x03, 0x00,
	0x00, 0xff, 0xff, 0xc9, 0x8a, 0xf4, 0xc9, 0xdd, 0x04, 0x00, 0x00,
}
