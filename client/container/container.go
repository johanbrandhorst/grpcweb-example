package container

import (
	"strings"

	"honnef.co/go/js/dom"
	"honnef.co/go/js/xhr"
	"myitcv.io/highlightjs"
	r "myitcv.io/react"

	"github.com/johanbrandhorst/grpcweb-example/client/book"
	"github.com/johanbrandhorst/grpcweb-example/client/proto/library"
)

//go:generate reactGen

// ContainerDef describes the component that models the core of the frontend
type ContainerDef struct {
	r.ComponentDef
}

// ContainerState contains the state for the Container
type ContainerState struct {
	client   library.BookServiceClient
	examples *exampleSource
}

// Container creates the Container
func Container() *ContainerElem {
	return buildContainerElem()
}

// GetInitialState returns in the initial state for the ContainerDef component
func (p ContainerDef) GetInitialState() ContainerState {
	return ContainerState{
		client:   nil,
		examples: newExampleSource(),
	}
}

// ComponentWillMount is a React lifecycle method for the ContainerDef component.
// It populates the sources from the source code on github.
func (p ContainerDef) ComponentWillMount() {
	newSt := p.State()
	if !fetchStarted {
		for i, e := range sources.Range() {
			go func(i exampleKey, e *source) {
				req := xhr.NewRequest("GET", "https://raw.githubusercontent.com/johanbrandhorst/grpcweb-example/master/client/"+e.file())
				err := req.Send(nil)
				if err != nil {
					panic(err)
				}

				sources = sources.Set(i, e.setSrc(req.ResponseText))

				newSt.examples = sources
				p.SetState(newSt)
			}(i, e)
		}

		fetchStarted = true
	}

	newSt.client = library.NewBookServiceClient(
		strings.TrimSuffix(dom.GetWindow().Document().BaseURI(), "/"),
	)

	p.SetState(newSt)
}

func (p ContainerDef) renderExample(key exampleKey, title, msg, elem r.Element) r.Element {
	var goSrc string
	src, _ := p.State().examples.Get(key)
	if src != nil {
		goSrc = src.src()
	}

	code := r.NewDangerousInnerHTML(highlightjs.Highlight("go", goSrc, true).Value)

	return r.Div(nil,
		r.H3(nil, title),
		msg,
		r.Div(&r.DivProps{ClassName: "row"},
			r.Div(&r.DivProps{ClassName: "col-md-8"},
				r.Div(&r.DivProps{ClassName: "panel panel-default"},
					r.Div(&r.DivProps{ClassName: "panel-body"},
						r.Pre(&r.PreProps{
							Style: &r.CSS{
								MaxHeight: "400px",
							},
							DangerouslySetInnerHTML: code,
						}),
					),
				),
			),
			r.Div(&r.DivProps{ClassName: "col-md-4"},
				plainPanel(elem),
			),
		),
	)
}

// Render renders the Container
func (p ContainerDef) Render() r.Element {
	navbar := r.Nav(&r.NavProps{ClassName: "navbar navbar-inverse navbar-fixed-top"},
		r.Div(&r.DivProps{ClassName: "container"},
			r.Div(&r.DivProps{ClassName: "navbar-header"},
				r.A(&r.AProps{ClassName: "navbar-brand"},
					r.S("GopherJS gRPC-Web Showcase"),
				),
			),
		),
	)

	content := []r.Element{
		r.H3(nil, r.S("GopherJS gRPC-Web Client Examples")),
		r.P(nil,
			r.S("This page shows a couple of examples of using the "),
			r.A(&r.AProps{Href: "https://github.com/johanbrandhorst/protoc-gen-gopherjs", Target: "_blank"},
				r.Code(nil,
					r.S("protoc-gen-gopherjs"),
				),
			),
			r.S(" gRPC-Web client together with a "),
			r.A(&r.AProps{Href: "https://myitcv.io/react", Target: "_blank"},
				r.S("GopherJS React frontend"),
			),
			r.S("."),
		),
		r.P(nil,
			r.S("This page is heavily inspired by the "),
			r.A(&r.AProps{Href: "http://blog.myitcv.io/gopherjs_examples_sites/examplesshowcase", Target: "_blank"},
				r.S("GopherJS React Examples Showcase"),
			),
			r.S("."),
		),
		r.P(nil,
			r.S("For the source code, raising issues, questions etc, please see "),
			r.A(&r.AProps{Href: "https://github.com/johanbrandhorst/grpcweb-example", Target: "_blank"},
				r.S("the Github repo"),
			),
			r.S("."),
		),
		r.P(nil,
			r.S("Note the examples below show the Go source code from "),
			r.Code(nil, r.S("master")),
			r.S("."),
		),
	}

	content = append(content,
		p.renderExample(
			exampleGetBook,
			r.Span(nil, r.S("Getting a book from the library by ISBN")),
			r.P(nil, r.S("Sends a GetBook request to the gRPC server asking for the book with the given ISBN. "+
				"Renders the returned book (or error).")),
			book.GetBook(book.GetBookProps{Client: p.State().client}),
		),
		p.renderExample(
			exampleQueryBooks,
			r.Span(nil, r.S("Querying for books in the library by author")),
			r.P(nil, r.S("Sends a request to the gRPC backend asking for all books by authors whose names "+
				"start with the provided string. Renders the returned book (or error).")),
			book.QueryBooks(book.QueryBooksProps{Client: p.State().client}),
		),
	)

	return r.Div(nil,
		navbar,
		r.Div(&r.DivProps{ClassName: "container"},
			content...,
		),
	)
}

func plainPanel(children ...r.Element) r.Element {
	return r.Div(&r.DivProps{ClassName: "panel panel-default panel-body"},
		children...,
	)
}
