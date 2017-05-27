// Copyright 2017 Johan Brandhorst. All Rights Reserved.
// See LICENSE for licensing terms.

package main

import (
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	"github.com/johanbrandhorst/gopherjs-improbable-grpc-web-example/client/compiled"
	"github.com/johanbrandhorst/gopherjs-improbable-grpc-web-example/server"
	"github.com/johanbrandhorst/gopherjs-improbable-grpc-web-example/server/proto/library"
)

var logger *logrus.Logger

// If you change this, you'll need to change the cert as well
const addr = "localhost:10000"

func init() {
	logger = logrus.StandardLogger()
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.Kitchen,
		DisableSorting:  true,
	})
	grpclog.SetLogger(logger)
}

func main() {
	gs := grpc.NewServer()
	library.RegisterBookServiceServer(gs, &server.BookService{})

	wrappedServer := grpcweb.WrapServer(gs)
	handler := func(resp http.ResponseWriter, req *http.Request) {
		if wrappedServer.IsGrpcWebRequest(req) {
			wrappedServer.ServeHttp(resp, req)
		} else {
			// Serve the GopherJS client
			http.FileServer(&assetfs.AssetFS{
				Asset:     compiled.Asset,
				AssetDir:  compiled.AssetDir,
				AssetInfo: compiled.AssetInfo,
			}).ServeHTTP(resp, req)
		}
	}
	httpServer := http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(handler),
	}

	logger.Warn("Serving on https://", addr)
	logger.Fatal(httpServer.ListenAndServeTLS("./insecure/localhost.crt", "./insecure/localhost.key"))
}
