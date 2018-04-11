// Copyright 2017 Johan Brandhorst. All Rights Reserved.
// See LICENSE for licensing terms.

package main

import (
	"crypto/tls"
	"flag"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/lpar/gzipped"
	"github.com/sirupsen/logrus"
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
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
		DisableSorting:  true,
	})
	// Should only be done from init functions
	grpclog.SetLogger(logger)
}

func main() {
	flag.Parse()

	gs := grpc.NewServer()
	library.RegisterBookServiceServer(gs, &server.BookService{})
	wrappedServer := grpcweb.WrapServer(gs, grpcweb.WithWebsockets(true))

	httpsSrv := &http.Server{
		// These interfere with websocket streams, disable for now
		// ReadTimeout: 5 * time.Second,
		// WriteTimeout: 10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
		Addr:              ":https",
		TLSConfig: &tls.Config{
			PreferServerCipherSuites: true,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},
		},
		Handler: hstsHandler(
			grpcTrafficSplitter(
				folderReader(
					gzipped.FileServer(compiled.Assets).ServeHTTP,
				),
				wrappedServer,
			),
		),
	}

	// Serve on localhost with localhost certs if no host provided
	if *host == "" {
		httpsSrv.Addr = "localhost:10000"
		logger.Info("Serving on https://localhost:10000")
		logger.Fatal(httpsSrv.ListenAndServeTLS("./insecure/cert.pem", "./insecure/key.pem"))
	}

	// Create auto-certificate https server
	m := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(*host),
		Cache:      autocert.DirCache("/certs"),
	}

	// Create server for redirecting HTTP to HTTPS
	httpSrv := &http.Server{
		Addr:         ":http",
		ReadTimeout:  httpsSrv.ReadTimeout,
		WriteTimeout: httpsSrv.WriteTimeout,
		IdleTimeout:  httpsSrv.IdleTimeout,
		Handler:      m.HTTPHandler(nil),
	}
	go func() {
		logger.Fatal(httpSrv.ListenAndServe())
	}()

	httpsSrv.TLSConfig.GetCertificate = m.GetCertificate
	logger.Info("Serving on https://0.0.0.0:443, authenticating for https://", *host)
	logger.Fatal(httpsSrv.ListenAndServeTLS("", ""))
}

// hstsHandler wraps an http.HandlerFunc such that it sets the HSTS header.
func hstsHandler(fn http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		fn(w, r)
	})
}

func folderReader(fn http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			// Use contents of index.html for directory, if present.
			r.URL.Path = path.Join(r.URL.Path, "index.html")
		}
		fn(w, r)
	})
}

func grpcTrafficSplitter(fallback http.HandlerFunc, grpcHandler http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Redirect gRPC and gRPC-Web requests to the gRPC Server
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") ||
			websocket.IsWebSocketUpgrade(r) {
			grpcHandler.ServeHTTP(w, r)
		} else {
			fallback(w, r)
		}
	})
}
