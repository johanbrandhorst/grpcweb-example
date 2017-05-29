/*
	protoc-gen-gopherjs is a plugin for the Google protocol buffer compiler to generate
	GopherJS code.  Run it by building this program and putting it in your path with
	the name
		protoc-gen-gopherjs
	That word 'gopherjs' at the end becomes part of the option string set for the
	protocol compiler, so once the protocol compiler (protoc) is installed
	you can run
		protoc --gopherjs_out=output_directory input_directory/file.proto
	to generate GopherJS bindings for the protocol defined by file.proto.
	With that input, the output will be written to
		output_directory/file.pb.gopherjs.go

*/
package documentation
