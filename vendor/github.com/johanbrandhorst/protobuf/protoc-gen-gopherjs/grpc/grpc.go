// Go support for Protocol Buffers - Google's data interchange format
//
// Copyright 2015 The Go Authors.  All rights reserved.
// https://github.com/golang/protobuf
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//     * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Package grpc outputs gRPC service descriptions in Go code.
// It runs as a plugin for the Go protocol buffer compiler plugin.
// It is linked in to protoc-gen-go.
package grpc

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	pb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/johanbrandhorst/protobuf/protoc-gen-gopherjs/generator"
)

// generatedCodeVersion indicates a version of the generated code.
// It is incremented whenever an incompatibility between the generated code and
// the grpcweb package is introduced; the generated code references
// a constant, grpcweb.GrpcWebPackageIsVersionN (where N is generatedCodeVersion).
const generatedCodeVersion = 3

// Paths for packages used by code generated in this file,
// relative to the import_prefix of the generator.Generator.
const (
	contextPkgPath = "context"
	grpcPkgPath    = "github.com/johanbrandhorst/protobuf/grpcweb"
)

func init() {
	generator.RegisterPlugin(new(grpc))
}

// grpc is an implementation of the Go protocol buffer compiler's
// plugin architecture.  It generates bindings for gRPC support.
type grpc struct {
	gen *generator.Generator
}

// Name returns the name of this plugin, "grpc".
func (g *grpc) Name() string {
	return "grpc"
}

// The names for packages imported in the generated code.
// They may vary from the final path component of the import path
// if the name is used by other packages.
var (
	contextPkg string
	grpcPkg    string
)

// Init initializes the plugin.
func (g *grpc) Init(gen *generator.Generator) {
	g.gen = gen
	contextPkg = generator.RegisterUniquePackageName("context", nil)
	grpcPkg = generator.RegisterUniquePackageName("grpcweb", nil)
}

// Given a type name defined in a .proto, return its object.
// Also record that we're using it, to guarantee the associated import.
func (g *grpc) objectNamed(name string) generator.Object {
	g.gen.RecordTypeUse(name)
	return g.gen.ObjectNamed(name)
}

// Given a type name defined in a .proto, return its name as we will print it.
func (g *grpc) typeName(str string) string {
	return g.gen.TypeName(g.objectNamed(str))
}

// P forwards to g.gen.P.
func (g *grpc) P(args ...interface{}) { g.gen.P(args...) }

// In forwards to g.gen.In.
func (g *grpc) In() { g.gen.In() }

// Out forwards to g.gen.Out.
func (g *grpc) Out() { g.gen.Out() }

// Generate generates code for the services in the given file.
func (g *grpc) Generate(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}

	g.P("// Reference imports to suppress errors if they are not otherwise used.")
	g.P("var _ ", contextPkg, ".Context")
	g.P("var _ ", grpcPkg, ".Client")
	g.P()

	// Assert version compatibility.
	g.P("// This is a compile-time assertion to ensure that this generated file")
	g.P("// is compatible with the grpcweb package it is being compiled against.")
	g.P("const _ = ", grpcPkg, ".GrpcWebPackageIsVersion", generatedCodeVersion)
	g.P()

	for i, service := range file.FileDescriptorProto.Service {
		g.generateService(file, service, i)
	}
}

// GenerateImports generates the import declaration for this file.
func (g *grpc) GenerateImports(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}
	g.P("import (")
	g.In()
	g.P(contextPkg, " ", strconv.Quote(path.Join(g.gen.ImportPrefix, contextPkgPath)))
	g.P()
	g.P(grpcPkg, " ", strconv.Quote(path.Join(g.gen.ImportPrefix, grpcPkgPath)))
	g.Out()
	g.P(")")
	g.P()
}

// reservedClientName records whether a client name is reserved on the client side.
var reservedClientName = map[string]bool{
	// TODO: do we need any in gRPC?
}

func unexport(s string) string { return strings.ToLower(s[:1]) + s[1:] }

// generateService generates all the code for the named service.
func (g *grpc) generateService(file *generator.FileDescriptor, service *pb.ServiceDescriptorProto, index int) {
	path := fmt.Sprintf("6,%d", index) // 6 means service.

	origServName := service.GetName()
	fullServName := origServName
	if pkg := file.GetPackage(); pkg != "" {
		fullServName = pkg + "." + fullServName
	}
	servName := generator.CamelCase(origServName)

	g.P()
	g.P("// Client API for ", servName, " service")
	g.P()

	// Client interface.
	g.gen.PrintComments(path)
	g.P("type ", servName, "Client interface {")
	g.In()
	for i, method := range service.Method {
		g.gen.PrintComments(fmt.Sprintf("%s,2,%d", path, i)) // 2 means method in a service.
		g.P(g.generateClientSignature(servName, method))
	}
	g.Out()
	g.P("}")
	g.P()

	// Client structure.
	g.P("type ", unexport(servName), "Client struct {")
	g.In()
	g.P("client *", grpcPkg, ".Client")
	g.Out()
	g.P("}")
	g.P()

	// NewClient factory.
	g.P("// New", servName, "Client creates a new gRPC-Web client.")
	g.P("func New", servName, "Client (hostname string, opts ...grpcweb.DialOption) ", servName, "Client {")
	g.In()
	g.P("return &", unexport(servName), "Client{")
	g.In()
	g.P("client: ", grpcPkg, `.NewClient(hostname, "`, fullServName, `", opts...),`)
	g.Out()
	g.P("}")
	g.Out()
	g.P("}")
	g.P()

	serviceDescVar := "_" + servName + "_serviceDesc"
	// Client method implementations.
	for _, method := range service.Method {
		g.generateClientMethod(servName, fullServName, serviceDescVar, method)
	}
}

// generateClientSignature returns the client-side signature for a method.
func (g *grpc) generateClientSignature(servName string, method *pb.MethodDescriptorProto) string {
	origMethName := method.GetName()
	methName := generator.CamelCase(origMethName)
	if reservedClientName[methName] {
		methName += "_"
	}
	reqArg := ", in *" + g.typeName(method.GetInputType())
	if method.GetClientStreaming() {
		reqArg = ""
	}
	respName := "*" + g.typeName(method.GetOutputType())
	if method.GetServerStreaming() || method.GetClientStreaming() {
		respName = servName + "_" + generator.CamelCase(origMethName) + "Client"
	}
	return fmt.Sprintf("%s(ctx %s.Context%s, opts ...%s.CallOption) (%s, error)", methName, contextPkg, reqArg, grpcPkg, respName)
}

func (g *grpc) generateClientMethod(servName, fullServName, serviceDescVar string, method *pb.MethodDescriptorProto) {
	methName := generator.CamelCase(method.GetName())
	outType := g.typeName(method.GetOutputType())
	inType := g.typeName(method.GetInputType())
	streamType := unexport(servName) + methName + "Client"

	g.P("func (c *", unexport(servName), "Client) ", g.generateClientSignature(servName, method), "{")
	g.In()
	switch {
	case !method.GetServerStreaming() && !method.GetClientStreaming():
		// Unary
		g.P(`resp, err := c.client.RPCCall(ctx, "`, method.GetName(), `", in.Marshal(), opts...)`)
		g.P("if err != nil {")
		g.In()
		g.P("return nil, err")
		g.Out()
		g.P("}")
		g.P()
		g.P("return new(", outType, ").Unmarshal(resp)")
		g.Out()
		g.P("}")
		g.P()
		return
	case method.GetServerStreaming() && !method.GetClientStreaming():
		// Server-side stream
		g.P(`srv, err := c.client.NewClientStream(ctx, false, true, "`, method.GetName(), `", opts...)`)
		g.P("if err != nil {")
		g.In()
		g.P("return nil, err")
		g.Out()
		g.P("}")
		g.P()
		g.P("err = srv.SendMsg(in.Marshal())")
	case method.GetClientStreaming() && !method.GetServerStreaming():
		// This case covers both client-side streaming and bidi streaming
		g.P(`srv, err := c.client.NewClientStream(ctx, true, false, "`, method.GetName(), `", opts...)`)
	case method.GetClientStreaming() && method.GetServerStreaming():
		g.P(`srv, err := c.client.NewClientStream(ctx, true, true, "`, method.GetName(), `", opts...)`)
	}

	g.P("if err != nil {")
	g.In()
	g.P("return nil, err")
	g.Out()
	g.P("}")
	g.P()
	g.P("return &", streamType, "{srv}, nil")
	g.Out()
	g.P("}")
	g.P()

	// Stream auxiliary types and methods.
	g.P("type ", servName, "_", methName, "Client interface {")
	g.In()
	if method.GetClientStreaming() {
		g.P("Send(*", inType, ") error")
	}
	if method.GetServerStreaming() {
		g.P("Recv() (*", outType, ", error)")
	}
	if method.GetClientStreaming() && !method.GetServerStreaming() {
		g.P("CloseAndRecv() (*", outType, ", error)")
	}
	g.P("grpcweb.ClientStream")
	g.Out()
	g.P("}")
	g.P()

	g.P("type ", streamType, " struct {")
	g.In()
	g.P("grpcweb.ClientStream")
	g.Out()
	g.P("}")
	g.P()

	if method.GetClientStreaming() {
		g.P("func (x *", streamType, ") Send(req *", inType, ") error {")
		g.In()
		g.P("return x.SendMsg(req.Marshal())")
		g.Out()
		g.P("}")
		g.P()
	}
	if method.GetServerStreaming() {
		g.P("func (x *", streamType, ") Recv() (*", outType, ", error) {")
		g.In()
		g.P("resp, err := x.RecvMsg()")
		g.P("if err != nil {")
		g.In()
		g.P("return nil, err")
		g.Out()
		g.P("}")
		g.P()
		g.P("return new(", outType, ").Unmarshal(resp)")
		g.Out()
		g.P("}")
		g.P()
	}
	if method.GetClientStreaming() && !method.GetServerStreaming() {
		g.P("func (x *", streamType, ") CloseAndRecv() (*", outType, ", error) {")
		g.In()
		g.P("err := x.CloseSend()")
		g.P("if err != nil {")
		g.In()
		g.P("return nil, err")
		g.Out()
		g.P("}")
		g.P()
		g.P("resp, err := x.RecvMsg()")
		g.P("if err != nil {")
		g.In()
		g.P("return nil, err")
		g.Out()
		g.P("}")
		g.P()
		g.P("return new(", outType, ").Unmarshal(resp)")
		g.Out()
		g.P("}")
		g.P()
	}
}
