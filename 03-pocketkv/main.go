// PocketKV is a single-process JSON store with synchronized, atomic snapshots.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"time"
)

type Store struct {
	mu   sync.RWMutex
	data map[string]string
	path string
}

func OpenStore(path string) (*Store, error) {
	s := &Store{data: make(map[string]string), path: path}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &s.data); err != nil {
		return nil, fmt.Errorf("decode snapshot: %w", err)
	}
	if s.data == nil {
		s.data = make(map[string]string)
	}
	return s, nil
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

// Set persists a copy before publishing it to readers. A failed write leaves
// memory unchanged. The lock serializes writers throughout the transaction.
func (s *Store) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := make(map[string]string, len(s.data)+1)
	for k, v := range s.data {
		next[k] = v
	}
	next[key] = value
	b, err := json.Marshal(next)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(s.path), ".pocketkv-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(f.Name(), s.path); err != nil {
		return err
	}
	s.data = next
	return nil
}

func Handler(s *Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/kv/{key}", func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		switch r.Method {
		case http.MethodGet:
			value, ok := s.Get(key)
			if !ok {
				http.Error(w, "key not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"value": value})
		case http.MethodPut:
			r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
			defer r.Body.Close()
			var body struct {
				Value *string `json:"value"`
			}
			decoder := json.NewDecoder(r.Body)
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&body); err != nil || body.Value == nil {
				http.Error(w, "expected JSON with a string value field (max 64 KiB)", http.StatusBadRequest)
				return
			}
			if err := decoder.Decode(new(any)); err != io.EOF {
				http.Error(w, "expected exactly one JSON object", http.StatusBadRequest)
				return
			}
			if err := s.Set(key, *body.Value); err != nil {
				http.Error(w, "could not persist value", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.Header().Set("Allow", "GET, PUT")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	return mux
}

func run() error {
	addr := flag.String("addr", "127.0.0.1:8080", "listen address")
	path := flag.String("data", "store.json", "snapshot file (parent directory must exist)")
	flag.Parse()
	s, err := OpenStore(*path)
	if err != nil {
		return err
	}
	server := &http.Server{Addr: *addr, Handler: Handler(s), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	done := make(chan error, 1)
	go func() { done <- server.ListenAndServe() }()
	fmt.Fprintln(os.Stderr, "PocketKV listening on", *addr)
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			_ = server.Close()
			return err
		}
		return nil
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
