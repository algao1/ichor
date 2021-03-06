// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.6.1
// source: glucose.proto

package pb

import (
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
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

type Features struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Features []*Feature `protobuf:"bytes,1,rep,name=features,proto3" json:"features,omitempty"`
}

func (x *Features) Reset() {
	*x = Features{}
	if protoimpl.UnsafeEnabled {
		mi := &file_glucose_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Features) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Features) ProtoMessage() {}

func (x *Features) ProtoReflect() protoreflect.Message {
	mi := &file_glucose_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Features.ProtoReflect.Descriptor instead.
func (*Features) Descriptor() ([]byte, []int) {
	return file_glucose_proto_rawDescGZIP(), []int{0}
}

func (x *Features) GetFeatures() []*Feature {
	if x != nil {
		return x.Features
	}
	return nil
}

type Feature struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Time    *timestamp.Timestamp `protobuf:"bytes,1,opt,name=time,proto3" json:"time,omitempty"`
	Glucose float64              `protobuf:"fixed64,2,opt,name=glucose,proto3" json:"glucose,omitempty"`
	Insulin float64              `protobuf:"fixed64,3,opt,name=insulin,proto3" json:"insulin,omitempty"`
	Carbs   float64              `protobuf:"fixed64,4,opt,name=carbs,proto3" json:"carbs,omitempty"`
}

func (x *Feature) Reset() {
	*x = Feature{}
	if protoimpl.UnsafeEnabled {
		mi := &file_glucose_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Feature) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Feature) ProtoMessage() {}

func (x *Feature) ProtoReflect() protoreflect.Message {
	mi := &file_glucose_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Feature.ProtoReflect.Descriptor instead.
func (*Feature) Descriptor() ([]byte, []int) {
	return file_glucose_proto_rawDescGZIP(), []int{1}
}

func (x *Feature) GetTime() *timestamp.Timestamp {
	if x != nil {
		return x.Time
	}
	return nil
}

func (x *Feature) GetGlucose() float64 {
	if x != nil {
		return x.Glucose
	}
	return 0
}

func (x *Feature) GetInsulin() float64 {
	if x != nil {
		return x.Insulin
	}
	return 0
}

func (x *Feature) GetCarbs() float64 {
	if x != nil {
		return x.Carbs
	}
	return 0
}

type Label struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Time  *timestamp.Timestamp `protobuf:"bytes,1,opt,name=time,proto3" json:"time,omitempty"`
	Value float64              `protobuf:"fixed64,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *Label) Reset() {
	*x = Label{}
	if protoimpl.UnsafeEnabled {
		mi := &file_glucose_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Label) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Label) ProtoMessage() {}

func (x *Label) ProtoReflect() protoreflect.Message {
	mi := &file_glucose_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Label.ProtoReflect.Descriptor instead.
func (*Label) Descriptor() ([]byte, []int) {
	return file_glucose_proto_rawDescGZIP(), []int{2}
}

func (x *Label) GetTime() *timestamp.Timestamp {
	if x != nil {
		return x.Time
	}
	return nil
}

func (x *Label) GetValue() float64 {
	if x != nil {
		return x.Value
	}
	return 0
}

type Labels struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Labels []*Label `protobuf:"bytes,1,rep,name=labels,proto3" json:"labels,omitempty"`
}

func (x *Labels) Reset() {
	*x = Labels{}
	if protoimpl.UnsafeEnabled {
		mi := &file_glucose_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Labels) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Labels) ProtoMessage() {}

func (x *Labels) ProtoReflect() protoreflect.Message {
	mi := &file_glucose_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Labels.ProtoReflect.Descriptor instead.
func (*Labels) Descriptor() ([]byte, []int) {
	return file_glucose_proto_rawDescGZIP(), []int{3}
}

func (x *Labels) GetLabels() []*Label {
	if x != nil {
		return x.Labels
	}
	return nil
}

var File_glucose_proto protoreflect.FileDescriptor

var file_glucose_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x67, 0x6c, 0x75, 0x63, 0x6f, 0x73, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x36, 0x0a, 0x08, 0x46, 0x65, 0x61, 0x74, 0x75,
	0x72, 0x65, 0x73, 0x12, 0x2a, 0x0a, 0x08, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x46, 0x65,
	0x61, 0x74, 0x75, 0x72, 0x65, 0x52, 0x08, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73, 0x22,
	0x83, 0x01, 0x0a, 0x07, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x2e, 0x0a, 0x04, 0x74,
	0x69, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x67,
	0x6c, 0x75, 0x63, 0x6f, 0x73, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x01, 0x52, 0x07, 0x67, 0x6c,
	0x75, 0x63, 0x6f, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x69, 0x6e, 0x73, 0x75, 0x6c, 0x69, 0x6e,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x01, 0x52, 0x07, 0x69, 0x6e, 0x73, 0x75, 0x6c, 0x69, 0x6e, 0x12,
	0x14, 0x0a, 0x05, 0x63, 0x61, 0x72, 0x62, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x01, 0x52, 0x05,
	0x63, 0x61, 0x72, 0x62, 0x73, 0x22, 0x4d, 0x0a, 0x05, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x12, 0x2e,
	0x0a, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x14,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x01, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x22, 0x2e, 0x0a, 0x06, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x12, 0x24,
	0x0a, 0x06, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0c,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x52, 0x06, 0x6c, 0x61,
	0x62, 0x65, 0x6c, 0x73, 0x32, 0x36, 0x0a, 0x07, 0x47, 0x6c, 0x75, 0x63, 0x6f, 0x73, 0x65, 0x12,
	0x2b, 0x0a, 0x07, 0x50, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x12, 0x0f, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2e, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73, 0x1a, 0x0d, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2e, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x22, 0x00, 0x42, 0x1c, 0x5a, 0x1a,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x6c, 0x67, 0x61, 0x6f,
	0x31, 0x2f, 0x69, 0x63, 0x68, 0x6f, 0x72, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_glucose_proto_rawDescOnce sync.Once
	file_glucose_proto_rawDescData = file_glucose_proto_rawDesc
)

func file_glucose_proto_rawDescGZIP() []byte {
	file_glucose_proto_rawDescOnce.Do(func() {
		file_glucose_proto_rawDescData = protoimpl.X.CompressGZIP(file_glucose_proto_rawDescData)
	})
	return file_glucose_proto_rawDescData
}

var file_glucose_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_glucose_proto_goTypes = []interface{}{
	(*Features)(nil),            // 0: proto.Features
	(*Feature)(nil),             // 1: proto.Feature
	(*Label)(nil),               // 2: proto.Label
	(*Labels)(nil),              // 3: proto.Labels
	(*timestamp.Timestamp)(nil), // 4: google.protobuf.Timestamp
}
var file_glucose_proto_depIdxs = []int32{
	1, // 0: proto.Features.features:type_name -> proto.Feature
	4, // 1: proto.Feature.time:type_name -> google.protobuf.Timestamp
	4, // 2: proto.Label.time:type_name -> google.protobuf.Timestamp
	2, // 3: proto.Labels.labels:type_name -> proto.Label
	0, // 4: proto.Glucose.Predict:input_type -> proto.Features
	3, // 5: proto.Glucose.Predict:output_type -> proto.Labels
	5, // [5:6] is the sub-list for method output_type
	4, // [4:5] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_glucose_proto_init() }
func file_glucose_proto_init() {
	if File_glucose_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_glucose_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Features); i {
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
		file_glucose_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Feature); i {
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
		file_glucose_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Label); i {
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
		file_glucose_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Labels); i {
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
			RawDescriptor: file_glucose_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_glucose_proto_goTypes,
		DependencyIndexes: file_glucose_proto_depIdxs,
		MessageInfos:      file_glucose_proto_msgTypes,
	}.Build()
	File_glucose_proto = out.File
	file_glucose_proto_rawDesc = nil
	file_glucose_proto_goTypes = nil
	file_glucose_proto_depIdxs = nil
}
