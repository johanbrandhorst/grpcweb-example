package gopherjs

import (
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

// GetGopherJSPackage returns the (gopherjs.gopherjs_package) option if
// specified, or an empty string if it was not.
func GetGopherJSPackage(file *descriptor.FileDescriptorProto) string {
	if file == nil || file.Options == nil {
		return ""
	}

	e, err := proto.GetExtension(file.Options, E_GopherjsPackage)
	if err != nil {
		return ""
	}

	if s, ok := e.(*string); ok {
		return *s
	}

	return ""
}
