package bidi

import (
	"context"
	"encoding/binary"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/transport"
)

type Proxy struct {
	h      http.Handler
	logger *logrus.Logger
}

func WrapServer(h http.Handler, logger *logrus.Logger) Proxy {
	return Proxy{
		h:      h,
		logger: logger,
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !websocket.IsWebSocketUpgrade(r) {
		p.h.ServeHTTP(w, r)
		return
	}

	p.proxy(w, r)
}

// TODO: allow modification of upgrader settings?
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func isClosedConnError(err error) bool {
	str := err.Error()
	if strings.Contains(str, "use of closed network connection") {
		return true
	}
	return websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway)
}

func (p *Proxy) proxy(w http.ResponseWriter, r *http.Request) {
	var responseHeader http.Header
	conn, err := upgrader.Upgrade(w, r, responseHeader)
	if err != nil {
		p.logger.Warnln("error upgrading websocket:", err)
		return
	}
	defer conn.Close()

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	t, err := transport.NewClientTransport(ctx, transport.TargetInfo{Addr: r.URL.Host}, transport.ConnectOptions{})
	if err != nil {
		p.logger.Warnln("error creating transport:", err)
		return
	}
	s, err := t.NewStream(ctx, &transport.CallHdr{
		Host:   r.RemoteAddr,
		Method: r.URL.Path,
	})
	if err != nil {
		p.logger.Warnln("error creating stream:", err)
		return
	}

	// read loop - reads from websocket and puts it on the stream
	go func() {
		for {
			select {
			case <-ctx.Done():
				p.logger.Debugln("[read] context canceled")
				return
			case <-s.Done():
				p.logger.Debugln("[read] done")
				return
			default:
			}
			p.logger.Debugln("[read] reading from socket.")
			_, payload, err := conn.ReadMessage()
			if err != nil {
				if isClosedConnError(err) {
					p.logger.Debugln("[read] websocket closed:", err)
					// Do I need to call this? err = t.GracefulClose()
					return
				}
				p.logger.Warnln("[read] error reading websocket message:", err)
				return
			}
			p.logger.Debugln("[read] read payload:", string(payload))
			p.logger.Debugln("[read] writing to transport")
			err = t.Write(s, payload, nil)
			p.logger.Debugln("[read] wrote to transport")
			if err != nil {
				p.logger.Warnln("[read] error writing message to transport", err)
				return
			}
		}
	}()

	// write loop -- take messages from stream and write to websocket
	var header [5]byte
	var msg []byte
	for {
		// Read header
		_, err := s.Read(header[:])
		if err != nil {
			p.logger.Warnln("[write] failed to read header:", err)
			return
		}

		// Ignore first byte, contains compression information
		len := binary.BigEndian.Uint32(header[1:])

		if len == 0 {
			// Ignore empty messages
			p.logger.Warnln("[write] empty message")
			continue
		}

		msg = make([]byte, int(len))
		if _, err := s.Read(msg); err != nil {
			p.logger.Warnln("[write] failed to read message:", err)
			return
		}

		p.logger.Debugln("[write] read", string(msg))
		if err = conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			p.logger.Warnln("[write] error writing websocket message:", err)
			return
		}
	}
}
