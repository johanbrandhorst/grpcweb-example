package book

import (
	"context"
	"strconv"

	"github.com/johanbrandhorst/protobuf/grpcweb/status"
	"honnef.co/go/js/dom"
	r "myitcv.io/react"

	"github.com/johanbrandhorst/grpcweb-example/client/proto/library"
)

//go:generate reactGen

// MakeCollectionDef defines the MakeCollection component
type MakeCollectionDef struct {
	r.ComponentDef
}

// MakeCollectionProps defines the properties of this component
type MakeCollectionProps struct {
	Client library.BookServiceClient
}

// MakeCollectionState holds the state for the MakeCollection component
type MakeCollectionState struct {
	isbnInput  string
	numAdded   int
	collection *library.Collection
	client     library.BookService_MakeCollectionClient
	err        string
}

// MakeCollection returns a new MakeCollectionElem
func MakeCollection(p MakeCollectionProps) *MakeCollectionElem {
	return buildMakeCollectionElem(p)
}

// Render renders the MakeCollection component
func (g MakeCollectionDef) Render() r.Element {
	st := g.State()
	content := []r.Element{
		r.P(nil, r.S("Assemble a collection of Books")),
		r.Form(&r.FormProps{ClassName: "form-inline"},
			r.Div(
				&r.DivProps{ClassName: "form-group"},
				r.Label(&r.LabelProps{ClassName: "sr-only", For: "isnbText"}, r.S("ISBN")),
				r.Input(&r.InputProps{
					Type:        "number",
					ClassName:   "form-control",
					ID:          "isnbText",
					Value:       st.isbnInput,
					OnChange:    isbnInputChange2{g},
					Placeholder: "Book ISBN",
				}),
				r.Button(&r.ButtonProps{
					Type:      "submit",
					ClassName: "btn btn-default",
					OnClick:   triggerAdd{g},
				}, r.S("Add book")),
			),
		),
	}

	if st.numAdded > 0 {
		content = append(content, r.P(nil,
			r.S("Books added:"+strconv.Itoa(st.numAdded)),
		))
	}

	content = append(content, r.Div(
		nil,
		r.Button(&r.ButtonProps{
			Type:      "submit",
			ClassName: "btn btn-default",
			OnClick:   triggerCollect{g},
		}, r.S("Get collection")),
	))

	if st.collection != nil {
		for _, bk := range st.collection.GetBooks() {
			content = append(content,
				r.Div(nil,
					r.Hr(nil),
					r.Div(nil,
						r.S("ISBN: "),
						r.Code(nil,
							r.S(strconv.Itoa(int(bk.GetIsbn()))),
						),
					),
				),
			)
		}
	}

	if st.err != "" {
		content = append(content,
			r.Div(nil,
				r.Hr(nil),
				r.S("Error: "+st.err),
			),
		)
	}

	return r.Div(nil, content...)
}

type isbnInputChange2 struct{ g MakeCollectionDef }
type triggerAdd struct{ g MakeCollectionDef }
type triggerCollect struct{ g MakeCollectionDef }

func (i isbnInputChange2) OnChange(se *r.SyntheticEvent) {
	target := se.Target().(*dom.HTMLInputElement)

	newSt := i.g.State()
	newSt.isbnInput = target.Value

	i.g.SetState(newSt)
}

func (t triggerAdd) OnClick(se *r.SyntheticMouseEvent) {
	// Wrapped in goroutine because MakeCollection is blocking
	go func() {
		newSt := t.g.State()
		defer func() {
			t.g.SetState(newSt)
		}()
		newSt.err = ""
		newSt.collection = nil

		isbn, err := strconv.Atoi(newSt.isbnInput)
		if err != nil {
			newSt.err = "ISBN must not be empty"
			return
		}

		if newSt.client == nil {
			newSt.client, err = t.g.Props().Client.MakeCollection(context.Background())
			if err != nil {
				sts := status.FromError(err)
				newSt.err = sts.Error()
				return
			}
		}

		err = newSt.client.Send(&library.Book{
			Isbn: int64(isbn),
		})
		newSt.isbnInput = ""
		if err != nil {
			sts := status.FromError(err)
			newSt.err = sts.Error()
			newSt.client = nil
			newSt.numAdded = 0
			return
		}

		newSt.numAdded++
	}()

	se.PreventDefault()
}

func (t triggerCollect) OnClick(se *r.SyntheticMouseEvent) {
	// Wrapped in goroutine because CloseAndRecv is blocking
	go func() {
		newSt := t.g.State()
		defer func() {
			t.g.SetState(newSt)
		}()
		newSt.err = ""
		newSt.collection = nil

		if newSt.numAdded == 0 || newSt.client == nil {
			newSt.err = "Must add at least one book"
			return
		}

		var err error
		newSt.collection, err = newSt.client.CloseAndRecv()
		newSt.client = nil
		newSt.numAdded = 0
		if err != nil {
			sts := status.FromError(err)
			newSt.err = sts.Error()
			return
		}
	}()

	se.PreventDefault()
}
