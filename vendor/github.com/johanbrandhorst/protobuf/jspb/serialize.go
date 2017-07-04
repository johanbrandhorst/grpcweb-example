package protobufjs

import (
	"github.com/gopherjs/gopherjs/js"
)

// ProtoMessage is implemented by... all *js.Objects.
// But it'll do for an interface to Serialize and Deserialize
type ProtoMessage interface {
	Call(string, ...interface{}) *js.Object
}

// Serialize marshals the provided ProtoMessage into
// a slice of bytes using the serializeBinary ProtobufJS function,
// returning an error if one was thrown.
func Serialize(m ProtoMessage) (resp []byte, err error) {
	// Recover any thrown JS errors
	defer func() {
		e := recover()
		if e == nil {
			return
		}

		if e, ok := e.(*js.Error); ok {
			err = e
		} else {
			panic(e)
		}
	}()

	return m.Call("serializeBinary").Interface().([]byte), err
}

// Deserialize unmarshals the provided ProtoMessage bytes into
// a generic *js.Object of the provided type,
// returning an error if one was thrown.
func Deserialize(m ProtoMessage, rawBytes []byte) (o *js.Object, err error) {
	// Recover any thrown JS errors
	defer func() {
		e := recover()
		if e == nil {
			return
		}

		if e, ok := e.(*js.Error); ok {
			err = e
		} else {
			panic(e)
		}
	}()

	return m.Call("deserializeBinary", rawBytes), err
}
