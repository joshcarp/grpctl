package descriptors

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type DataValue struct {
	Kind  protoreflect.Kind `json:"-"`
	Proto bool              `json:"-"`
	Value interface{}       `json:"value"`
	Empty bool              `json:"-"`
}

type DataMap map[string]*DataValue

func (d DataMap) ToJson() ([]byte, error) {
	jsonVal := d.ToInterfaceMap()
	return json.Marshal(jsonVal)
}

func (d DataMap) ToInterfaceMap() map[string]interface{} {
	jsonVal := map[string]interface{}{}
	for key, val := range d {
		if val.Empty {
			continue
		}
		jsonVal[key] = val.Value
	}
	return jsonVal
}

func ToInterfaceMap(v interface{}) (map[string]interface{}, error) {
	marshal, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	m := map[string]interface{}{}
	err = json.Unmarshal(marshal, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (v *DataValue) String() string {
	return fmt.Sprintf("%v", v.Value)
}

func (v *DataValue) Set(val string) error {
	var err error
	if !v.Proto {
		switch reflect.TypeOf(v.Value).Kind() {
		case reflect.Map:
			vals := strings.Split(val, "=")
			if len(vals) != 2 {
				return fmt.Errorf("map type should be length of 2")
			}
			v.Value = map[string]interface{}{vals[0]: vals[1]}
			v.Empty = false
			return nil
		case reflect.Bool:
			v.Value, err = strconv.ParseBool(val)
			v.Empty = false
			return err
		}
		m, err := ToInterfaceMap(DataValue{Value: val})
		if err != nil {
			return nil
		}
		marshal, err := json.Marshal(m)
		if err != nil {
			return err
		}
		err = json.Unmarshal(marshal, &v)
		if err != nil {
			return err
		}
		v.Empty = false
		return nil
	}
	switch v.Kind {
	case protoreflect.BoolKind:
		v.Value, err = strconv.ParseBool(val)
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
