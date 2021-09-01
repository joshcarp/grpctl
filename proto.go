package grpctl

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

// Adapted from https://github.com/fullstorydev/grpcurl/blob/de25c898228e36e8539862ed08de69598e64cb76/grpcurl.go#L400
func MakeJsonTemplate(md protoreflect.MessageDescriptor) (map[string]interface{}, string) {
	toString, err := protojson.Marshal(makeTemplate(md, nil))
	if err != nil {
		return nil, ""
	}
	var m map[string]interface{}
	err = json.Unmarshal(toString, &m)
	if err != nil {
		return nil, ""
	}
	return m, string(toString)
}

func makeTemplate(md protoreflect.MessageDescriptor, path []protoreflect.MessageDescriptor) proto.Message {
	switch md.FullName() {
	case "google.protobuf.Any":
		var any anypb.Any
		_ = anypb.MarshalFrom(&any, &emptypb.Empty{}, proto.MarshalOptions{})
		return &any
	case "google.protobuf.Value":
		return &structpb.Value{
			Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"google.protobuf.Value": {Kind: &structpb.Value_StringValue{
						StringValue: "supports arbitrary JSON",
					}},
				},
			}},
		}
	case "google.protobuf.ListValue":
		return &structpb.ListValue{
			Values: []*structpb.Value{
				{
					Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"google.protobuf.ListValue": {Kind: &structpb.Value_StringValue{
								StringValue: "is an array of arbitrary JSON values",
							}},
						},
					}},
				},
			},
		}
	case "google.protobuf.Struct":
		return &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"google.protobuf.Struct": {Kind: &structpb.Value_StringValue{
					StringValue: "supports arbitrary JSON objects",
				}},
			},
		}
	}
	dm := dynamicpb.NewMessage(md)
	for _, seen := range path {
		if seen == md {
			return dm
		}
	}
	for i := 0; i < dm.Descriptor().Fields().Len(); i++ {
		fd := dm.Descriptor().Fields().Get(i)
		var val protoreflect.Value
		switch fd.Kind() {
		case protoreflect.BoolKind:
			val = protoreflect.ValueOfBool(true)
		case protoreflect.EnumKind:
			val = protoreflect.ValueOfEnum(1)
		case protoreflect.Int32Kind:
			val = protoreflect.ValueOfInt32(1)
		case protoreflect.Sint32Kind:
			val = protoreflect.ValueOfInt32(1)
		case protoreflect.Uint32Kind:
			val = protoreflect.ValueOfInt32(1)
		case protoreflect.Int64Kind:
			val = protoreflect.ValueOfInt64(1)
		case protoreflect.Sint64Kind:
			val = protoreflect.ValueOfInt64(1)
		case protoreflect.Uint64Kind:
			val = protoreflect.ValueOfInt64(1)
		case protoreflect.Sfixed32Kind:
			val = protoreflect.ValueOfInt32(1)
		case protoreflect.Fixed32Kind:
			val = protoreflect.ValueOfFloat32(1.1)
		case protoreflect.FloatKind:
			val = protoreflect.ValueOfFloat32(1.1)
		case protoreflect.Sfixed64Kind:
			val = protoreflect.ValueOfInt64(1)
		case protoreflect.Fixed64Kind:
			val = protoreflect.ValueOfFloat64(1.1)
		case protoreflect.DoubleKind:
			val = protoreflect.ValueOfFloat64(1.1)
		case protoreflect.StringKind:
			val = protoreflect.ValueOfString("string")
		case protoreflect.BytesKind:
			val = protoreflect.ValueOfBytes([]byte(fd.JSONName()))
		case protoreflect.MessageKind:
			val = protoreflect.ValueOfMessage(makeTemplate(fd.Message(), nil).ProtoReflect())
		}
		if fd.Cardinality() == protoreflect.Repeated {
			val = protoreflect.ValueOfList(&List{vals: []protoreflect.Value{val}})
			continue
		}
		dm.Set(fd, val)

	}
	return dm
}

type List struct {
	vals []protoreflect.Value
}

// Len reports the number of entries in the List.
// Get, Set, and Truncate panic with out of bound indexes.
func (l *List) Len() int {
	return len(l.vals)
}

// Get retrieves the value at the given index.
// It never returns an invalid value.
func (l *List) Get(i int) protoreflect.Value {
	return l.vals[i]
}

// Set stores a value for the given index.
// When setting a composite type, it is unspecified whether the set
// value aliases the source's memory in any way.
//
// Set is a mutating operation and unsafe for concurrent use.
func (l *List) Set(i int, val protoreflect.Value) {
	l.vals[i] = val
}

// Append appends the provided value to the end of the list.
// When appending a composite type, it is unspecified whether the appended
// value aliases the source's memory in any way.
//
// Append is a mutating operation and unsafe for concurrent use.
func (l *List) Append(v protoreflect.Value) {
	l.vals = append(l.vals, v)
}

// AppendMutable appends a new, empty, mutable message value to the end
// of the list and returns it.
// It panics if the list does not contain a message type.
func (l *List) AppendMutable() protoreflect.Value {
	return protoreflect.ValueOfMessage(nil)
}

// Truncate truncates the list to a smaller length.
//
// Truncate is a mutating operation and unsafe for concurrent use.
func (l *List) Truncate(i int) {
	l.vals = l.vals[0:i]
}

// NewElement returns a new value for a list element.
// For enums, this returns the first enum value.
// For other scalars, this returns the zero value.
// For messages, this returns a new, empty, mutable value.
func (l *List) NewElement() protoreflect.Value {
	return protoreflect.ValueOfMessage(nil)
}

// IsValid reports whether the list is valid.
//
// An invalid list is an empty, read-only value.
//
// Validity is not part of the protobuf data model, and may not
// be preserved in marshaling or other operations.
func (l *List) IsValid() bool {
	return true
}
