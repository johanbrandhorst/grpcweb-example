package main

import (
	"honnef.co/go/js/dom"
	r "myitcv.io/react"

	"github.com/johanbrandhorst/grpcweb-example/client/container"
)

//go:generate gopherjs build app.go -o html/client.js
//go:generate go-bindata -pkg compiled -nometadata -o compiled/client.go -prefix html ./html
//go:generate bash -c "rm html/*.js*"

func main() {
	domTarget := dom.GetWindow().Document().GetElementByID("app")

	r.Render(container.Container(), domTarget)
}
