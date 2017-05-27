// This is what protoc-gen-gopherjs should generate
package library

import (
	"context"

	"github.com/gopherjs/gopherjs/js"

	// gRPC-web Bindings
	grpcweb "github.com/johanbrandhorst/gopherjs-improbable-grpc-web"
)

var pkg = js.Global.Get("proto").Get("library")

type Book struct {
	*js.Object
}

func NewBook(isbn int64, title, author string) *Book {
	x := &Book{
		Object: pkg.Get("Book").New([]interface{}{
			isbn,
			title,
			author,
		}),
	}
	return x
}

func (x *Book) GetIsbn() int64 { return x.Call("getIsbn").Int64() }

func (x *Book) SetIsbn(v int64) { x.Call("setIsbn", v) }

func (x *Book) GetTitle() string { return x.Call("getTitle").String() }

func (x *Book) SetTitle(v string) { x.Call("setTitle", v) }

func (x *Book) GetAuthor() string { return x.Call("getAuthor").String() }

func (x *Book) SetAuthor(v string) { x.Call("setAuthor", v) }

func (x *Book) Serialize() (rawBytes []byte, err error) {
	return grpcweb.Serialize(x)
}

func DeserializeBook(rawBytes []byte) (x *Book, err error) {
	obj, err := grpcweb.Deserialize(pkg.Get("Book"), rawBytes)
	if err != nil {
		return nil, err
	}

	return &Book{
		Object: obj,
	}, nil
}

type GetBookRequest struct {
	*js.Object
}

func NewGetBookRequest(isbn int64) *GetBookRequest {
	x := &GetBookRequest{
		Object: pkg.Get("GetBookRequest").New([]interface{}{
			isbn,
		}),
	}
	return x
}

func (x *GetBookRequest) GetIsbn() int64 { return x.Call("getIsbn").Int64() }

func (x *GetBookRequest) SetIsbn(v int64) { x.Call("setIsbn", v) }

func (x *GetBookRequest) Serialize() (rawBytes []byte, err error) {
	return grpcweb.Serialize(x)
}

func DeserializeGetBookRequest(rawBytes []byte) (x *GetBookRequest, err error) {
	obj, err := grpcweb.Deserialize(pkg.Get("GetBookRequest"), rawBytes)
	if err != nil {
		return nil, err
	}

	return &GetBookRequest{
		Object: obj,
	}, nil
}

type QueryBooksRequest struct {
	*js.Object
	AuthorPrefix string `js:"authorPrefix"`
}

func NewQueryBooksRequest(authorPrefix string) *QueryBooksRequest {
	x := &QueryBooksRequest{
		Object: pkg.Get("QueryBooksRequest").New([]interface{}{
			authorPrefix,
		}),
	}

	return x
}

func (x *QueryBooksRequest) Serialize() (rawBytes []byte, err error) {
	return grpcweb.Serialize(x)
}

func DeserializeQueryBooksRequest(rawBytes []byte) (x *QueryBooksRequest, err error) {
	obj, err := grpcweb.Deserialize(pkg.Get("QueryBooksRequest"), rawBytes)
	if err != nil {
		return nil, err
	}

	return &QueryBooksRequest{
		Object: obj,
	}, nil
}

type BookServiceClient interface {
	GetBook(context.Context, *GetBookRequest, ...grpcweb.CallOption) (*Book, error)
	QueryBooks(context.Context, *QueryBooksRequest, ...grpcweb.CallOption) (BookService_QueryBooksClient, error)
}

type bookServiceClient struct {
	*js.Object
	client *grpcweb.Client
}

type BookService_QueryBooksClient interface {
	Recv() (*Book, error)
}

type bookServiceQueryBooksClient struct {
	stream *grpcweb.StreamClient
}

func (x *bookServiceQueryBooksClient) Recv() (*Book, error) {
	resp, err := x.stream.Recv()
	if err != nil {
		// Could be EOF, on success
		return nil, err
	}

	return DeserializeBook(resp)
}

func NewBookServiceClient(hostname string) BookServiceClient {
	bsc := &bookServiceClient{
		Object: js.Global.Get("Object").New(),
	}
	bsc.client = grpcweb.NewClient(hostname, "library.BookService")
	return bsc
}

func (x *bookServiceClient) GetBook(ctx context.Context, in *GetBookRequest, opts ...grpcweb.CallOption) (*Book, error) {
	req, err := in.Serialize()
	if err != nil {
		return nil, err
	}

	resp, err := x.client.RPCCall(ctx, "GetBook", req, opts...)
	if err != nil {
		return nil, err
	}

	return DeserializeBook(resp)
}

func (x *bookServiceClient) QueryBooks(ctx context.Context, in *QueryBooksRequest, opts ...grpcweb.CallOption) (BookService_QueryBooksClient, error) {
	req, err := in.Serialize()
	if err != nil {
		return nil, err
	}

	srv, err := x.client.Stream(ctx, "QueryBooks", req, opts...)
	if err != nil {
		return nil, err
	}

	return &bookServiceQueryBooksClient{
		stream: srv,
	}, nil
}
