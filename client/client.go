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

	"github.com/gopherjs/gopherjs/js"
	"github.com/johanbrandhorst/gopherjs-improbable-grpc-web-example/client/proto/library"
)

var document = js.Global.Get("document")
var baseURI = strings.TrimSuffix(document.Get("baseURI").String(), "/")

func main() {
	ctx := context.Background()
	client := library.NewBookServiceClient(baseURI)
	book, err := client.GetBook(ctx, library.NewGetBookRequest(140008381))
	if err != nil {
		println("Got request error:", err.Error())
		return
	}

	println(book.GetAuthor(), book.GetIsbn(), book.GetTitle())

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
				return
			}

			println("Got request error:", err.Error())
			return
		}

		println(book.GetAuthor(), book.GetIsbn(), book.GetTitle())
	}
}
