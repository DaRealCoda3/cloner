// Code generated by protoc-gen-go. DO NOT EDIT.
// source: logutil.proto

package logutil

import (
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	vttime "vitess.io/vitess/go/vt/proto/vttime"
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

// Level is the level of the log messages.
type Level int32

const (
	// The usual logging levels.
	// Should be logged using logging facility.
	Level_INFO    Level = 0
	Level_WARNING Level = 1
	Level_ERROR   Level = 2
	// For messages that may contains non-logging events.
	// Should be logged to console directly.
	Level_CONSOLE Level = 3
)

var Level_name = map[int32]string{
	0: "INFO",
	1: "WARNING",
	2: "ERROR",
	3: "CONSOLE",
}

var Level_value = map[string]int32{
	"INFO":    0,
	"WARNING": 1,
	"ERROR":   2,
	"CONSOLE": 3,
}

func (x Level) String() string {
	return proto.EnumName(Level_name, int32(x))
}

func (Level) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_31f5dd3702a8edf9, []int{0}
}

// Event is a single logging event
type Event struct {
	Time                 *vttime.Time `protobuf:"bytes,1,opt,name=time,proto3" json:"time,omitempty"`
	Level                Level        `protobuf:"varint,2,opt,name=level,proto3,enum=logutil.Level" json:"level,omitempty"`
	File                 string       `protobuf:"bytes,3,opt,name=file,proto3" json:"file,omitempty"`
	Line                 int64        `protobuf:"varint,4,opt,name=line,proto3" json:"line,omitempty"`
	Value                string       `protobuf:"bytes,5,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *Event) Reset()         { *m = Event{} }
func (m *Event) String() string { return proto.CompactTextString(m) }
func (*Event) ProtoMessage()    {}
func (*Event) Descriptor() ([]byte, []int) {
	return fileDescriptor_31f5dd3702a8edf9, []int{0}
}

func (m *Event) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Event.Unmarshal(m, b)
}
func (m *Event) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Event.Marshal(b, m, deterministic)
}
func (m *Event) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Event.Merge(m, src)
}
func (m *Event) XXX_Size() int {
	return xxx_messageInfo_Event.Size(m)
}
func (m *Event) XXX_DiscardUnknown() {
	xxx_messageInfo_Event.DiscardUnknown(m)
}

var xxx_messageInfo_Event proto.InternalMessageInfo

func (m *Event) GetTime() *vttime.Time {
	if m != nil {
		return m.Time
	}
	return nil
}

func (m *Event) GetLevel() Level {
	if m != nil {
		return m.Level
	}
	return Level_INFO
}

func (m *Event) GetFile() string {
	if m != nil {
		return m.File
	}
	return ""
}

func (m *Event) GetLine() int64 {
	if m != nil {
		return m.Line
	}
	return 0
}

func (m *Event) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

func init() {
	proto.RegisterEnum("logutil.Level", Level_name, Level_value)
	proto.RegisterType((*Event)(nil), "logutil.Event")
}

func init() { proto.RegisterFile("logutil.proto", fileDescriptor_31f5dd3702a8edf9) }

var fileDescriptor_31f5dd3702a8edf9 = []byte{
	// 236 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x34, 0x8f, 0x5f, 0x4b, 0xc3, 0x30,
	0x14, 0xc5, 0xcd, 0xda, 0x38, 0x77, 0x37, 0x47, 0xb9, 0xf8, 0x10, 0x7c, 0x0a, 0x32, 0xa4, 0xf8,
	0xd0, 0xc0, 0x04, 0xdf, 0x55, 0xaa, 0x0c, 0x46, 0x0b, 0x57, 0x41, 0xf0, 0x4d, 0xe1, 0x3a, 0x02,
	0xd9, 0x22, 0x2e, 0xcd, 0xb7, 0xf0, 0x3b, 0x4b, 0xd3, 0xfa, 0x76, 0xce, 0xef, 0x1c, 0xee, 0x1f,
	0x38, 0x77, 0x7e, 0xd7, 0x05, 0xeb, 0xaa, 0xef, 0x1f, 0x1f, 0x3c, 0x4e, 0x47, 0x7b, 0xb9, 0x88,
	0x21, 0xd8, 0x3d, 0x0f, 0xf8, 0xea, 0x57, 0x80, 0xac, 0x23, 0x1f, 0x02, 0x6a, 0xc8, 0x7b, 0xae,
	0x84, 0x16, 0xe5, 0x7c, 0xbd, 0xa8, 0xc6, 0xda, 0xab, 0xdd, 0x33, 0xa5, 0x04, 0x57, 0x20, 0x1d,
	0x47, 0x76, 0x6a, 0xa2, 0x45, 0xb9, 0x5c, 0x2f, 0xab, 0xff, 0x0d, 0xdb, 0x9e, 0xd2, 0x10, 0x22,
	0x42, 0xfe, 0x65, 0x1d, 0xab, 0x4c, 0x8b, 0x72, 0x46, 0x49, 0xf7, 0xcc, 0xd9, 0x03, 0xab, 0x5c,
	0x8b, 0x32, 0xa3, 0xa4, 0xf1, 0x02, 0x64, 0xfc, 0x70, 0x1d, 0x2b, 0x99, 0x8a, 0x83, 0xb9, 0xb9,
	0x03, 0x99, 0xa6, 0xe1, 0x19, 0xe4, 0x9b, 0xe6, 0xa9, 0x2d, 0x4e, 0x70, 0x0e, 0xd3, 0xb7, 0x7b,
	0x6a, 0x36, 0xcd, 0x73, 0x21, 0x70, 0x06, 0xb2, 0x26, 0x6a, 0xa9, 0x98, 0xf4, 0xfc, 0xb1, 0x6d,
	0x5e, 0xda, 0x6d, 0x5d, 0x64, 0x0f, 0xd7, 0xef, 0xab, 0x68, 0x03, 0x1f, 0x8f, 0x95, 0xf5, 0x66,
	0x50, 0x66, 0xe7, 0x4d, 0x0c, 0x26, 0xfd, 0x69, 0xc6, 0x53, 0x3f, 0x4f, 0x93, 0xbd, 0xfd, 0x0b,
	0x00, 0x00, 0xff, 0xff, 0xa4, 0x27, 0x83, 0x63, 0x1e, 0x01, 0x00, 0x00,
}