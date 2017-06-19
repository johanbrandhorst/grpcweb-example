package container

//go:generate immutableGen

type exampleKey int

const (
	exampleGetBook exampleKey = iota
	exampleQueryBooks
)

// _Imm_source generates a *source immutable type
// for storing information about a source file
type _Imm_source struct {
	file string
	src  string
}

// _Imm_exampleSource generates an immutable map type
// for mapping from the ExampleKey enum to a source type
type _Imm_exampleSource map[exampleKey]*source

// sources is a package scope variable of the immutable
// example source map type.
var sources = newExampleSource(func(es *exampleSource) {
	es.Set(exampleGetBook, new(source).setFile("book/getbook.go"))
	es.Set(exampleQueryBooks, new(source).setFile("book/querybooks.go"))
})

// fetchStarted tracks whether the downloading of the source files has
// already been triggered
var fetchStarted bool
