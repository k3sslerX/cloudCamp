package tokens

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrClientNotFound    = errors.New("client not found")
)

type ClientConfig struct {
	Capacity  int
	Rate      int
	Frequency time.Duration
}

type TokenBucket struct {
	config      ClientConfig
	tokens      int
	lastUpdated time.Time
	mu          sync.Mutex
}

type RateLimiter struct {
	clients map[string]*TokenBucket
	configs map[string]ClientConfig
	mu      sync.RWMutex
	stop    chan struct{}
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*TokenBucket),
		configs: make(map[string]ClientConfig),
		stop:    make(chan struct{}),
	}
}

func (rl *RateLimiter) Start() {
	go rl.refillBuckets()
}

func (rl *RateLimiter) Stop() {
	close(rl.stop)
}

func (rl *RateLimiter) SetClientConfig(clientID string, config ClientConfig) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.configs[clientID] = config

	if bucket, exists := rl.clients[clientID]; exists {
		bucket.mu.Lock()
		bucket.config = config
		bucket.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(clientID string) (bool, error) {
	rl.mu.RLock()
	bucket, exists := rl.clients[clientID]
	config, configExists := rl.configs[clientID]
	rl.mu.RUnlock()

	if !configExists {
		return false, ErrClientNotFound
	}

	if !exists {
		rl.mu.Lock()
		if _, exists = rl.clients[clientID]; !exists {
			bucket = &TokenBucket{
				config:      config,
				tokens:      config.Capacity,
				lastUpdated: time.Now(),
			}
			rl.clients[clientID] = bucket
		}
		rl.mu.Unlock()
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	rl.refill(bucket)

	if bucket.tokens > 0 {
		bucket.tokens--
		return true, nil
	}

	return false, ErrRateLimitExceeded
}

func (rl *RateLimiter) refill(bucket *TokenBucket) {
	now := time.Now()
	elapsed := now.Sub(bucket.lastUpdated)

	tokensToAdd := int(elapsed/bucket.config.Frequency) * bucket.config.Rate
	if tokensToAdd > 0 {
		bucket.tokens = min(bucket.tokens+tokensToAdd, bucket.config.Capacity)
		bucket.lastUpdated = now
	}
}

func (rl *RateLimiter) refillBuckets() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanupInactiveBuckets()
		case <-rl.stop:
			return
		}
	}
}

func (rl *RateLimiter) cleanupInactiveBuckets() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for clientID, bucket := range rl.clients {
		bucket.mu.Lock()
		inactiveFor := now.Sub(bucket.lastUpdated)
		bucket.mu.Unlock()

		if inactiveFor > 5*time.Minute {
			delete(rl.clients, clientID)
		}
	}
}
