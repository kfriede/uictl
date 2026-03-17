package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
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
)

// Client is the UniFi API client.
type Client struct {
	httpClient   *http.Client
	baseURL      string
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

// Do performs an HTTP request with retry logic.
func (c *Client) Do(method, path string, body io.Reader) (*http.Response, error) {
	if !c.pathDetected {
		c.detectAPIPath()
		c.pathDetected = true
	}

	url := c.baseURL + path

	var lastErr error
	for attempt := range maxRetries {
		req, err := http.NewRequest(method, url, body)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Accept", "application/json")
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
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
			lastErr = fmt.Errorf("HTTP %d from %s %s", resp.StatusCode, method, path)
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
