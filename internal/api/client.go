package api

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
	maxRetries     = 3
	retryBaseDelay = 1 * time.Second
	userAgent      = "uictl/0.1.0"

	// UniFi OS consoles (UDM, UDM Pro, etc.) use the proxy path
	unifiOSAPIPath = "/proxy/network/integration"
	// Standalone Network Application uses the direct path
	standaloneAPIPath = "/integration"
	// UniFi Protect integration API on a UniFi OS console.
	protectIntegrationPrefix = "/proxy/protect/integration"

	wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
)

// protectIntegrationPaths are Protect integration routes. Network detection
// must not run for these; they 404 on the Network base.
var protectIntegrationPaths = []string{
	"/v1/alarm-hubs",
	"/v1/alarm-manager",
	"/v1/arm-profiles",
	"/v1/bridges",
	"/v1/cameras",
	"/v1/chimes",
	"/v1/files",
	"/v1/fobs",
	"/v1/lights",
	"/v1/link-stations",
	"/v1/liveviews",
	"/v1/meta",
	"/v1/nvrs",
	"/v1/pos",
	"/v1/relays",
	"/v1/sensors",
	"/v1/sirens",
	"/v1/speakers",
	"/v1/subscribe",
	"/v1/ulp-users",
	"/v1/users",
	"/v1/viewers",
}

func isProtectIntegrationPath(path string) bool {
	for _, prefix := range protectIntegrationPaths {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}

// Client is the UniFi API client.
type Client struct {
	httpClient   *http.Client
	hostURL      string // original host URL without API path prefix
	baseURL      string // host URL + detected API path prefix (/integration or /proxy/network/integration)
	apiKey       string
	verbose      bool
	debug        bool
	errWriter    io.Writer
	pathDetected bool
}

// ClientConfig configures a Client.
type ClientConfig struct {
	Host      string
	APIKey    string
	Insecure  bool
	Timeout   time.Duration
	Verbose   bool
	Debug     bool
	ErrWriter io.Writer
}

// NewClient creates a new UniFi API client.
func NewClient(cfg ClientConfig) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.Insecure,
		},
	}

	baseURL := strings.TrimRight(cfg.Host, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "https://" + baseURL
	}

	return &Client{
		httpClient: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
		hostURL:   baseURL,
		baseURL:   baseURL,
		apiKey:    cfg.APIKey,
		verbose:   cfg.Verbose,
		debug:     cfg.Debug,
		errWriter: cfg.ErrWriter,
	}
}

// detectAPIPath probes the controller to find the correct API base path.
// UniFi OS consoles use /proxy/network/integration, standalone uses /integration.
func (c *Client) detectAPIPath() {
	// Try UniFi OS path first (more common with modern hardware)
	testURL := c.baseURL + unifiOSAPIPath + "/v1/info"
	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		c.baseURL += standaloneAPIPath
		return
	}
	req.Header.Set("User-Agent", userAgent)
	if c.apiKey != "" {
		req.Header.Set("X-API-KEY", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.baseURL += standaloneAPIPath
		return
	}
	defer func() { _ = resp.Body.Close() }()

	// If we get JSON back (even a 401), the path is correct
	ct := resp.Header.Get("Content-Type")
	if strings.Contains(ct, "application/json") {
		c.baseURL += unifiOSAPIPath
		if c.verbose {
			_, _ = fmt.Fprintf(c.errWriter, "  detected UniFi OS console (proxy path)\n")
		}
		return
	}

	// Got HTML or other — try standalone path
	c.baseURL += standaloneAPIPath
	if c.verbose {
		_, _ = fmt.Fprintf(c.errWriter, "  using standalone controller path\n")
	}
}

// APIError represents an error returned by the UniFi API.
type APIError struct {
	StatusCode int    `json:"statusCode"`
	StatusName string `json:"statusName"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	RequestID  string `json:"requestId"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s (%d): %s", e.StatusName, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("%s (%d)", e.StatusName, e.StatusCode)
}

// PageResponse wraps a paginated API response.
type PageResponse struct {
	Offset     int              `json:"offset"`
	Limit      int              `json:"limit"`
	Count      int              `json:"count"`
	TotalCount int              `json:"totalCount"`
	Data       []map[string]any `json:"data"`
}

// protectBase is the Protect integration origin. Hosts that already include
// the proxy prefix are used as-is.
func (c *Client) protectBase() string {
	if strings.Contains(c.hostURL, protectIntegrationPrefix) {
		return c.hostURL
	}
	return c.hostURL + protectIntegrationPrefix
}

// Do performs an HTTP request with retry logic.
// Protect paths go to the Protect integration base and skip Network detection.
// Other paths are appended to the detected Network integration base URL.
func (c *Client) Do(method, path string, body io.Reader) (*http.Response, error) {
	if isProtectIntegrationPath(path) {
		return c.doRequest(method, c.protectBase()+path, body)
	}
	if !c.pathDetected {
		c.detectAPIPath()
		c.pathDetected = true
	}

	return c.doRequest(method, c.baseURL+path, body)
}

// DoRaw performs an HTTP request against the controller host directly,
// bypassing the /integration/ API prefix. Use this for classic API
// endpoints like /proxy/network/api/s/{site}/...
func (c *Client) DoRaw(method, path string, body io.Reader) (*http.Response, error) {
	return c.doRequest(method, c.hostURL+path, body)
}

func (c *Client) doRequest(method, url string, body io.Reader, contentType ...string) (*http.Response, error) {

	var lastErr error
	for attempt := range maxRetries {
		if body != nil && attempt > 0 {
			if seeker, ok := body.(io.Seeker); ok {
				if _, err := seeker.Seek(0, io.SeekStart); err != nil {
					return nil, fmt.Errorf("rewinding request body: %w", err)
				}
			}
		}
		req, err := http.NewRequest(method, url, body)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Accept", "application/json")
		if body != nil {
			ct := "application/json"
			if len(contentType) > 0 && contentType[0] != "" {
				ct = contentType[0]
			}
			req.Header.Set("Content-Type", ct)
		}
		if c.apiKey != "" {
			req.Header.Set("X-API-KEY", c.apiKey)
		}

		if c.verbose {
			_, _ = fmt.Fprintf(c.errWriter, "[%s] %s\n", method, url)
		}

		start := time.Now()
		resp, err := c.httpClient.Do(req)
		elapsed := time.Since(start)

		if c.verbose {
			if err != nil {
				_, _ = fmt.Fprintf(c.errWriter, "  error: %v (%.1fs)\n", err, elapsed.Seconds())
			} else {
				_, _ = fmt.Fprintf(c.errWriter, "  %d (%.1fs)\n", resp.StatusCode, elapsed.Seconds())
			}
		}

		if err != nil {
			lastErr = err
			if attempt < maxRetries-1 {
				time.Sleep(retryDelay(attempt))
			}
			continue
		}

		// Retry on 429 and 5xx
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d from %s %s", resp.StatusCode, method, url)
			if attempt < maxRetries-1 {
				delay := retryDelay(attempt)
				if c.verbose {
					_, _ = fmt.Fprintf(c.errWriter, "  retrying in %v...\n", delay)
				}
				time.Sleep(delay)
			}
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries, lastErr)
}

// Get performs a GET request and returns the response body.
func (c *Client) Get(path string) ([]byte, error) {
	resp, err := c.Do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if c.debug {
		_, _ = fmt.Fprintf(c.errWriter, "  response body: %s\n", truncate(string(data), 2000))
	}

	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp.StatusCode, data)
	}

	return data, nil
}

// GetJSON performs a GET request and unmarshals the response.
func (c *Client) GetJSON(path string, target any) error {
	data, err := c.Get(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// Post performs a POST request with a JSON body.
func (c *Client) Post(path string, body any) ([]byte, error) {
	return c.mutate("POST", path, body)
}

// Put performs a PUT request with a JSON body.
func (c *Client) Put(path string, body any) ([]byte, error) {
	return c.mutate("PUT", path, body)
}

// Patch performs a PATCH request with a JSON body.
func (c *Client) Patch(path string, body any) ([]byte, error) {
	return c.mutate("PATCH", path, body)
}

// Delete performs a DELETE request.
func (c *Client) Delete(path string) error {
	resp, err := c.Do("DELETE", path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		return parseAPIError(resp.StatusCode, data)
	}

	return nil
}

// GetRaw performs a GET request bypassing the integration API prefix.
func (c *Client) GetRaw(path string) ([]byte, error) {
	resp, err := c.DoRaw("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if c.debug {
		_, _ = fmt.Fprintf(c.errWriter, "  response body: %s\n", truncate(string(data), 2000))
	}

	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp.StatusCode, data)
	}

	return data, nil
}

// GetRawJSON performs a GET request bypassing the integration API prefix and unmarshals the response.
func (c *Client) GetRawJSON(path string, target any) error {
	data, err := c.GetRaw(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// PutRaw performs a PUT request bypassing the integration API prefix.
func (c *Client) PutRaw(path string, body any) ([]byte, error) {
	return c.mutateRaw("PUT", path, body)
}

// PostRaw performs a POST request bypassing the integration API prefix.
func (c *Client) PostRaw(path string, body any) ([]byte, error) {
	return c.mutateRaw("POST", path, body)
}

func (c *Client) mutateRaw(method, path string, body any) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		if c.debug {
			_, _ = fmt.Fprintf(c.errWriter, "  request body: %s\n", truncate(string(data), 2000))
		}
		bodyReader = strings.NewReader(string(data))
	}

	resp, err := c.DoRaw(method, path, bodyReader)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if c.debug {
		_, _ = fmt.Fprintf(c.errWriter, "  response body: %s\n", truncate(string(respData), 2000))
	}

	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp.StatusCode, respData)
	}

	return respData, nil
}

// GetAllPages fetches all pages of a paginated endpoint.
func (c *Client) GetAllPages(path string) ([]map[string]any, error) {
	var allData []map[string]any
	offset := 0
	limit := 100

	for {
		separator := "?"
		if strings.Contains(path, "?") {
			separator = "&"
		}
		pagedPath := fmt.Sprintf("%s%soffset=%d&limit=%d", path, separator, offset, limit)

		var page PageResponse
		if err := c.GetJSON(pagedPath, &page); err != nil {
			return nil, err
		}

		allData = append(allData, page.Data...)

		if offset+page.Count >= page.TotalCount {
			break
		}
		offset += page.Count
	}

	return allData, nil
}

func (c *Client) mutate(method, path string, body any) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		if c.debug {
			_, _ = fmt.Fprintf(c.errWriter, "  request body: %s\n", truncate(string(data), 2000))
		}
		bodyReader = strings.NewReader(string(data))
	}

	resp, err := c.Do(method, path, bodyReader)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if c.debug {
		_, _ = fmt.Fprintf(c.errWriter, "  response body: %s\n", truncate(string(respData), 2000))
	}

	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp.StatusCode, respData)
	}

	return respData, nil
}

func parseAPIError(statusCode int, data []byte) *APIError {
	apiErr := &APIError{StatusCode: statusCode}
	if err := json.Unmarshal(data, apiErr); err != nil {
		apiErr.Message = string(data)
	}
	if apiErr.StatusName == "" {
		apiErr.StatusName = http.StatusText(statusCode)
	}
	return apiErr
}

func retryDelay(attempt int) time.Duration {
	return time.Duration(math.Pow(2, float64(attempt))) * retryBaseDelay
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// PostFile uploads a multipart file to a Protect integration path.
// The caller passes the full path, for example fmt.Sprintf("/v1/files/%s", fileType).
// Do is not used: Do sets Content-Type application/json when a body is present.
func (c *Client) PostFile(path, filename string, r io.Reader) ([]byte, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("creating multipart file: %w", err)
	}
	if _, err := io.Copy(part, r); err != nil {
		return nil, fmt.Errorf("writing multipart file: %w", err)
	}
	if err := mw.Close(); err != nil {
		return nil, fmt.Errorf("closing multipart body: %w", err)
	}

	resp, err := c.doRequest("POST", c.protectBase()+path, bytes.NewReader(buf.Bytes()), mw.FormDataContentType())
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	if c.debug {
		_, _ = fmt.Fprintf(c.errWriter, "  response body: %s\n", truncate(string(data), 2000))
	}
	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp.StatusCode, data)
	}
	return data, nil
}

// Subscribe opens a Protect websocket and calls onMessage for each text or
// binary frame until the context is canceled or the server closes.
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
