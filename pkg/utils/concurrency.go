package utils

import (
	"context"
	"sync"
	"time"
)

// BatchResult represents the result of a batched operation
type BatchResult[T any] struct {
	Value T
	Error error
	Index int
}

// BatchExecutor executes operations in batches with concurrency control
type BatchExecutor[T any, R any] struct {
	concurrency int
}

// NewBatchExecutor creates a new batch executor with the given concurrency
func NewBatchExecutor[T any, R any](concurrency int) *BatchExecutor[T, R] {
	if concurrency <= 0 {
		concurrency = 10
	}
	return &BatchExecutor[T, R]{concurrency: concurrency}
}

// Execute runs the given function for each input with controlled concurrency
func (b *BatchExecutor[T, R]) Execute(ctx context.Context, inputs []T, fn func(context.Context, T) (R, error)) []BatchResult[R] {
	results := make([]BatchResult[R], len(inputs))

	if len(inputs) == 0 {
		return results
	}

	sem := make(chan struct{}, b.concurrency)
	var wg sync.WaitGroup

	for i, input := range inputs {
		select {
		case <-ctx.Done():
			results[i] = BatchResult[R]{Error: ctx.Err(), Index: i}
			continue
		case sem <- struct{}{}:
		}

		wg.Add(1)
		go func(idx int, in T) {
			defer wg.Done()
			defer func() { <-sem }()

			value, err := fn(ctx, in)
			results[idx] = BatchResult[R]{Value: value, Error: err, Index: idx}
		}(i, input)
	}

	wg.Wait()
	return results
}

// Retry executes a function with exponential backoff retry
func Retry[T any](ctx context.Context, maxAttempts int, initialDelay time.Duration, fn func(context.Context) (T, error)) (T, error) {
	var lastErr error
	var zero T

	delay := initialDelay

	for attempt := 0; attempt < maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		default:
		}

		result, err := fn(ctx)
		if err == nil {
			return result, nil
		}

		lastErr = err

		if attempt < maxAttempts-1 {
			select {
			case <-ctx.Done():
				return zero, ctx.Err()
			case <-time.After(delay):
			}
			delay *= 2
			if delay > 30*time.Second {
				delay = 30 * time.Second
			}
		}
	}

	return zero, lastErr
}

// WaitForCondition waits for a condition to be true with timeout
func WaitForCondition(ctx context.Context, interval time.Duration, fn func(context.Context) (bool, error)) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			done, err := fn(ctx)
			if err != nil {
				return err
			}
			if done {
				return nil
			}
		}
	}
}

// Pool is a simple worker pool for concurrent operations
type Pool struct {
	workers int
	jobs    chan func()
	done    chan struct{}
}

// NewPool creates a new worker pool
func NewPool(workers int) *Pool {
	if workers <= 0 {
		workers = 10
	}

	p := &Pool{
		workers: workers,
		jobs:    make(chan func(), workers*2),
		done:    make(chan struct{}),
	}

	for i := 0; i < workers; i++ {
		go p.worker()
	}

	return p
}

// worker processes jobs from the queue
func (p *Pool) worker() {
	for {
		select {
		case <-p.done:
			return
		case job, ok := <-p.jobs:
			if !ok {
				return
			}
			job()
		}
	}
}

// Submit adds a job to the pool
func (p *Pool) Submit(job func()) {
	select {
	case <-p.done:
		return
	case p.jobs <- job:
	}
}

// Stop stops the worker pool
func (p *Pool) Stop() {
	close(p.done)
	close(p.jobs)
}

// RateLimiter implements a simple rate limiter
type RateLimiter struct {
	rate     int
	interval time.Duration
	tokens   chan struct{}
	done     chan struct{}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate int, interval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		rate:     rate,
		interval: interval,
		tokens:   make(chan struct{}, rate),
		done:     make(chan struct{}),
	}

	// Fill initial tokens
	for i := 0; i < rate; i++ {
		rl.tokens <- struct{}{}
	}

	// Start token refill goroutine
	go rl.refill()

	return rl
}

// refill adds tokens at the configured rate
func (rl *RateLimiter) refill() {
	ticker := time.NewTicker(rl.interval / time.Duration(rl.rate))
	defer ticker.Stop()

	for {
		select {
		case <-rl.done:
			return
		case <-ticker.C:
			select {
			case rl.tokens <- struct{}{}:
			default:
				// Token bucket is full
			}
		}
	}
}

// Wait blocks until a token is available
func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-rl.tokens:
		return nil
	}
}

// Stop stops the rate limiter
func (rl *RateLimiter) Stop() {
	close(rl.done)
}

// Cache is a simple thread-safe cache with TTL
type Cache[K comparable, V any] struct {
	mu      sync.RWMutex
	items   map[K]cacheItem[V]
	ttl     time.Duration
	cleanup *time.Ticker
}

type cacheItem[V any] struct {
	value   V
	expires time.Time
}

// NewCache creates a new cache with the given TTL
func NewCache[K comparable, V any](ttl time.Duration) *Cache[K, V] {
	c := &Cache[K, V]{
		items:   make(map[K]cacheItem[V]),
		ttl:     ttl,
		cleanup: time.NewTicker(ttl),
	}

	go c.cleanupLoop()

	return c
}

// cleanupLoop removes expired items
func (c *Cache[K, V]) cleanupLoop() {
	for range c.cleanup.C {
		c.mu.Lock()
		now := time.Now()
		for k, v := range c.items {
			if now.After(v.expires) {
				delete(c.items, k)
			}
		}
		c.mu.Unlock()
	}
}

// Get retrieves a value from the cache
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if !ok || time.Now().After(item.expires) {
		var zero V
		return zero, false
	}

	return item.value, true
}

// Set stores a value in the cache
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = cacheItem[V]{
		value:   value,
		expires: time.Now().Add(c.ttl),
	}
}

// Delete removes a value from the cache
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Stop stops the cache cleanup
func (c *Cache[K, V]) Stop() {
	c.cleanup.Stop()
}
