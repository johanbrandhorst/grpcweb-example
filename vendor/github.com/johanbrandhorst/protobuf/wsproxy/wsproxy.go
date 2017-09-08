package wsproxy

import (
	"context"
	"encoding/binary"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/transport"

	"github.com/johanbrandhorst/protobuf/internal"
)

// Logger is the interface used by the proxy to log events
type Logger interface {
	Debugln(...interface{})
	Warnln(...interface{})
}

type noopLogger struct{}

func (n noopLogger) Debugln(_ ...interface{}) {}
func (n noopLogger) Warnln(_ ...interface{})  {}

// proxy wraps a handler with a websocket to perform
// bidirectional messaging between a gRPC backend and a web frontend.
type proxy struct {
	h      http.Handler
	logger Logger
	creds  credentials.TransportCredentials
}

// WrapServer wraps the input handler with a Websocket-to-Bidi-Streaming proxy.
func WrapServer(h http.Handler, opts ...Option) http.Handler {
	p := &proxy{
		h:      h,
		logger: noopLogger{},
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Option specifies the type of function that can be used to configure the server.
type Option func(p *proxy)

// WithTransportCredentials specifies credentials to use for the transport.
func WithTransportCredentials(creds credentials.TransportCredentials) Option {
	return func(p *proxy) {
		p.creds = creds
	}
}

// WithLogger configures the proxy to use the logger for logging.
func WithLogger(logger Logger) Option {
	return func(p *proxy) {
		p.logger = logger
	}
}

// TODO: allow modification of upgrader settings?
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Enforce only local origins
		return true
	},
}

func isClosedConnError(err error) bool {
	str := err.Error()
	if strings.Contains(str, "use of closed network connection") {
		return true
	}
	return websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway)
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !websocket.IsWebSocketUpgrade(r) {
		p.h.ServeHTTP(w, r)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		p.logger.Warnln("Failed to upgrade Websocket:", err)
		return
	}

	defer func() {
		err = conn.Close()
		if err != nil {
			p.logger.Warnln("Failed to close connection:", err)
			return
		}
		p.logger.Debugln("Closed connection")
	}()

	ctx, cancelFn := context.WithCancel(r.Context())
	defer cancelFn()

	host := withPort(r.Host)
	p.logger.Debugln("Creating new transport with addr:", host)
	t, err := transport.NewClientTransport(ctx,
		transport.TargetInfo{Addr: host},
		transport.ConnectOptions{
			TransportCredentials: p.creds,
		})
	if err != nil {
		closeMsg := formatCloseMessage(websocket.CloseInternalServerErr, err.Error())
		_ = conn.WriteMessage(websocket.CloseMessage, closeMsg)
		p.logger.Warnln("Failed to create transport:", err)
		return
	}
	defer func() {
		err = t.GracefulClose()
		if err != nil {
			p.logger.Warnln("Failed to close transport:", err)
		}
	}()

	p.logger.Debugln("Creating new stream with host:", r.RemoteAddr, " and method:", r.RequestURI)
	s, err := t.NewStream(ctx, &transport.CallHdr{
		Host:   r.RemoteAddr,
		Method: r.RequestURI,
	})
	if err != nil {
		closeMsg := formatCloseMessage(websocket.CloseInternalServerErr, err.Error())
		_ = conn.WriteMessage(websocket.CloseMessage, closeMsg)
		p.logger.Warnln("Failed to create stream:", err)
		return
	}

	// Read loop - reads from websocket and puts it on the stream
	go func() {
		for {
			select {
			case <-s.Context().Done():
				p.logger.Debugln("[READ] Context canceled, returning")
				return
			default:
			}
			_, payload, err := conn.ReadMessage()
			if err != nil {
				cancelFn()
				if isClosedConnError(err) {
					p.logger.Warnln("[READ] Websocket closed")
					return
				}
				p.logger.Warnln("[READ] Failed to read Websocket message:", err)
				return
			}
			p.logger.Debugln("[READ] Read payload:", payload)
			if internal.IsCloseMessage(payload) {
				err = t.Write(s, nil, &transport.Options{Last: true})
				if err == io.EOF || err == nil {
					// Do not want to cancel context here, want
					// Writer to read io.EOF then exit.
					return
				}
			} else {
				// Append header
				payload = append(make([]byte, 5), payload...)
				// Skip first byte to indicate no compression
				// TODO: Add compression?
				// Encode size of payload to byte 1-4
				binary.BigEndian.PutUint32(payload[1:5], uint32(len(payload)-5))
				err = t.Write(s, payload, &transport.Options{Last: false})
			}

			if err != nil {
				cancelFn()
				p.logger.Warnln("[READ] Failed to write message to transport:", err)
				if _, ok := err.(transport.ConnectionError); !ok {
					t.CloseStream(s, err)
				}
				return
			}
		}
	}()

	// Write loop -- take messages from stream and write to websocket
	var header [5]byte
	var msg []byte
	for {
		// Read header
		_, err := s.Read(header[:])
		if err != nil {
			if err == io.EOF {
				p.logger.Debugln("[WRITE] Stream closed")
				// Wait for status to be received
				<-s.Done()
				p.sendStatus(conn, s.Status())
			} else {
				p.logger.Warnln("[WRITE] Failed to read header:", err)
				if se, ok := err.(transport.StreamError); ok {
					p.sendStatus(conn, status.New(se.Code, se.Desc))
				} else {
					p.sendStatus(conn, status.New(codes.Internal, err.Error()))
				}
			}

			return
		}

		// TODO: Add compression?
		isCompressed := uint8(header[0]) != 0
		if isCompressed {
			// If payload is compressed, bail out
			p.logger.Warnln("[WRITE] Reply was compressed, bailing")
			p.sendStatus(conn, status.New(codes.FailedPrecondition, "Server sent compressed data"))
			return
		}
		len := int(binary.BigEndian.Uint32(header[1:]))

		// TODO: Reuse buffer and resize as necessary instead
		msg = make([]byte, int(len))
		if n, err := s.Read(msg); err != nil || n != len {
			p.logger.Warnln("[WRITE] Failed to read message:", err)
			// Wait for status to be received
			<-s.Done()
			p.sendStatus(conn, s.Status())
			return
		}

		if err = conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
			p.logger.Warnln("[WRITE] Failed to write message:", err)
			return
		}
		p.logger.Debugln("[WRITE] Sent payload:", msg)
	}

}

func formatCloseMessage(code int, message string) []byte {
	closeMsg := websocket.FormatCloseMessage(code, message)
	if len(closeMsg) > 125 {
		t := []byte("[truncated]")
		closeMsg = append(closeMsg[:125-len(t)], t...)
	}
	return closeMsg
}

func (p *proxy) sendStatus(conn *websocket.Conn, st *status.Status) {
	p.logger.Debugln("[WRITE] Sending status: Msg:", st.Message(), ", Code:", st.Code().String())

	closeMsg := formatCloseMessage(internal.FormatErrorCode(st.Code()), st.Message())
	err := conn.WriteMessage(websocket.CloseMessage, closeMsg)
	if err != nil {
		p.logger.Warnln("[WRITE] Failed to write Websocket trailer:", err)
	}

	p.logger.Debugln("[WRITE] Sent close message")
	return
}

// withPort adds ":443" if another port isn't already present.
func withPort(host string) string {
	if _, _, err := net.SplitHostPort(host); err != nil {
		return net.JoinHostPort(host, "443")
	}
	return host
}
