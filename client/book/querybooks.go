package book

import (
	"context"
	"io"
	"time"

	"honnef.co/go/js/dom"
	r "myitcv.io/react"

	"github.com/johanbrandhorst/grpcweb-example/client/proto/library"
	"github.com/johanbrandhorst/grpcweb/status"
)

//go:generate reactGen
//go:generate immutableGen

// QueryBooksDef defines the QueryBooks component
type QueryBooksDef struct {
	r.ComponentDef
	client library.BookServiceClient
}

// _Imm_books is generated to an immutable
// type *books which we use in the state
type _Imm_books []*library.Book

type QueryBooksState struct {
	authorInput string
	books       *books
	err         string
}

func (q *QueryBooksDef) GetInitialState() QueryBooksState {
	return QueryBooksState{
		books: newBooks(),
	}
}

func QueryBooks(client library.BookServiceClient) *QueryBooksDef {
	res := &QueryBooksDef{
		client: client,
	}
	r.BlessElement(res, nil)

	return res
}

func (q *QueryBooksDef) Render() r.Element {
	st := q.State()
	content := []r.Element{
		r.P(nil, r.S("Search for books by author name prefix (for example, George).")),
		r.Form(&r.FormProps{ClassName: "form-inline"},
			r.Div(
				&r.DivProps{ClassName: "form-group"},
				r.Label(&r.LabelProps{ClassName: "sr-only", For: "authorInput"}, r.S("Author")),
				r.Input(&r.InputProps{
					Type:      "text",
					ClassName: "form-control",
					ID:        "authorInput",
					Value:     st.authorInput,
					OnChange:  authorInputChange{q},
				}),
				r.Button(&r.ButtonProps{
					Type:      "submit",
					ClassName: "btn btn-default",
					OnClick:   triggerQuery{q},
				}, r.S("Find books")),
			),
		),
	}

	if st.books.Len() != 0 {
		for _, bk := range st.books.Range() {
			content = append(content, renderBook(bk))
		}
	}

	if st.err != "" {
		content = append(content,
			r.Div(nil,
				r.HR(nil),
				r.S("Error: "+st.err),
			),
		)
	}

	return r.Div(nil, content...)
}

type authorInputChange struct{ q *QueryBooksDef }
type triggerQuery struct{ q *QueryBooksDef }

func (a authorInputChange) OnChange(se *r.SyntheticEvent) {
	target := se.Target().(*dom.HTMLInputElement)

	newSt := a.q.State()
	newSt.authorInput = target.Value

	a.q.SetState(newSt)
}

func (t triggerQuery) OnClick(se *r.SyntheticMouseEvent) {
	// Wrapped in goroutine because Recv is blocking
	go func() {
		newSt := t.q.State()
		newSt.err = ""
		newSt.books = newBooks()

		// 10 second timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		srv, err := t.q.client.QueryBooks(ctx, new(library.QueryBooksRequest).New(newSt.authorInput))
		if err != nil {
			sts := status.FromError(err)
			newSt.err = sts.Message
			t.q.SetState(newSt)
			return
		}

		for {
			// Blocks until book received
			bk, err := srv.Recv()
			if err != nil {
				if err == io.EOF {
					// Success!
					if newSt.books.Len() == 0 {
						newSt.err = "No books found for that author"
						t.q.SetState(newSt)
					}

					return
				}
				sts := status.FromError(err)
				newSt.err = sts.Message
				t.q.SetState(newSt)
				return
			}

			newSt.books = newSt.books.Append(bk)
			t.q.SetState(newSt)
		}
	}()

	se.PreventDefault()
}
