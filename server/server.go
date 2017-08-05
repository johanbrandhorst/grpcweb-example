// Copyright 2017 Johan Brandhorst. All Rights Reserved.
// See LICENSE for licensing terms.

package server

import (
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/johanbrandhorst/grpcweb-example/server/proto/library"
)

type BookService struct{}

var books = []*library.Book{
	&library.Book{
		Isbn:     60929871,
		Title:    "Brave New World",
		Author:   "Aldous Huxley",
		BookType: library.BookType_HARDCOVER,
		PublishingMethod: &library.Book_Publisher{
			Publisher: &library.Publisher{
				Name: "Chatto & Windus",
			},
		},
		PublicationDate: &timestamp.Timestamp{
			Seconds: time.Date(1932, time.January, 1, 0, 0, 0, 0, time.UTC).Unix(),
		},
	},
	&library.Book{
		Isbn:     140009728,
		Title:    "Nineteen Eighty-Four",
		Author:   "George Orwell",
		BookType: library.BookType_PAPERBACK,
		PublishingMethod: &library.Book_Publisher{
			Publisher: &library.Publisher{
				Name: "Secker & Warburg",
			},
		},
		PublicationDate: &timestamp.Timestamp{
			Seconds: time.Date(1949, time.June, 8, 0, 0, 0, 0, time.UTC).Unix(),
		},
	},
	&library.Book{
		Isbn:     9780140301694,
		Title:    "Alice's Adventures in Wonderland",
		Author:   "Lewis Carroll",
		BookType: library.BookType_AUDIOBOOK,
		PublishingMethod: &library.Book_Publisher{
			Publisher: &library.Publisher{
				Name: "Macmillan",
			},
		},
		PublicationDate: &timestamp.Timestamp{
			Seconds: time.Date(1865, time.November, 26, 0, 0, 0, 0, time.UTC).Unix(),
		},
	},
	&library.Book{
		Isbn:     140008381,
		Title:    "Animal Farm",
		Author:   "George Orwell",
		BookType: library.BookType_HARDCOVER,
		PublishingMethod: &library.Book_Publisher{
			Publisher: &library.Publisher{
				Name: "Secker & Warburg",
			},
		},
		PublicationDate: &timestamp.Timestamp{
			Seconds: time.Date(1945, time.August, 17, 0, 0, 0, 0, time.UTC).Unix(),
		},
	},
	&library.Book{
		Isbn:     1501107739,
		Title:    "Still Alice",
		Author:   "Lisa Genova",
		BookType: library.BookType_PAPERBACK,
		PublishingMethod: &library.Book_SelfPublished{
			SelfPublished: true,
		},
		PublicationDate: &timestamp.Timestamp{
			Seconds: time.Date(2007, time.January, 1, 0, 0, 0, 0, time.UTC).Unix(),
		},
	},
}

func (s *BookService) GetBook(ctx context.Context, bookQuery *library.GetBookRequest) (book *library.Book, err error) {
	for _, bk := range books {
		if bk.Isbn == bookQuery.Isbn {
			book = bk
			return
		}
	}

	return nil, grpc.Errorf(codes.NotFound, "Book could not be found")
}

func (s *BookService) QueryBooks(bookQuery *library.QueryBooksRequest, stream library.BookService_QueryBooksServer) error {
	for _, book := range books {
		if strings.HasPrefix(book.Author, bookQuery.AuthorPrefix) {
			stream.Send(book)
		}
	}
	return nil
}
