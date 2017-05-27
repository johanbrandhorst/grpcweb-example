// Copyright 2017 Johan Brandhorst. All Rights Reserved.
// See LICENSE for licensing terms.

package server

import (
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"github.com/johanbrandhorst/gopherjs-improbable-grpc-web-example/server/proto/library"
)

type BookService struct{}

var books = []*library.Book{
	&library.Book{
		Isbn:   60929871,
		Title:  "Brave New World",
		Author: "Aldous Huxley",
	},
	&library.Book{
		Isbn:   140009728,
		Title:  "Nineteen Eighty-Four",
		Author: "George Orwell",
	},
	&library.Book{
		Isbn:   9780140301694,
		Title:  "Alice's Adventures in Wonderland",
		Author: "Lewis Carroll",
	},
	&library.Book{
		Isbn:   140008381,
		Title:  "Animal Farm",
		Author: "George Orwell",
	},
}

func (s *BookService) GetBook(ctx context.Context, bookQuery *library.GetBookRequest) (book *library.Book, err error) {
	grpc.SendHeader(ctx, metadata.Pairs("Pre-Response-Metadata", "Is-sent-as-headers-unary"))
	grpc.SetTrailer(ctx, metadata.Pairs("Post-Response-Metadata", "Is-sent-as-trailers-unary"))

	for _, bk := range books {
		if bk.Isbn == bookQuery.Isbn {
			book = bk
			return
		}
	}

	return nil, grpc.Errorf(codes.NotFound, "Book could not be found")
}

func (s *BookService) QueryBooks(bookQuery *library.QueryBooksRequest, stream library.BookService_QueryBooksServer) error {
	stream.SendHeader(metadata.Pairs("Pre-Response-Metadata", "Is-sent-as-headers-stream"))
	for _, book := range books {
		if strings.HasPrefix(book.Author, bookQuery.AuthorPrefix) {
			stream.Send(book)
		}
	}
	stream.SetTrailer(metadata.Pairs("Post-Response-Metadata", "Is-sent-as-trailers-stream"))
	return nil
}
