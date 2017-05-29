// Copyright 2017 Johan Brandhorst. All Rights Reserved.
// See LICENSE for licensing terms.

package main

//go:generate gopherjs build client.go -o html/index.js
//go:generate go-bindata -pkg compiled -nometadata -o compiled/client.go -prefix html ./html
//go:generate bash -c "rm html/*.js*"

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/johanbrandhorst/gopherjs-improbable-grpc-web-example/client/proto/library"
)

var document = js.Global.Get("document")
var baseURI = strings.TrimSuffix(document.Get("baseURI").String(), "/")

func printBook(b *library.Book) {
	var publisher string
	switch b.GetPublishingMethod().(type) {
	case *library.Book_Publisher:
		publisher = b.GetPublisher().GetName()
	case *library.Book_SelfPublished:
		publisher = "Self Published"
	default:
		println(b.GetPublishingMethod())
	}

	publishDate := time.Unix(b.GetPublicationDate().GetSeconds(),
		int64(b.GetPublicationDate().GetNanos()))

	println(b.GetAuthor(), b.GetIsbn(), b.GetTitle(),
		b.GetBookType().String(), publisher, publishDate.Format(time.RFC1123))
}

func main() {
	ctx := context.Background()
	client := library.NewBookServiceClient(baseURI)
	book, err := client.GetBook(ctx, library.NewGetBookRequest(140008381))
	if err != nil {
		println("Got request error:", err.Error())
		return
	}

	printBook(book)

	srv, err := client.QueryBooks(ctx, library.NewQueryBooksRequest("George"))
	if err != nil {
		println("Got request error:", err)
		return
	}

	for {
		book, err := srv.Recv()
		if err != nil {
			if err == io.EOF {
				// Success!
				break
			}

			println("Got request error:", err.Error())
			break
		}

		printBook(book)
	}

	srv, err = client.QueryBooks(ctx, library.NewQueryBooksRequest("Lisa"))
	if err != nil {
		println("Got request error:", err)
		return
	}

	for {
		book, err := srv.Recv()
		if err != nil {
			if err == io.EOF {
				// Success!
				break
			}

			println("Got request error:", err.Error())
			break
		}

		printBook(book)
	}
}
