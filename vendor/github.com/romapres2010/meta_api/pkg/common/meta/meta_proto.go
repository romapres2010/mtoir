package meta

//
//import (
//	"google.golang.org/protobuf/reflect/protoreflect"
//	"google.golang.org/protobuf/runtime/protoimpl"
//	"google.golang.org/protobuf/types/known/timestamppb"
//	"reflect"
//	"sync"
//)
//
//type ProtoMessage struct {
//	state         protoimpl.MessageState
//	sizeCache     protoimpl.SizeCache
//	unknownFields protoimpl.UnknownFields
//
//	V interface{} `protobuf:"bytes,1,opt,name=val" json:"val,omitempty"`
//}
//
//func (x *ProtoMessage) Reset() {
//	*x = ProtoMessage{}
//	if protoimpl.UnsafeEnabled {
//		mi := &file_io_prometheus_client_metrics_proto_msgTypes[11]
//		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
//		ms.StoreMessageInfo(mi)
//	}
//}
//
//func (x *ProtoMessage) String() string {
//	return protoimpl.X.MessageStringOf(x)
//}
//
//func (*ProtoMessage) ProtoMessage() {}
//
//func (x *ProtoMessage) ProtoReflect() protoreflect.Message {
//	mi := &file_io_prometheus_client_metrics_proto_msgTypes[11]
//	if protoimpl.UnsafeEnabled && x != nil {
//		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
//		if ms.LoadMessageInfo() == nil {
//			ms.StoreMessageInfo(mi)
//		}
//		return ms
//	}
//	return mi.MessageOf(x)
//}
//
//var File_io_prometheus_client_metrics_proto protoreflect.FileDescriptor
//
//var file_io_prometheus_client_metrics_proto_rawDesc = []byte{
//	0x0a, 0x22, 0x69, 0x6f, 0x2f, 0x70, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75, 0x73, 0x2f,
//	0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2f, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2e, 0x70,
//	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14, 0x69, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68,
//	0x65, 0x75, 0x73, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67,
//	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65,
//	0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x35, 0x0a, 0x09, 0x4c,
//	0x61, 0x62, 0x65, 0x6c, 0x50, 0x61, 0x69, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
//	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05,
//	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c,
//	0x75, 0x65, 0x22, 0x1d, 0x0a, 0x05, 0x47, 0x61, 0x75, 0x67, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76,
//	0x61, 0x6c, 0x75, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x01, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
//	0x65, 0x22, 0x5b, 0x0a, 0x07, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x65, 0x72, 0x12, 0x14, 0x0a, 0x05,
//	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x01, 0x52, 0x05, 0x76, 0x61, 0x6c,
//	0x75, 0x65, 0x12, 0x3a, 0x0a, 0x08, 0x65, 0x78, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x72, 0x18, 0x02,
//	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x69, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x65, 0x74,
//	0x68, 0x65, 0x75, 0x73, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2e, 0x45, 0x78, 0x65, 0x6d,
//	0x70, 0x6c, 0x61, 0x72, 0x52, 0x08, 0x65, 0x78, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x72, 0x22, 0x3c,
//	0x0a, 0x08, 0x51, 0x75, 0x61, 0x6e, 0x74, 0x69, 0x6c, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x71, 0x75,
//	0x61, 0x6e, 0x74, 0x69, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x01, 0x52, 0x08, 0x71, 0x75,
//	0x61, 0x6e, 0x74, 0x69, 0x6c, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
//	0x02, 0x20, 0x01, 0x28, 0x01, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x87, 0x01, 0x0a,
//	0x07, 0x53, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x12, 0x21, 0x0a, 0x0c, 0x73, 0x61, 0x6d, 0x70,
//	0x6c, 0x65, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0b,
//	0x73, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x73,
//	0x61, 0x6d, 0x70, 0x6c, 0x65, 0x5f, 0x73, 0x75, 0x6d, 0x18, 0x02, 0x20, 0x01, 0x28, 0x01, 0x52,
//	0x09, 0x73, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x53, 0x75, 0x6d, 0x12, 0x3a, 0x0a, 0x08, 0x71, 0x75,
//	0x61, 0x6e, 0x74, 0x69, 0x6c, 0x65, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x69,
//	0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75, 0x73, 0x2e, 0x63, 0x6c, 0x69,
//	0x65, 0x6e, 0x74, 0x2e, 0x51, 0x75, 0x61, 0x6e, 0x74, 0x69, 0x6c, 0x65, 0x52, 0x08, 0x71, 0x75,
//	0x61, 0x6e, 0x74, 0x69, 0x6c, 0x65, 0x22, 0x1f, 0x0a, 0x07, 0x55, 0x6e, 0x74, 0x79, 0x70, 0x65,
//	0x64, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x01,
//	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0xe3, 0x04, 0x0a, 0x09, 0x48, 0x69, 0x73, 0x74,
//	0x6f, 0x67, 0x72, 0x61, 0x6d, 0x12, 0x21, 0x0a, 0x0c, 0x73, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x5f,
//	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0b, 0x73, 0x61, 0x6d,
//	0x70, 0x6c, 0x65, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x2c, 0x0a, 0x12, 0x73, 0x61, 0x6d, 0x70,
//	0x6c, 0x65, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x66, 0x6c, 0x6f, 0x61, 0x74, 0x18, 0x04,
//	0x20, 0x01, 0x28, 0x01, 0x52, 0x10, 0x73, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x43, 0x6f, 0x75, 0x6e,
//	0x74, 0x46, 0x6c, 0x6f, 0x61, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x61, 0x6d, 0x70, 0x6c, 0x65,
//	0x5f, 0x73, 0x75, 0x6d, 0x18, 0x02, 0x20, 0x01, 0x28, 0x01, 0x52, 0x09, 0x73, 0x61, 0x6d, 0x70,
//	0x6c, 0x65, 0x53, 0x75, 0x6d, 0x12, 0x34, 0x0a, 0x06, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x18,
//	0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x69, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x65,
//	0x74, 0x68, 0x65, 0x75, 0x73, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2e, 0x42, 0x75, 0x63,
//	0x6b, 0x65, 0x74, 0x52, 0x06, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x73,
//	0x63, 0x68, 0x65, 0x6d, 0x61, 0x18, 0x05, 0x20, 0x01, 0x28, 0x11, 0x52, 0x06, 0x73, 0x63, 0x68,
//	0x65, 0x6d, 0x61, 0x12, 0x25, 0x0a, 0x0e, 0x7a, 0x65, 0x72, 0x6f, 0x5f, 0x74, 0x68, 0x72, 0x65,
//	0x73, 0x68, 0x6f, 0x6c, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x01, 0x52, 0x0d, 0x7a, 0x65, 0x72,
//	0x6f, 0x54, 0x68, 0x72, 0x65, 0x73, 0x68, 0x6f, 0x6c, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x7a, 0x65,
//	0x72, 0x6f, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x04, 0x52, 0x09,
//	0x7a, 0x65, 0x72, 0x6f, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x28, 0x0a, 0x10, 0x7a, 0x65, 0x72,
//	0x6f, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x66, 0x6c, 0x6f, 0x61, 0x74, 0x18, 0x08, 0x20,
//	0x01, 0x28, 0x01, 0x52, 0x0e, 0x7a, 0x65, 0x72, 0x6f, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x46, 0x6c,
//	0x6f, 0x61, 0x74, 0x12, 0x45, 0x0a, 0x0d, 0x6e, 0x65, 0x67, 0x61, 0x74, 0x69, 0x76, 0x65, 0x5f,
//	0x73, 0x70, 0x61, 0x6e, 0x18, 0x09, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x69, 0x6f, 0x2e,
//	0x70, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75, 0x73, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e,
//	0x74, 0x2e, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x53, 0x70, 0x61, 0x6e, 0x52, 0x0c, 0x6e, 0x65,
//	0x67, 0x61, 0x74, 0x69, 0x76, 0x65, 0x53, 0x70, 0x61, 0x6e, 0x12, 0x25, 0x0a, 0x0e, 0x6e, 0x65,
//	0x67, 0x61, 0x74, 0x69, 0x76, 0x65, 0x5f, 0x64, 0x65, 0x6c, 0x74, 0x61, 0x18, 0x0a, 0x20, 0x03,
//	0x28, 0x12, 0x52, 0x0d, 0x6e, 0x65, 0x67, 0x61, 0x74, 0x69, 0x76, 0x65, 0x44, 0x65, 0x6c, 0x74,
//	0x61, 0x12, 0x25, 0x0a, 0x0e, 0x6e, 0x65, 0x67, 0x61, 0x74, 0x69, 0x76, 0x65, 0x5f, 0x63, 0x6f,
//	0x75, 0x6e, 0x74, 0x18, 0x0b, 0x20, 0x03, 0x28, 0x01, 0x52, 0x0d, 0x6e, 0x65, 0x67, 0x61, 0x74,
//	0x69, 0x76, 0x65, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x45, 0x0a, 0x0d, 0x70, 0x6f, 0x73, 0x69,
//	0x74, 0x69, 0x76, 0x65, 0x5f, 0x73, 0x70, 0x61, 0x6e, 0x18, 0x0c, 0x20, 0x03, 0x28, 0x0b, 0x32,
//	0x20, 0x2e, 0x69, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75, 0x73, 0x2e,
//	0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2e, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x53, 0x70, 0x61,
//	0x6e, 0x52, 0x0c, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x76, 0x65, 0x53, 0x70, 0x61, 0x6e, 0x12,
//	0x25, 0x0a, 0x0e, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x76, 0x65, 0x5f, 0x64, 0x65, 0x6c, 0x74,
//	0x61, 0x18, 0x0d, 0x20, 0x03, 0x28, 0x12, 0x52, 0x0d, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x76,
//	0x65, 0x44, 0x65, 0x6c, 0x74, 0x61, 0x12, 0x25, 0x0a, 0x0e, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x69,
//	0x76, 0x65, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x0e, 0x20, 0x03, 0x28, 0x01, 0x52, 0x0d,
//	0x70, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x76, 0x65, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0xc6, 0x01,
//	0x0a, 0x06, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x29, 0x0a, 0x10, 0x63, 0x75, 0x6d, 0x75,
//	0x6c, 0x61, 0x74, 0x69, 0x76, 0x65, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01,
//	0x28, 0x04, 0x52, 0x0f, 0x63, 0x75, 0x6d, 0x75, 0x6c, 0x61, 0x74, 0x69, 0x76, 0x65, 0x43, 0x6f,
//	0x75, 0x6e, 0x74, 0x12, 0x34, 0x0a, 0x16, 0x63, 0x75, 0x6d, 0x75, 0x6c, 0x61, 0x74, 0x69, 0x76,
//	0x65, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x66, 0x6c, 0x6f, 0x61, 0x74, 0x18, 0x04, 0x20,
//	0x01, 0x28, 0x01, 0x52, 0x14, 0x63, 0x75, 0x6d, 0x75, 0x6c, 0x61, 0x74, 0x69, 0x76, 0x65, 0x43,
//	0x6f, 0x75, 0x6e, 0x74, 0x46, 0x6c, 0x6f, 0x61, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x75, 0x70, 0x70,
//	0x65, 0x72, 0x5f, 0x62, 0x6f, 0x75, 0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x01, 0x52, 0x0a,
//	0x75, 0x70, 0x70, 0x65, 0x72, 0x42, 0x6f, 0x75, 0x6e, 0x64, 0x12, 0x3a, 0x0a, 0x08, 0x65, 0x78,
//	0x65, 0x6d, 0x70, 0x6c, 0x61, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x69,
//	0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75, 0x73, 0x2e, 0x63, 0x6c, 0x69,
//	0x65, 0x6e, 0x74, 0x2e, 0x45, 0x78, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x72, 0x52, 0x08, 0x65, 0x78,
//	0x65, 0x6d, 0x70, 0x6c, 0x61, 0x72, 0x22, 0x3c, 0x0a, 0x0a, 0x42, 0x75, 0x63, 0x6b, 0x65, 0x74,
//	0x53, 0x70, 0x61, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x18, 0x01,
//	0x20, 0x01, 0x28, 0x11, 0x52, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x12, 0x16, 0x0a, 0x06,
//	0x6c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x06, 0x6c, 0x65,
//	0x6e, 0x67, 0x74, 0x68, 0x22, 0x91, 0x01, 0x0a, 0x08, 0x45, 0x78, 0x65, 0x6d, 0x70, 0x6c, 0x61,
//	0x72, 0x12, 0x35, 0x0a, 0x05, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
//	0x32, 0x1f, 0x2e, 0x69, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75, 0x73,
//	0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2e, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x50, 0x61, 0x69,
//	0x72, 0x52, 0x05, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75,
//	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x01, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x38,
//	0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28,
//	0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
//	0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x74,
//	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x22, 0xff, 0x02, 0x0a, 0x06, 0x4d, 0x65, 0x74,
//	0x72, 0x69, 0x63, 0x12, 0x35, 0x0a, 0x05, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x18, 0x01, 0x20, 0x03,
//	0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x69, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65,
//	0x75, 0x73, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2e, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x50,
//	0x61, 0x69, 0x72, 0x52, 0x05, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x12, 0x31, 0x0a, 0x05, 0x67, 0x61,
//	0x75, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x69, 0x6f, 0x2e, 0x70,
//	0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75, 0x73, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74,
//	0x2e, 0x47, 0x61, 0x75, 0x67, 0x65, 0x52, 0x05, 0x67, 0x61, 0x75, 0x67, 0x65, 0x12, 0x37, 0x0a,
//	0x07, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d,
//	0x2e, 0x69, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75, 0x73, 0x2e, 0x63,
//	0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2e, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x65, 0x72, 0x52, 0x07, 0x63,
//	0x6f, 0x75, 0x6e, 0x74, 0x65, 0x72, 0x12, 0x37, 0x0a, 0x07, 0x73, 0x75, 0x6d, 0x6d, 0x61, 0x72,
//	0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x69, 0x6f, 0x2e, 0x70, 0x72, 0x6f,
//	0x6d, 0x65, 0x74, 0x68, 0x65, 0x75, 0x73, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2e, 0x53,
//	0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x52, 0x07, 0x73, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x12,
//	0x37, 0x0a, 0x07, 0x75, 0x6e, 0x74, 0x79, 0x70, 0x65, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b,
//	0x32, 0x1d, 0x2e, 0x69, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75, 0x73,
//	0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2e, 0x55, 0x6e, 0x74, 0x79, 0x70, 0x65, 0x64, 0x52,
//	0x07, 0x75, 0x6e, 0x74, 0x79, 0x70, 0x65, 0x64, 0x12, 0x3d, 0x0a, 0x09, 0x68, 0x69, 0x73, 0x74,
//	0x6f, 0x67, 0x72, 0x61, 0x6d, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x69, 0x6f,
//	0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75, 0x73, 0x2e, 0x63, 0x6c, 0x69, 0x65,
//	0x6e, 0x74, 0x2e, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x67, 0x72, 0x61, 0x6d, 0x52, 0x09, 0x68, 0x69,
//	0x73, 0x74, 0x6f, 0x67, 0x72, 0x61, 0x6d, 0x12, 0x21, 0x0a, 0x0c, 0x74, 0x69, 0x6d, 0x65, 0x73,
//	0x74, 0x61, 0x6d, 0x70, 0x5f, 0x6d, 0x73, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0b, 0x74,
//	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x4d, 0x73, 0x22, 0xa2, 0x01, 0x0a, 0x0c, 0x4d,
//	0x65, 0x74, 0x72, 0x69, 0x63, 0x46, 0x61, 0x6d, 0x69, 0x6c, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x6e,
//	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12,
//	0x12, 0x0a, 0x04, 0x68, 0x65, 0x6c, 0x70, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68,
//	0x65, 0x6c, 0x70, 0x12, 0x34, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28,
//	0x0e, 0x32, 0x20, 0x2e, 0x69, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75,
//	0x73, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x54,
//	0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x34, 0x0a, 0x06, 0x6d, 0x65, 0x74,
//	0x72, 0x69, 0x63, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x69, 0x6f, 0x2e, 0x70,
//	0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75, 0x73, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74,
//	0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x06, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x2a,
//	0x62, 0x0a, 0x0a, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0b, 0x0a,
//	0x07, 0x43, 0x4f, 0x55, 0x4e, 0x54, 0x45, 0x52, 0x10, 0x00, 0x12, 0x09, 0x0a, 0x05, 0x47, 0x41,
//	0x55, 0x47, 0x45, 0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07, 0x53, 0x55, 0x4d, 0x4d, 0x41, 0x52, 0x59,
//	0x10, 0x02, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x54, 0x59, 0x50, 0x45, 0x44, 0x10, 0x03, 0x12,
//	0x0d, 0x0a, 0x09, 0x48, 0x49, 0x53, 0x54, 0x4f, 0x47, 0x52, 0x41, 0x4d, 0x10, 0x04, 0x12, 0x13,
//	0x0a, 0x0f, 0x47, 0x41, 0x55, 0x47, 0x45, 0x5f, 0x48, 0x49, 0x53, 0x54, 0x4f, 0x47, 0x52, 0x41,
//	0x4d, 0x10, 0x05, 0x42, 0x52, 0x0a, 0x14, 0x69, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x65, 0x74,
//	0x68, 0x65, 0x75, 0x73, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5a, 0x3a, 0x67, 0x69, 0x74,
//	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65,
//	0x75, 0x73, 0x2f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f,
//	0x67, 0x6f, 0x3b, 0x69, 0x6f, 0x5f, 0x70, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75, 0x73,
//	//0x5f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74,
//	0x5f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x75,
//}
//
//var (
//	file_io_prometheus_client_metrics_proto_rawDescOnce sync.Once
//	file_io_prometheus_client_metrics_proto_rawDescData = file_io_prometheus_client_metrics_proto_rawDesc
//)
//
//func file_io_prometheus_client_metrics_proto_rawDescGZIP() []byte {
//	file_io_prometheus_client_metrics_proto_rawDescOnce.Do(func() {
//		file_io_prometheus_client_metrics_proto_rawDescData = protoimpl.X.CompressGZIP(file_io_prometheus_client_metrics_proto_rawDescData)
//	})
//	return file_io_prometheus_client_metrics_proto_rawDescData
//}
//
//var file_io_prometheus_client_metrics_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
//var file_io_prometheus_client_metrics_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
//var file_io_prometheus_client_metrics_proto_goTypes = []interface{}{
//	(*ProtoMessage)(nil),          // 12: io.prometheus.client.ProtoMessage
//	(*timestamppb.Timestamp)(nil), // 13: google.protobuf.Timestamp
//}
//var file_io_prometheus_client_metrics_proto_depIdxs = []int32{
//	10, // 0: io.prometheus.client.Counter.exemplar:type_name -> io.prometheus.client.Exemplar
//	4,  // 1: io.prometheus.client.Summary.quantile:type_name -> io.prometheus.client.Quantile
//	8,  // 2: io.prometheus.client.Histogram.bucket:type_name -> io.prometheus.client.Bucket
//	9,  // 3: io.prometheus.client.Histogram.negative_span:type_name -> io.prometheus.client.BucketSpan
//	9,  // 4: io.prometheus.client.Histogram.positive_span:type_name -> io.prometheus.client.BucketSpan
//	10, // 5: io.prometheus.client.Bucket.exemplar:type_name -> io.prometheus.client.Exemplar
//	1,  // 6: io.prometheus.client.Exemplar.label:type_name -> io.prometheus.client.LabelPair
//	13, // 7: io.prometheus.client.Exemplar.timestamp:type_name -> google.protobuf.Timestamp
//	1,  // 8: io.prometheus.client.Metric.label:type_name -> io.prometheus.client.LabelPair
//	2,  // 9: io.prometheus.client.Metric.gauge:type_name -> io.prometheus.client.Gauge
//	3,  // 10: io.prometheus.client.Metric.counter:type_name -> io.prometheus.client.Counter
//	5,  // 11: io.prometheus.client.Metric.summary:type_name -> io.prometheus.client.Summary
//	6,  // 12: io.prometheus.client.Metric.untyped:type_name -> io.prometheus.client.Untyped
//	7,  // 13: io.prometheus.client.Metric.histogram:type_name -> io.prometheus.client.Histogram
//	0,  // 14: io.prometheus.client.ProtoMessage.type:type_name -> io.prometheus.client.MetricType
//	11, // 15: io.prometheus.client.ProtoMessage.metric:type_name -> io.prometheus.client.Metric
//	16, // [16:16] is the sub-list for method output_type
//	16, // [16:16] is the sub-list for method input_type
//	16, // [16:16] is the sub-list for extension type_name
//	16, // [16:16] is the sub-list for extension extendee
//	0,  // [0:16] is the sub-list for field type_name
//}
//
//func init() { file_io_prometheus_client_metrics_proto_init() }
//func file_io_prometheus_client_metrics_proto_init() {
//	if File_io_prometheus_client_metrics_proto != nil {
//		return
//	}
//	if !protoimpl.UnsafeEnabled {
//		file_io_prometheus_client_metrics_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
//			switch v := v.(*ProtoMessage); i {
//			case 0:
//				return &v.state
//			case 1:
//				return &v.sizeCache
//			case 2:
//				return &v.unknownFields
//			default:
//				return nil
//			}
//		}
//	}
//	type x struct{}
//	out := protoimpl.TypeBuilder{
//		File: protoimpl.DescBuilder{
//			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
//			RawDescriptor: file_io_prometheus_client_metrics_proto_rawDesc,
//			NumEnums:      1,
//			NumMessages:   12,
//			NumExtensions: 0,
//			NumServices:   0,
//		},
//		GoTypes:           file_io_prometheus_client_metrics_proto_goTypes,
//		DependencyIndexes: file_io_prometheus_client_metrics_proto_depIdxs,
//		EnumInfos:         file_io_prometheus_client_metrics_proto_enumTypes,
//		MessageInfos:      file_io_prometheus_client_metrics_proto_msgTypes,
//	}.Build()
//	File_io_prometheus_client_metrics_proto = out.File
//	file_io_prometheus_client_metrics_proto_rawDesc = nil
//	file_io_prometheus_client_metrics_proto_goTypes = nil
//	file_io_prometheus_client_metrics_proto_depIdxs = nil
//}