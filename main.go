// Copyright 2017 Johan Brandhorst. All Rights Reserved.
// See LICENSE for licensing terms.

package main

import (
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/lpar/gzipped"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	"github.com/johanbrandhorst/grpcweb-example/client/compiled"
	"github.com/johanbrandhorst/grpcweb-example/server"
	"github.com/johanbrandhorst/grpcweb-example/server/proto/library"
)

var logger *logrus.Logger

func init() {
	logger = logrus.StandardLogger()
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		DisableSorting:   true,
	})
	// Should only be done from init functions
	grpclog.SetLogger(logger)
}

func main() {
	gs := grpc.NewServer()
	library.RegisterBookServiceServer(gs, &server.BookService{})
	wrappedServer := grpcweb.WrapServer(gs, grpcweb.WithWebsockets(true))

	httpSrv := &http.Server{
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
		Addr:              ":" + os.Getenv("PORT"),
		Handler: grpcTrafficSplitter(
			folderReader(
				gzipped.FileServer(compiled.Assets),
			),
			wrappedServer,
		),
	}

	logger.Info("Serving on http://0.0.0.0:" + os.Getenv("PORT"))
	logger.Fatal(httpSrv.ListenAndServe())
}

func folderReader(fn http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			// Use contents of index.html for directory, if present.
			r.URL.Path = path.Join(r.URL.Path, "index.html")
		}
		fn.ServeHTTP(w, r)
	})
}

func grpcTrafficSplitter(fallback http.Handler, grpcHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Redirect gRPC and gRPC-Web requests to the gRPC Server
		if strings.Contains(r.Header.Get("Content-Type"), "application/grpc") ||
			websocket.IsWebSocketUpgrade(r) {
			grpcHandler.ServeHTTP(w, r)
		} else {
			fallback.ServeHTTP(w, r)
		}
	})
}
