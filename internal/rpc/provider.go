package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// Provider manages JSON-RPC connections with failover support
type Provider struct {
	urls         []string
	currentIndex uint32
	client       *http.Client
	mu           sync.RWMutex
	timeout      time.Duration
	maxRetries   int
}

// NewProvider creates a new RPC provider with multiple URLs for failover
func NewProvider(urls []string, timeout time.Duration) *Provider {
	if len(urls) == 0 {
		panic("at least one RPC URL is required")
	}

	return &Provider{
		urls: urls,
		client: &http.Client{
			Timeout: timeout,
		},
		timeout:    timeout,
		maxRetries: len(urls),
	}
}

// Request represents a JSON-RPC request
type Request struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

// Response represents a JSON-RPC response
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error implements the error interface
func (e *RPCError) Error() string {
	return fmt.Sprintf("RPC error %d: %s", e.Code, e.Message)
}

// Call makes a JSON-RPC call with automatic failover
func (p *Provider) Call(ctx context.Context, method string, params ...interface{}) (json.RawMessage, error) {
	req := &Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	var lastErr error
	for attempt := 0; attempt < p.maxRetries; attempt++ {
		url := p.getCurrentURL()

		result, err := p.makeRequest(ctx, url, req)
		if err == nil {
			return result, nil
		}

		lastErr = err
		p.rotateURL()

		// Exponential backoff
		if attempt < p.maxRetries-1 {
			backoff := time.Duration(attempt+1) * 100 * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}
	}

	return nil, fmt.Errorf("all RPC endpoints failed: %w", lastErr)
}

// makeRequest makes a single HTTP request to the RPC endpoint
func (p *Provider) makeRequest(ctx context.Context, url string, req *Request) (json.RawMessage, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var rpcResp Response
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, rpcResp.Error
	}

	return rpcResp.Result, nil
}

// getCurrentURL returns the current RPC URL
func (p *Provider) getCurrentURL() string {
	index := atomic.LoadUint32(&p.currentIndex)
	return p.urls[index%uint32(len(p.urls))]
}

// rotateURL rotates to the next RPC URL
func (p *Provider) rotateURL() {
	atomic.AddUint32(&p.currentIndex, 1)
}

// BatchRequest allows batching multiple RPC calls
type BatchRequest struct {
	provider *Provider
	requests []*Request
	nextID   int
}

// NewBatch creates a new batch request
func (p *Provider) NewBatch() *BatchRequest {
	return &BatchRequest{
		provider: p,
		requests: make([]*Request, 0),
		nextID:   1,
	}
}

// Add adds a call to the batch
func (b *BatchRequest) Add(method string, params ...interface{}) {
	b.requests = append(b.requests, &Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      b.nextID,
	})
	b.nextID++
}

// Send sends the batch request
func (b *BatchRequest) Send(ctx context.Context) ([]json.RawMessage, error) {
	if len(b.requests) == 0 {
		return []json.RawMessage{}, nil
	}

	body, err := json.Marshal(b.requests)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch: %w", err)
	}

	url := b.provider.getCurrentURL()
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := b.provider.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	var responses []Response
	if err := json.NewDecoder(resp.Body).Decode(&responses); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	results := make([]json.RawMessage, len(responses))
	for i, r := range responses {
		if r.Error != nil {
			return nil, r.Error
		}
		results[i] = r.Result
	}

	return results, nil
}

// Close closes the provider
func (p *Provider) Close() error {
	p.client.CloseIdleConnections()
	return nil
}
