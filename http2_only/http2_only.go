package http2_only

import (
	// "log"
	// "net/http"
	// "os"
	// "strings"
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/textproto"
	"os"
	"strings"

	"golang.org/x/net/http/httpguts"
	// "golang.org/x/net/http2"
	"golang.org/x/net/http2"
)

var (
	http2VerboseLogs bool
)

func init() {
	e := os.Getenv("GODEBUG")
	if strings.Contains(e, "http2debug=1") || strings.Contains(e, "http2debug=2") {
		http2VerboseLogs = true
	}
}

func NewHandler(h http.Handler, s *http2.Server) http.Handler {
	return &http2onlyHandler{
		Handler: h,
		s:       s,
	}
}

type http2onlyHandler struct {
	Handler http.Handler
	s       *http2.Server
}

// ServeHTTP implements http.Handler.
func (s *http2onlyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle h2c with prior knowledge (RFC 7540 Section 3.4)
	if r.Method == "PRI" && len(r.Header) == 0 && r.URL.Path == "*" && r.Proto == "HTTP/2.0" {
		if http2VerboseLogs {
			log.Print("h2c: attempting h2c with prior knowledge.")
		}
		conn, err := initH2CWithPriorKnowledge(w)
		if err != nil {
			if http2VerboseLogs {
				log.Printf("h2c: error h2c with prior knowledge: %v", err)
			}
			return
		}
		defer conn.Close()
		s.s.ServeConn(conn, &http2.ServeConnOpts{
			Context:          r.Context(),
			BaseConfig:       extractServer(r),
			Handler:          s.Handler,
			SawClientPreface: true,
		})
		return
	}
	// Handle Upgrade to h2c (RFC 7540 Section 3.2)
	if isH2CUpgrade(r.Header) {
		conn, settings, err := h2cUpgrade(w, r)
		if err != nil {
			if http2VerboseLogs {
				log.Printf("h2c: error h2c upgrade: %v", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer conn.Close()
		s.s.ServeConn(conn, &http2.ServeConnOpts{
			Context:        r.Context(),
			BaseConfig:     extractServer(r),
			Handler:        s.Handler,
			UpgradeRequest: r,
			Settings:       settings,
		})
		return
	}
	// s.Handler.ServeHTTP(w, r)
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Bad Request\n" + "h2c only"))

	// return
}
func extractServer(r *http.Request) *http.Server {
	server, ok := r.Context().Value(http.ServerContextKey).(*http.Server)
	if ok {
		return server
	}
	return new(http.Server)
}
func isH2CUpgrade(h http.Header) bool {
	return httpguts.HeaderValuesContainsToken(h[textproto.CanonicalMIMEHeaderKey("Upgrade")], "h2c") &&
		httpguts.HeaderValuesContainsToken(h[textproto.CanonicalMIMEHeaderKey("Connection")], "HTTP2-Settings")
}
func h2cUpgrade(w http.ResponseWriter, r *http.Request) (_ net.Conn, settings []byte, err error) {
	settings, err = getH2Settings(r.Header)
	if err != nil {
		return nil, nil, err
	}
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("h2c: connection does not support Hijack")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	conn, rw, err := hijacker.Hijack()
	if err != nil {
		return nil, nil, err
	}

	rw.Write([]byte("HTTP/1.1 101 Switching Protocols\r\n" +
		"Connection: Upgrade\r\n" +
		"Upgrade: h2c\r\n\r\n"))
	return newBufConn(conn, rw), settings, nil
}
func initH2CWithPriorKnowledge(w http.ResponseWriter) (net.Conn, error) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return nil, errors.New("h2c: connection does not support Hijack")
	}
	conn, rw, err := hijacker.Hijack()
	if err != nil {
		return nil, err
	}

	const expectedBody = "SM\r\n\r\n"

	buf := make([]byte, len(expectedBody))
	n, err := io.ReadFull(rw, buf)
	if err != nil {
		return nil, errors.New("h2c: error reading client preface: " + err.Error())
	}

	if string(buf[:n]) == expectedBody {
		return newBufConn(conn, rw), nil
	}

	conn.Close()
	return nil, errors.New("h2c: invalid client preface")
}
func getH2Settings(h http.Header) ([]byte, error) {
	vals, ok := h[textproto.CanonicalMIMEHeaderKey("HTTP2-Settings")]
	if !ok {
		return nil, errors.New("missing HTTP2-Settings header")
	}
	if len(vals) != 1 {
		return nil, errors.New("expected 1 HTTP2-Settings. Got: " + fmt.Sprint(vals))
	}
	settings, err := base64.RawURLEncoding.DecodeString(vals[0])
	if err != nil {
		return nil, err
	}
	return settings, nil
}
func newBufConn(conn net.Conn, rw *bufio.ReadWriter) net.Conn {
	rw.Flush()
	if rw.Reader.Buffered() == 0 {
		// If there's no buffered data to be read,
		// we can just discard the bufio.ReadWriter.
		return conn
	}
	return &bufConn{conn, rw.Reader}
}

type bufConn struct {
	net.Conn
	*bufio.Reader
}

func (c *bufConn) Read(p []byte) (int, error) {
	if c.Reader == nil {
		return c.Conn.Read(p)
	}
	n := c.Reader.Buffered()
	if n == 0 {
		c.Reader = nil
		return c.Conn.Read(p)
	}
	if n < len(p) {
		p = p[:n]
	}
	return c.Reader.Read(p)
}
