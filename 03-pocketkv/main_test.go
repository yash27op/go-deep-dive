package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestHTTPAndRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "store.json")
	s, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	h := Handler(s)
	for _, tc := range []struct {
		method, path, body string
		status             int
	}{
		{"GET", "/kv/name", "", 404},
		{"PUT", "/kv/name", `{"value":"gopher"}`, 204},
		{"GET", "/kv/name", "", 200},
		{"PUT", "/kv/name", `{"value":12}`, 400},
		{"PUT", "/kv/name", `{"value":"a"} {}`, 400},
		{"PUT", "/kv/name", `{}`, 400},
		{"PUT", "/kv/name", `{"value":"` + strings.Repeat("x", 65536) + `"}`, 400},
		{"DELETE", "/kv/name", "", 405},
	} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)))
		if w.Code != tc.status {
			t.Fatalf("%s %s: got %d body=%s", tc.method, tc.path, w.Code, w.Body)
		}
	}
	reopened, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := reopened.Get("name"); !ok || v != "gopher" {
		t.Fatalf("restart lost value: %q", v)
	}
}

func TestFailedWriteDoesNotPublish(t *testing.T) {
	s, err := OpenStore(filepath.Join(t.TempDir(), "missing", "store.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Set("key", "value"); err == nil {
		t.Fatal("expected persistence error")
	}
	if _, ok := s.Get("key"); ok {
		t.Fatal("failed write changed memory")
	}
}

func TestConcurrentRequests(t *testing.T) {
	s, err := OpenStore(filepath.Join(t.TempDir(), "store.json"))
	if err != nil {
		t.Fatal(err)
	}
	h := Handler(s)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			path := fmt.Sprintf("/kv/key%d", i)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest(http.MethodPut, path, strings.NewReader(`{"value":"ok"}`)))
			if w.Code != 204 {
				t.Errorf("write: %d", w.Code)
			}
			w = httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
			if w.Code != 200 {
				t.Errorf("read: %d", w.Code)
			}
		}(i)
	}
	wg.Wait()
	for i := 0; i < 20; i++ {
		if _, ok := s.Get(fmt.Sprintf("key%d", i)); !ok {
			t.Fatalf("lost key %d", i)
		}
	}
}
