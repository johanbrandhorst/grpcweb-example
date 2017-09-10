// Copyright 2017 Johan Brandhorst. All Rights Reserved.
// See LICENSE for licensing terms.

package server

import (
	"io"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/johanbrandhorst/grpcweb-example/server/proto/library"
)

type BookService struct {
	b broadcaster
}

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

func (s *BookService) MakeCollection(srv library.BookService_MakeCollectionServer) error {
	collection := &library.Collection{}
	for {
		bk, err := srv.Recv()
		if err == io.EOF {
			return srv.SendAndClose(collection)
		}
		if err != nil {
			return err
		}

		collection.Books = append(collection.Books, bk)
	}
}

type broadcaster struct {
	listenerMu sync.RWMutex
	listeners  map[string]chan<- string
}

func (b *broadcaster) Add(name string, listener chan<- string) error {
	b.listenerMu.Lock()
	defer b.listenerMu.Unlock()
	if b.listeners == nil {
		b.listeners = map[string]chan<- string{}
	}
	if _, ok := b.listeners[name]; ok {
		return status.Errorf(codes.AlreadyExists, "The name %q is already in use by someone", name)
	}
	b.listeners[name] = listener
	return nil
}

func (b *broadcaster) Remove(name string) {
	b.listenerMu.Lock()
	defer b.listenerMu.Unlock()
	if c, ok := b.listeners[name]; ok {
		close(c)
		delete(b.listeners, name)
	}
}

func (b *broadcaster) Broadcast(ctx context.Context, msg string) {
	b.listenerMu.RLock()
	defer b.listenerMu.RUnlock()
	for _, listener := range b.listeners {
		select {
		case listener <- msg:
		case <-ctx.Done():
			return
		}
	}
}

func (s *BookService) BookChat(srv library.BookService_BookChatServer) error {
	// Listen for initial message with name
	msg, err := srv.Recv()
	if err == io.EOF {
		// Uhh... if you insist!
		return nil
	}
	if err != nil {
		return err
	}
	name := msg.GetName()
	if name == "" {
		return status.Error(codes.FailedPrecondition, "first message should be the name of the user")
	}

	// Send join message before user joins
	s.b.Broadcast(srv.Context(), name+" has joined the chat")

	listener := make(chan string)
	err = s.b.Add(name, listener)
	if err != nil {
		return err
	}
	defer func() {
		s.b.Remove(name)
		s.b.Broadcast(context.Background(), name+" has left the chat")
	}()

	sendErrChan := make(chan error)
	go func() {
		for {
			select {
			case msg, ok := <-listener:
				if !ok {
					// Listener is closed in broadcaster.Remove,
					// so this must mean the function has exited.
					return
				}
				err = srv.Send(&library.BookResponse{Message: msg})
				if err != nil {
					sendErrChan <- err
					return
				}
			case <-srv.Context().Done():
				return
			}
		}
	}()

	recvErrChan := make(chan error)
	go func() {
		for {
			msg, err := srv.Recv()
			if err == io.EOF {
				// Done
				close(recvErrChan)
				return
			}
			if err != nil {
				recvErrChan <- err
				return
			}
			s.b.Broadcast(srv.Context(), name+": "+msg.GetMessage())
		}
	}()

	select {
	case err, ok := <-recvErrChan:
		if !ok {
			// Success!
			return nil
		}
		return err
	case err := <-sendErrChan:
		return err
	case <-srv.Context().Done():
		return srv.Context().Err()
	}
}
