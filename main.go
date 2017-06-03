// Copyright 2017 Johan Brandhorst. All Rights Reserved.
// See LICENSE for licensing terms.

package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	"github.com/johanbrandhorst/grpcweb-example/client/compiled"
	"github.com/johanbrandhorst/grpcweb-example/server"
	"github.com/johanbrandhorst/grpcweb-example/server/proto/library"
)

var logger *logrus.Logger
var host = flag.String("host", "", "host to get LetsEncrypt certificate for")

func init() {
	logger = logrus.StandardLogger()
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.Kitchen,
		DisableSorting:  true,
	})
	// Should only be done from init functions
	grpclog.SetLogger(logger)
}

func main() {
	flag.Parse()

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

	// Serve on localhost with localhost certs if no host provided
	if *host == "" {
		httpServer := http.Server{
			Handler: http.HandlerFunc(handler),
			Addr:    "localhost:10000",
		}
		logger.Info("Serving on https://localhost:10000")
		logger.Fatal(httpServer.ListenAndServeTLS("./insecure/localhost.crt", "./insecure/localhost.key"))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	logger.Info("Serving on https://0.0.0.0:443, authenticating for https://", *host)
	logger.Fatal(http.Serve(autocert.NewListener(*host), mux))
}
