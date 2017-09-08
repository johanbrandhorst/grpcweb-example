// Code generated by protoc-gen-gopherjs. DO NOT EDIT.
// source: proto/test/test.proto

/*
	Package test is a generated protocol buffer package.

	It is generated from these files:
		proto/test/test.proto

	It has these top-level messages:
		ExtraStuff
		PingRequest
		PingResponse
*/
package test

import jspb "github.com/johanbrandhorst/protobuf/jspb"
import google_protobuf "github.com/johanbrandhorst/protobuf/ptypes/empty"

import (
	context "context"

	grpcweb "github.com/johanbrandhorst/protobuf/grpcweb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the jspb package it is being compiled against.
const _ = jspb.JspbPackageIsVersion2

type PingRequest_FailureType int

const (
	PingRequest_NONE PingRequest_FailureType = 0
	PingRequest_CODE PingRequest_FailureType = 1
	PingRequest_DROP PingRequest_FailureType = 2
)

var PingRequest_FailureType_name = map[int]string{
	0: "NONE",
	1: "CODE",
	2: "DROP",
}
var PingRequest_FailureType_value = map[string]int{
	"NONE": 0,
	"CODE": 1,
	"DROP": 2,
}

func (x PingRequest_FailureType) String() string {
	return PingRequest_FailureType_name[int(x)]
}

type ExtraStuff struct {
	Addresses map[int32]string
	// Types that are valid to be assigned to Title:
	//	*ExtraStuff_FirstName
	//	*ExtraStuff_IdNumber
	Title       isExtraStuff_Title
	CardNumbers []uint32
}

// isExtraStuff_Title is used to distinguish types assignable to Title
type isExtraStuff_Title interface {
	isExtraStuff_Title()
}

// ExtraStuff_FirstName is assignable to Title
type ExtraStuff_FirstName struct {
	FirstName string
}

// ExtraStuff_IdNumber is assignable to Title
type ExtraStuff_IdNumber struct {
	IdNumber int32
}

func (*ExtraStuff_FirstName) isExtraStuff_Title() {}
func (*ExtraStuff_IdNumber) isExtraStuff_Title()  {}

// GetTitle gets the Title of the ExtraStuff.
func (m *ExtraStuff) GetTitle() (x isExtraStuff_Title) {
	if m == nil {
		return x
	}
	return m.Title
}

// GetAddresses gets the Addresses of the ExtraStuff.
func (m *ExtraStuff) GetAddresses() (x map[int32]string) {
	if m == nil {
		return x
	}
	return m.Addresses
}

// GetFirstName gets the FirstName of the ExtraStuff.
func (m *ExtraStuff) GetFirstName() (x string) {
	if v, ok := m.GetTitle().(*ExtraStuff_FirstName); ok {
		return v.FirstName
	}
	return x
}

// GetIdNumber gets the IdNumber of the ExtraStuff.
func (m *ExtraStuff) GetIdNumber() (x int32) {
	if v, ok := m.GetTitle().(*ExtraStuff_IdNumber); ok {
		return v.IdNumber
	}
	return x
}

// GetCardNumbers gets the CardNumbers of the ExtraStuff.
func (m *ExtraStuff) GetCardNumbers() (x []uint32) {
	if m == nil {
		return x
	}
	return m.CardNumbers
}

// MarshalToWriter marshals ExtraStuff to the provided writer.
func (m *ExtraStuff) MarshalToWriter(writer jspb.Writer) {
	if m == nil {
		return
	}

	switch t := m.Title.(type) {
	case *ExtraStuff_FirstName:
		if len(t.FirstName) > 0 {
			writer.WriteString(2, t.FirstName)
		}
	case *ExtraStuff_IdNumber:
		if t.IdNumber != 0 {
			writer.WriteInt32(3, t.IdNumber)
		}
	}

	if len(m.Addresses) > 0 {
		for key, value := range m.Addresses {
			writer.WriteMessage(1, func() {
				writer.WriteInt32(1, key)
				writer.WriteString(2, value)
			})
		}
	}

	if len(m.CardNumbers) > 0 {
		writer.WriteUint32Slice(4, m.CardNumbers)
	}

	return
}

// Marshal marshals ExtraStuff to a slice of bytes.
func (m *ExtraStuff) Marshal() []byte {
	writer := jspb.NewWriter()
	m.MarshalToWriter(writer)
	return writer.GetResult()
}

// UnmarshalFromReader unmarshals a ExtraStuff from the provided reader.
func (m *ExtraStuff) UnmarshalFromReader(reader jspb.Reader) *ExtraStuff {
	for reader.Next() {
		if m == nil {
			m = &ExtraStuff{}
		}

		switch reader.GetFieldNumber() {
		case 1:
			if m.Addresses == nil {
				m.Addresses = map[int32]string{}
			}
			reader.ReadMessage(func() {
				var key int32
				var value string
				for reader.Next() {
					switch reader.GetFieldNumber() {
					case 1:
						key = reader.ReadInt32()
					case 2:
						value = reader.ReadString()
					}
					m.Addresses[key] = value
				}
			})
		case 2:
			m.Title = &ExtraStuff_FirstName{
				FirstName: reader.ReadString(),
			}
		case 3:
			m.Title = &ExtraStuff_IdNumber{
				IdNumber: reader.ReadInt32(),
			}
		case 4:
			m.CardNumbers = reader.ReadUint32Slice()
		default:
			reader.SkipField()
		}
	}

	return m
}

// Unmarshal unmarshals a ExtraStuff from a slice of bytes.
func (m *ExtraStuff) Unmarshal(rawBytes []byte) (*ExtraStuff, error) {
	reader := jspb.NewReader(rawBytes)

	m = m.UnmarshalFromReader(reader)

	if err := reader.Err(); err != nil {
		return nil, err
	}

	return m, nil
}

type PingRequest struct {
	Value             string
	ResponseCount     int32
	ErrorCodeReturned uint32
	FailureType       PingRequest_FailureType
	CheckMetadata     bool
	SendHeaders       bool
	SendTrailers      bool
	MessageLatencyMs  int32
}

// GetValue gets the Value of the PingRequest.
func (m *PingRequest) GetValue() (x string) {
	if m == nil {
		return x
	}
	return m.Value
}

// GetResponseCount gets the ResponseCount of the PingRequest.
func (m *PingRequest) GetResponseCount() (x int32) {
	if m == nil {
		return x
	}
	return m.ResponseCount
}

// GetErrorCodeReturned gets the ErrorCodeReturned of the PingRequest.
func (m *PingRequest) GetErrorCodeReturned() (x uint32) {
	if m == nil {
		return x
	}
	return m.ErrorCodeReturned
}

// GetFailureType gets the FailureType of the PingRequest.
func (m *PingRequest) GetFailureType() (x PingRequest_FailureType) {
	if m == nil {
		return x
	}
	return m.FailureType
}

// GetCheckMetadata gets the CheckMetadata of the PingRequest.
func (m *PingRequest) GetCheckMetadata() (x bool) {
	if m == nil {
		return x
	}
	return m.CheckMetadata
}

// GetSendHeaders gets the SendHeaders of the PingRequest.
func (m *PingRequest) GetSendHeaders() (x bool) {
	if m == nil {
		return x
	}
	return m.SendHeaders
}

// GetSendTrailers gets the SendTrailers of the PingRequest.
func (m *PingRequest) GetSendTrailers() (x bool) {
	if m == nil {
		return x
	}
	return m.SendTrailers
}

// GetMessageLatencyMs gets the MessageLatencyMs of the PingRequest.
func (m *PingRequest) GetMessageLatencyMs() (x int32) {
	if m == nil {
		return x
	}
	return m.MessageLatencyMs
}

// MarshalToWriter marshals PingRequest to the provided writer.
func (m *PingRequest) MarshalToWriter(writer jspb.Writer) {
	if m == nil {
		return
	}

	if len(m.Value) > 0 {
		writer.WriteString(1, m.Value)
	}

	if m.ResponseCount != 0 {
		writer.WriteInt32(2, m.ResponseCount)
	}

	if m.ErrorCodeReturned != 0 {
		writer.WriteUint32(3, m.ErrorCodeReturned)
	}

	if int(m.FailureType) != 0 {
		writer.WriteEnum(4, int(m.FailureType))
	}

	if m.CheckMetadata {
		writer.WriteBool(5, m.CheckMetadata)
	}

	if m.SendHeaders {
		writer.WriteBool(6, m.SendHeaders)
	}

	if m.SendTrailers {
		writer.WriteBool(7, m.SendTrailers)
	}

	if m.MessageLatencyMs != 0 {
		writer.WriteInt32(8, m.MessageLatencyMs)
	}

	return
}

// Marshal marshals PingRequest to a slice of bytes.
func (m *PingRequest) Marshal() []byte {
	writer := jspb.NewWriter()
	m.MarshalToWriter(writer)
	return writer.GetResult()
}

// UnmarshalFromReader unmarshals a PingRequest from the provided reader.
func (m *PingRequest) UnmarshalFromReader(reader jspb.Reader) *PingRequest {
	for reader.Next() {
		if m == nil {
			m = &PingRequest{}
		}

		switch reader.GetFieldNumber() {
		case 1:
			m.Value = reader.ReadString()
		case 2:
			m.ResponseCount = reader.ReadInt32()
		case 3:
			m.ErrorCodeReturned = reader.ReadUint32()
		case 4:
			m.FailureType = PingRequest_FailureType(reader.ReadEnum())
		case 5:
			m.CheckMetadata = reader.ReadBool()
		case 6:
			m.SendHeaders = reader.ReadBool()
		case 7:
			m.SendTrailers = reader.ReadBool()
		case 8:
			m.MessageLatencyMs = reader.ReadInt32()
		default:
			reader.SkipField()
		}
	}

	return m
}

// Unmarshal unmarshals a PingRequest from a slice of bytes.
func (m *PingRequest) Unmarshal(rawBytes []byte) (*PingRequest, error) {
	reader := jspb.NewReader(rawBytes)

	m = m.UnmarshalFromReader(reader)

	if err := reader.Err(); err != nil {
		return nil, err
	}

	return m, nil
}

type PingResponse struct {
	Value   string
	Counter int32
}

// GetValue gets the Value of the PingResponse.
func (m *PingResponse) GetValue() (x string) {
	if m == nil {
		return x
	}
	return m.Value
}

// GetCounter gets the Counter of the PingResponse.
func (m *PingResponse) GetCounter() (x int32) {
	if m == nil {
		return x
	}
	return m.Counter
}

// MarshalToWriter marshals PingResponse to the provided writer.
func (m *PingResponse) MarshalToWriter(writer jspb.Writer) {
	if m == nil {
		return
	}

	if len(m.Value) > 0 {
		writer.WriteString(1, m.Value)
	}

	if m.Counter != 0 {
		writer.WriteInt32(2, m.Counter)
	}

	return
}

// Marshal marshals PingResponse to a slice of bytes.
func (m *PingResponse) Marshal() []byte {
	writer := jspb.NewWriter()
	m.MarshalToWriter(writer)
	return writer.GetResult()
}

// UnmarshalFromReader unmarshals a PingResponse from the provided reader.
func (m *PingResponse) UnmarshalFromReader(reader jspb.Reader) *PingResponse {
	for reader.Next() {
		if m == nil {
			m = &PingResponse{}
		}

		switch reader.GetFieldNumber() {
		case 1:
			m.Value = reader.ReadString()
		case 2:
			m.Counter = reader.ReadInt32()
		default:
			reader.SkipField()
		}
	}

	return m
}

// Unmarshal unmarshals a PingResponse from a slice of bytes.
func (m *PingResponse) Unmarshal(rawBytes []byte) (*PingResponse, error) {
	reader := jspb.NewReader(rawBytes)

	m = m.UnmarshalFromReader(reader)

	if err := reader.Err(); err != nil {
		return nil, err
	}

	return m, nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpcweb.Client

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpcweb package it is being compiled against.
const _ = grpcweb.GrpcWebPackageIsVersion2

// Client API for TestService service

type TestServiceClient interface {
	PingEmpty(ctx context.Context, in *google_protobuf.Empty, opts ...grpcweb.CallOption) (*PingResponse, error)
	Ping(ctx context.Context, in *PingRequest, opts ...grpcweb.CallOption) (*PingResponse, error)
	PingError(ctx context.Context, in *PingRequest, opts ...grpcweb.CallOption) (*google_protobuf.Empty, error)
	PingList(ctx context.Context, in *PingRequest, opts ...grpcweb.CallOption) (TestService_PingListClient, error)
	PingClientStream(ctx context.Context, opts ...grpcweb.CallOption) (TestService_PingClientStreamClient, error)
	PingClientStreamError(ctx context.Context, opts ...grpcweb.CallOption) (TestService_PingClientStreamErrorClient, error)
	PingBidiStream(ctx context.Context, opts ...grpcweb.CallOption) (TestService_PingBidiStreamClient, error)
	PingBidiStreamError(ctx context.Context, opts ...grpcweb.CallOption) (TestService_PingBidiStreamErrorClient, error)
}

type testServiceClient struct {
	client *grpcweb.Client
}

// NewTestServiceClient creates a new gRPC-Web client.
func NewTestServiceClient(hostname string, opts ...grpcweb.DialOption) TestServiceClient {
	return &testServiceClient{
		client: grpcweb.NewClient(hostname, "test.TestService", opts...),
	}
}

func (c *testServiceClient) PingEmpty(ctx context.Context, in *google_protobuf.Empty, opts ...grpcweb.CallOption) (*PingResponse, error) {
	resp, err := c.client.RPCCall(ctx, "PingEmpty", in.Marshal(), opts...)
	if err != nil {
		return nil, err
	}

	return new(PingResponse).Unmarshal(resp)
}

func (c *testServiceClient) Ping(ctx context.Context, in *PingRequest, opts ...grpcweb.CallOption) (*PingResponse, error) {
	resp, err := c.client.RPCCall(ctx, "Ping", in.Marshal(), opts...)
	if err != nil {
		return nil, err
	}

	return new(PingResponse).Unmarshal(resp)
}

func (c *testServiceClient) PingError(ctx context.Context, in *PingRequest, opts ...grpcweb.CallOption) (*google_protobuf.Empty, error) {
	resp, err := c.client.RPCCall(ctx, "PingError", in.Marshal(), opts...)
	if err != nil {
		return nil, err
	}

	return new(google_protobuf.Empty).Unmarshal(resp)
}

func (c *testServiceClient) PingList(ctx context.Context, in *PingRequest, opts ...grpcweb.CallOption) (TestService_PingListClient, error) {
	srv, err := c.client.NewServerStream(ctx, "PingList", in.Marshal(), opts...)
	if err != nil {
		return nil, err
	}

	return &testServicePingListClient{
		stream: srv,
	}, nil
}

type TestService_PingListClient interface {
	Recv() (*PingResponse, error)
	Context() context.Context
}

type testServicePingListClient struct {
	stream grpcweb.ServerStream
}

func (x *testServicePingListClient) Recv() (*PingResponse, error) {
	resp, err := x.stream.RecvMsg()
	if err != nil {
		return nil, err
	}

	return new(PingResponse).Unmarshal(resp)
}

func (x *testServicePingListClient) Context() context.Context {
	return x.stream.Context()
}

func (c *testServiceClient) PingClientStream(ctx context.Context, opts ...grpcweb.CallOption) (TestService_PingClientStreamClient, error) {
	srv, err := c.client.NewClientStream(ctx, "PingClientStream")
	if err != nil {
		return nil, err
	}

	return &testServicePingClientStreamClient{stream: srv}, nil
}

type TestService_PingClientStreamClient interface {
	Send(*PingRequest) error
	CloseAndRecv() (*PingResponse, error)
	Context() context.Context
}

type testServicePingClientStreamClient struct {
	stream grpcweb.ClientStream
}

func (x *testServicePingClientStreamClient) Send(req *PingRequest) error {
	return x.stream.SendMsg(req.Marshal())
}

func (x *testServicePingClientStreamClient) CloseAndRecv() (*PingResponse, error) {
	resp, err := x.stream.CloseAndRecv()
	if err != nil {
		return nil, err
	}

	return new(PingResponse).Unmarshal(resp)
}

func (x *testServicePingClientStreamClient) Context() context.Context {
	return x.stream.Context()
}

func (c *testServiceClient) PingClientStreamError(ctx context.Context, opts ...grpcweb.CallOption) (TestService_PingClientStreamErrorClient, error) {
	srv, err := c.client.NewClientStream(ctx, "PingClientStreamError")
	if err != nil {
		return nil, err
	}

	return &testServicePingClientStreamErrorClient{stream: srv}, nil
}

type TestService_PingClientStreamErrorClient interface {
	Send(*PingRequest) error
	CloseAndRecv() (*PingResponse, error)
	Context() context.Context
}

type testServicePingClientStreamErrorClient struct {
	stream grpcweb.ClientStream
}

func (x *testServicePingClientStreamErrorClient) Send(req *PingRequest) error {
	return x.stream.SendMsg(req.Marshal())
}

func (x *testServicePingClientStreamErrorClient) CloseAndRecv() (*PingResponse, error) {
	resp, err := x.stream.CloseAndRecv()
	if err != nil {
		return nil, err
	}

	return new(PingResponse).Unmarshal(resp)
}

func (x *testServicePingClientStreamErrorClient) Context() context.Context {
	return x.stream.Context()
}

func (c *testServiceClient) PingBidiStream(ctx context.Context, opts ...grpcweb.CallOption) (TestService_PingBidiStreamClient, error) {
	srv, err := c.client.NewClientStream(ctx, "PingBidiStream")
	if err != nil {
		return nil, err
	}

	return &testServicePingBidiStreamClient{stream: srv}, nil
}

type TestService_PingBidiStreamClient interface {
	Send(*PingRequest) error
	Recv() (*PingResponse, error)
	CloseSend() error
	Context() context.Context
}

type testServicePingBidiStreamClient struct {
	stream grpcweb.ClientStream
}

func (x *testServicePingBidiStreamClient) Send(req *PingRequest) error {
	return x.stream.SendMsg(req.Marshal())
}

func (x *testServicePingBidiStreamClient) Recv() (*PingResponse, error) {
	resp, err := x.stream.RecvMsg()
	if err != nil {
		return nil, err
	}

	return new(PingResponse).Unmarshal(resp)
}

func (x *testServicePingBidiStreamClient) CloseSend() error {
	return x.stream.CloseSend()
}

func (x *testServicePingBidiStreamClient) Context() context.Context {
	return x.stream.Context()
}

func (c *testServiceClient) PingBidiStreamError(ctx context.Context, opts ...grpcweb.CallOption) (TestService_PingBidiStreamErrorClient, error) {
	srv, err := c.client.NewClientStream(ctx, "PingBidiStreamError")
	if err != nil {
		return nil, err
	}

	return &testServicePingBidiStreamErrorClient{stream: srv}, nil
}

type TestService_PingBidiStreamErrorClient interface {
	Send(*PingRequest) error
	Recv() (*PingResponse, error)
	CloseSend() error
	Context() context.Context
}

type testServicePingBidiStreamErrorClient struct {
	stream grpcweb.ClientStream
}

func (x *testServicePingBidiStreamErrorClient) Send(req *PingRequest) error {
	return x.stream.SendMsg(req.Marshal())
}

func (x *testServicePingBidiStreamErrorClient) Recv() (*PingResponse, error) {
	resp, err := x.stream.RecvMsg()
	if err != nil {
		return nil, err
	}

	return new(PingResponse).Unmarshal(resp)
}

func (x *testServicePingBidiStreamErrorClient) CloseSend() error {
	return x.stream.CloseSend()
}

func (x *testServicePingBidiStreamErrorClient) Context() context.Context {
	return x.stream.Context()
}
