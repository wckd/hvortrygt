package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"time"
)

// newMux sets up the HTTP router with middleware and static file serving.
func newMux(staticFS fs.FS, cache *Cache) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/search", handleSearch)
	mux.HandleFunc("GET /api/risk", handleRisk(cache))
	mux.Handle("GET /", http.FileServerFS(staticFS))

	return withLogging(withRecovery(mux))
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

func withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// cachedGet fetches a URL with caching. Returns the response body bytes.
func cachedGet(ctx context.Context, cache *Cache, url string, ttl time.Duration) ([]byte, error) {
	if data, ok := cache.Get(url); ok {
		return data, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "hvortrygt/1.0 github.com/hvortrygt")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching %s: status %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20)) // 2 MB limit
	if err != nil {
		return nil, fmt.Errorf("reading response from %s: %w", url, err)
	}

	cache.Set(url, body, ttl)
	return body, nil
}
