package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/pragma-proto/api/internal/auth"
)

// bucket is a token-bucket rate limiter for a single key.
type bucket struct {
	mu       sync.Mutex
	tokens   float64
	capacity float64
	rate     float64 // tokens per second
	last     time.Time
}

func newBucket(capacity float64, rate float64) *bucket {
	return &bucket{
		tokens:   capacity,
		capacity: capacity,
		rate:     rate,
		last:     time.Now(),
	}
}

func (b *bucket) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.last).Seconds()
	b.last = now
	b.tokens += elapsed * b.rate
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// limiterStore manages per-key buckets.
type limiterStore struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	capacity float64
	rate     float64
}

func newStore(capacity float64, rate float64) *limiterStore {
	return &limiterStore{
		buckets:  make(map[string]*bucket),
		capacity: capacity,
		rate:     rate,
	}
}

func (s *limiterStore) allow(key string) bool {
	s.mu.Lock()
	b, ok := s.buckets[key]
	if !ok {
		b = newBucket(s.capacity, s.rate)
		s.buckets[key] = b
	}
	s.mu.Unlock()
	return b.Allow()
}

// Global rate limiter stores.
var (
	generalLimiter  = newStore(100, 100.0/60.0)  // 100/min per user
	loginLimiter    = newStore(10, 10.0/3600.0)   // 10/hr per IP
	aiLimiter       = newStore(20, 20.0/60.0)     // 20/min per user
	docLimiter      = newStore(5, 5.0/86400.0)    // 5/day per user
	uploadLimiter   = newStore(20, 20.0/3600.0)   // 20/hr per user
	passwordLimiter = newStore(5, 5.0/3600.0)     // 5/hr per email
)

// RateLimitGeneral limits authenticated API requests to 100/min per user.
func RateLimitGeneral(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if claims, ok := auth.ClaimsFromContext(r.Context()); ok {
			if !generalLimiter.allow(claims.UserID.String()) {
				http.Error(w, `{"error":"rate_limit_exceeded"}`, http.StatusTooManyRequests)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RateLimitLogin limits login attempts to 10/hr per IP address.
func RateLimitLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ip = xff
		}
		if !loginLimiter.allow(ip) {
			http.Error(w, `{"error":"too_many_login_attempts"}`, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RateLimitAI limits AI feature requests to 20/min per user.
func RateLimitAI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if claims, ok := auth.ClaimsFromContext(r.Context()); ok {
			if !aiLimiter.allow(claims.UserID.String()) {
				http.Error(w, `{"error":"ai_rate_limit_exceeded"}`, http.StatusTooManyRequests)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RateLimitDocumentGeneration limits document generation to 5/day per user.
func RateLimitDocumentGeneration(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if claims, ok := auth.ClaimsFromContext(r.Context()); ok {
			if !docLimiter.allow(claims.UserID.String()) {
				http.Error(w, `{"error":"document_generation_limit_exceeded"}`, http.StatusTooManyRequests)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RateLimitFileUpload limits file uploads to 20/hr per user.
func RateLimitFileUpload(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if claims, ok := auth.ClaimsFromContext(r.Context()); ok {
			if !uploadLimiter.allow(claims.UserID.String()) {
				http.Error(w, `{"error":"upload_rate_limit_exceeded"}`, http.StatusTooManyRequests)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RateLimitPasswordReset limits password reset requests to 5/hr per email.
// The key must be provided by the caller (the email address from the request body).
func PasswordResetAllowed(email string) bool {
	return passwordLimiter.allow(email)
}
