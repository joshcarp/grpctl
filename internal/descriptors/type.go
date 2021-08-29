package descriptors

import (
	"encoding/json"
	"fmt"
	"strconv"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type DataValue struct {
	Kind  protoreflect.Kind
	Value interface{}
}

type DataMap map[string]*DataValue

func (d DataMap) ToJson() ([]byte, error) {
	jsonVal := map[string]interface{}{}
	for key, val := range d {
		jsonVal[key] = val.Value
	}
	return json.Marshal(jsonVal)
}

func (v *DataValue) String() string {
	return fmt.Sprintf("%v", v.Value)
}

func (v *DataValue) Set(val string) error {
	var err error
	switch v.Kind {
	case protoreflect.BoolKind:
		v.Value = val == "true"
	case protoreflect.EnumKind:
		v.Value, err = strconv.ParseInt(val, 10, 64)
	case protoreflect.Int32Kind:
		v.Value, err = strconv.ParseInt(val, 10, 32)
	case protoreflect.Sint32Kind:
		v.Value, err = strconv.ParseInt(val, 10, 32)
	case protoreflect.Uint32Kind:
		v.Value, err = strconv.ParseInt(val, 10, 32)
	case protoreflect.Int64Kind:
		v.Value, err = strconv.ParseInt(val, 10, 64)
	case protoreflect.Sint64Kind:
		v.Value, err = strconv.ParseInt(val, 10, 64)
	case protoreflect.Uint64Kind:
		v.Value, err = strconv.ParseInt(val, 10, 64)
	case protoreflect.Sfixed32Kind:
		v.Value, err = strconv.ParseInt(val, 10, 32)
	case protoreflect.Fixed32Kind:
		v.Value, err = strconv.ParseInt(val, 10, 32)
	case protoreflect.FloatKind:
		v.Value, err = strconv.ParseFloat(val, 64)
	case protoreflect.Sfixed64Kind:
		v.Value, err = strconv.ParseInt(val, 10, 64)
	case protoreflect.Fixed64Kind:
		v.Value, err = strconv.ParseInt(val, 10, 64)
	case protoreflect.DoubleKind:
		v.Value, err = strconv.ParseFloat(val, 64)
	case protoreflect.StringKind:
		v.Value = val
	case protoreflect.BytesKind:
		v.Value = val
	}
	return err
}

func (v *DataValue) Type() string {
	return v.Kind.String()
}
