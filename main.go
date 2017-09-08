// Copyright 2017 Johan Brandhorst. All Rights Reserved.
// See LICENSE for licensing terms.

package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"net/http"
	"strings"
	"time"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"

	"github.com/johanbrandhorst/grpcweb-example/client/compiled"
	"github.com/johanbrandhorst/grpcweb-example/server"
	"github.com/johanbrandhorst/grpcweb-example/server/proto/library"
	"github.com/johanbrandhorst/protobuf/wsproxy"
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
	wrappedServer := grpcweb.WrapServer(gs)

	var clientCreds credentials.TransportCredentials
	if *host == "" {
		var err error
		clientCreds, err = credentials.NewClientTLSFromFile("./insecure/localhost.crt", "localhost:10000")
		if err != nil {
			logger.Fatalln("Failed to get local server client credentials:", err)
		}
	} else {
		cp, err := x509.SystemCertPool()
		if err != nil {
			logger.Fatalln("Failed to get local system certpool:", err)
		}
		clientCreds = credentials.NewTLS(&tls.Config{RootCAs: cp})
	}

	wsproxy := wsproxy.WrapServer(
		http.HandlerFunc(wrappedServer.ServeHTTP),
		wsproxy.WithLogger(logger),
		wsproxy.WithTransportCredentials(clientCreds))

	handler := func(resp http.ResponseWriter, req *http.Request) {
		// Redirect gRPC and gRPC-Web requests to the gRPC Server
		if req.ProtoMajor == 2 && strings.Contains(req.Header.Get("Content-Type"), "application/grpc") ||
			websocket.IsWebSocketUpgrade(req) {
			wsproxy.ServeHTTP(resp, req)
		} else {
			// Serve the GopherJS client
			http.FileServer(&assetfs.AssetFS{
				Asset:     compiled.Asset,
				AssetDir:  compiled.AssetDir,
				AssetInfo: compiled.AssetInfo,
			}).ServeHTTP(resp, req)
		}
	}

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
		Handler: hstsHandler(handler),
	}

	// Serve on localhost with localhost certs if no host provided
	if *host == "" {
		httpsSrv.Addr = "localhost:10000"
		logger.Info("Serving on https://localhost:10000")
		logger.Fatal(httpsSrv.ListenAndServeTLS("./insecure/localhost.crt", "./insecure/localhost.key"))
	}

	// Create server for redirecting HTTP to HTTPS
	httpSrv := &http.Server{
		ReadTimeout:  httpsSrv.ReadTimeout,
		WriteTimeout: httpsSrv.WriteTimeout,
		IdleTimeout:  httpsSrv.IdleTimeout,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Connection", "close")
			url := "https://" + req.Host + req.URL.String()
			http.Redirect(w, req, url, http.StatusMovedPermanently)
		}),
	}
	go func() {
		logger.Fatal(httpSrv.ListenAndServe())
	}()

	// Create auto-certificate https server
	m := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(*host),
		Cache:      autocert.DirCache("/certs"),
	}
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
