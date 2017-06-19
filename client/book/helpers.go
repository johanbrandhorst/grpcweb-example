package book

import (
	"strconv"
	"time"

	r "myitcv.io/react"

	"github.com/johanbrandhorst/grpcweb-example/client/proto/library"
)

func renderBook(bk *library.Book) r.Element {
	var publisher string
	switch bk.GetPublishingMethod().(type) {
	case *library.Book_Publisher:
		publisher = bk.GetPublisher().GetName()
	case *library.Book_SelfPublished:
		publisher = "Self-published"
	}
	return r.Div(nil,
		r.HR(nil),
		r.Div(nil,
			r.S("Title: "),
			r.Code(nil,
				r.S(bk.GetTitle()),
			),
		),
		r.Div(nil,
			r.S("Author: "),
			r.Code(nil,
				r.S(bk.GetAuthor()),
			),
		),
		r.Div(nil,
			r.S("Publisher: "),
			r.Code(nil,
				r.S(publisher),
			),
		),
		r.Div(nil,
			r.S("Publication date: "),
			r.Code(nil,
				r.S(time.Unix(
					bk.GetPublicationDate().GetSeconds(),
					int64(bk.GetPublicationDate().GetNanos()),
				).Format("Monday, 02 Jan 2006")),
			),
		),
		r.Div(nil,
			r.S("Book type: "),
			r.Code(nil,
				r.S(bk.GetBookType().String()),
			),
		),
		r.Div(nil,
			r.S("ISBN: "),
			r.Code(nil,
				r.S(strconv.Itoa(int(bk.GetIsbn()))),
			),
		),
	)
}
