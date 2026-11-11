package api

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

const wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

// Subscribe opens a Protect websocket and calls onMessage for each text or
// binary frame until the context is canceled or the server closes.
// path may include a query string; it is preserved on the request target.
func (c *Client) Subscribe(ctx context.Context, path string, onMessage func([]byte) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if onMessage == nil {
		return fmt.Errorf("onMessage is required")
	}

	base, err := url.Parse(c.protectBase())
	if err != nil {
		return fmt.Errorf("parsing protect base: %w", err)
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	target := strings.TrimRight(base.EscapedPath(), "/") + path

	addr := base.Host
	if _, _, splitErr := net.SplitHostPort(base.Host); splitErr != nil {
		port := "443"
		if base.Scheme == "http" {
			port = "80"
		}
		addr = net.JoinHostPort(base.Hostname(), port)
	}

	dialer := &net.Dialer{}
	raw, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("dialing websocket: %w", err)
	}

	conn := raw
	if base.Scheme != "http" {
		tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}
		if tr, ok := c.httpClient.Transport.(*http.Transport); ok && tr.TLSClientConfig != nil {
			tlsCfg = tr.TLSClientConfig.Clone()
		}
		if tlsCfg.ServerName == "" {
			tlsCfg.ServerName = base.Hostname()
		}
		tlsConn := tls.Client(raw, tlsCfg)
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			_ = raw.Close()
			return fmt.Errorf("tls handshake: %w", err)
		}
		conn = tlsConn
	}

	done := make(chan struct{})
	defer close(done)
	defer func() { _ = conn.Close() }()
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-done:
		}
	}()

	keyRaw := make([]byte, 16)
	if _, err := rand.Read(keyRaw); err != nil {
		return fmt.Errorf("websocket key: %w", err)
	}
	wsKey := base64.StdEncoding.EncodeToString(keyRaw)

	var handshake strings.Builder
	fmt.Fprintf(&handshake, "GET %s HTTP/1.1\r\n", target)
	fmt.Fprintf(&handshake, "Host: %s\r\n", base.Host)
	fmt.Fprintf(&handshake, "Upgrade: websocket\r\n")
	fmt.Fprintf(&handshake, "Connection: Upgrade\r\n")
	fmt.Fprintf(&handshake, "Sec-WebSocket-Version: 13\r\n")
	fmt.Fprintf(&handshake, "Sec-WebSocket-Key: %s\r\n", wsKey)
	if c.apiKey != "" {
		fmt.Fprintf(&handshake, "X-API-KEY: %s\r\n", c.apiKey)
	}
	fmt.Fprintf(&handshake, "Accept: */*\r\n")
	fmt.Fprintf(&handshake, "\r\n")
	if _, err := io.WriteString(conn, handshake.String()); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("writing websocket handshake: %w", err)
	}

	br := bufio.NewReader(conn)
	status, err := br.ReadString('\n')
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("reading websocket status: %w", err)
	}
	if !strings.Contains(status, " 101 ") {
		return fmt.Errorf("websocket upgrade failed: %s", strings.TrimSpace(status))
	}
	var accept string
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("reading websocket headers: %w", err)
		}
		if line == "\r\n" || line == "\n" {
			break
		}
		trimmed := strings.TrimRight(line, "\r\n")
		name, value, ok := strings.Cut(trimmed, ":")
		if ok && strings.EqualFold(strings.TrimSpace(name), "Sec-WebSocket-Accept") {
			accept = strings.TrimSpace(value)
		}
	}
	if accept != "" && accept != websocketAccept(wsKey) {
		return fmt.Errorf("websocket accept mismatch")
	}

	var frag []byte
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		opcode, payload, fin, err := readWebsocketFrame(br)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}
		switch opcode {
		case 0x1, 0x2:
			if !fin {
				frag = append(frag[:0], payload...)
				continue
			}
			if err := onMessage(payload); err != nil {
				return err
			}
		case 0x0:
			frag = append(frag, payload...)
			if fin {
				msg := frag
				frag = nil
				if err := onMessage(msg); err != nil {
					return err
				}
			}
		case 0x8:
			return nil
		case 0x9:
			if err := writeMaskedWebsocketFrame(conn, 0xA, payload); err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				return err
			}
		}
	}
}

func websocketAccept(key string) string {
	sum := sha1.Sum([]byte(key + wsGUID))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func readWebsocketFrame(r *bufio.Reader) (opcode byte, payload []byte, fin bool, err error) {
	var hdr [2]byte
	if _, err = io.ReadFull(r, hdr[:]); err != nil {
		return 0, nil, false, err
	}
	fin = hdr[0]&0x80 != 0
	opcode = hdr[0] & 0x0F
	masked := hdr[1]&0x80 != 0
	length := uint64(hdr[1] & 0x7F)
	switch length {
	case 126:
		var ext [2]byte
		if _, err = io.ReadFull(r, ext[:]); err != nil {
			return 0, nil, false, err
		}
		length = uint64(binary.BigEndian.Uint16(ext[:]))
	case 127:
		var ext [8]byte
		if _, err = io.ReadFull(r, ext[:]); err != nil {
			return 0, nil, false, err
		}
		length = binary.BigEndian.Uint64(ext[:])
	}
	if length > 8<<20 {
		return 0, nil, false, fmt.Errorf("websocket frame too large")
	}
	var maskKey [4]byte
	if masked {
		if _, err = io.ReadFull(r, maskKey[:]); err != nil {
			return 0, nil, false, err
		}
	}
	payload = make([]byte, length)
	if length > 0 {
		if _, err = io.ReadFull(r, payload); err != nil {
			return 0, nil, false, err
		}
	}
	if masked {
		for i := range payload {
			payload[i] ^= maskKey[i%4]
		}
	}
	return opcode, payload, fin, nil
}

func writeMaskedWebsocketFrame(w io.Writer, opcode byte, payload []byte) error {
	var maskKey [4]byte
	if _, err := rand.Read(maskKey[:]); err != nil {
		return err
	}
	header := make([]byte, 0, 14+len(payload))
	header = append(header, 0x80|opcode)
	n := len(payload)
	switch {
	case n < 126:
		header = append(header, byte(n)|0x80)
	case n <= 65535:
		header = append(header, 126|0x80, byte(n>>8), byte(n))
	default:
		var ext [8]byte
		binary.BigEndian.PutUint64(ext[:], uint64(n))
		header = append(header, 127|0x80)
		header = append(header, ext[:]...)
	}
	header = append(header, maskKey[:]...)
	masked := make([]byte, n)
	for i := range payload {
		masked[i] = payload[i] ^ maskKey[i%4]
	}
	if _, err := w.Write(header); err != nil {
		return err
	}
	_, err := w.Write(masked)
	return err
}
