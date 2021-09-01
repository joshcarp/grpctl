package grpctl

import (
	"reflect"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestMakeJsonTemplate(t *testing.T) {
	type args struct {
		md protoreflect.MessageDescriptor
	}
	tests := []struct {
		name  string
		args  args
		want  map[string]interface{}
		want1 string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := MakeJsonTemplate(tt.args.md)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeJsonTemplate() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("MakeJsonTemplate() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
