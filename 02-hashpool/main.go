// Hashpool uses bounded parallelism and propagates cancellation through a pipeline.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"sync"
)

type Result struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256,omitempty"`
	Error  string `json:"error,omitempty"`
}

func hashFile(ctx context.Context, path string) Result {
	r := Result{Path: path}
	f, err := os.Open(path)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	defer f.Close()
	h := sha256.New()
	buf := make([]byte, 32*1024)
	for {
		if err := ctx.Err(); err != nil {
			r.Error = err.Error()
			return r
		}
		n, err := f.Read(buf)
		if n > 0 {
			_, _ = h.Write(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			r.Error = err.Error()
			return r
		}
	}
	r.SHA256 = hex.EncodeToString(h.Sum(nil))
	return r
}

// HashAll returns partial results plus ctx.Err on cancellation. File failures
// belong to individual results and do not stop unrelated jobs.
func HashAll(ctx context.Context, paths []string, workers int) ([]Result, error) {
	if workers < 1 {
		return nil, fmt.Errorf("workers must be positive")
	}
	jobs := make(chan string)
	results := make(chan Result)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case path, ok := <-jobs:
					if !ok {
						return
					}
					r := hashFile(ctx, path)
					select {
					case results <- r:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}
	go func() {
		defer close(jobs) // The sender owns closing the jobs channel.
		for _, path := range paths {
			select {
			case jobs <- path:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() { wg.Wait(); close(results) }()
	var out []Result
	for r := range results {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, ctx.Err()
}

func run() error {
	workers := flag.Int("workers", runtime.NumCPU(), "maximum concurrent file readers")
	flag.Parse()
	if flag.NArg() == 0 {
		return fmt.Errorf("usage: hashpool [-workers N] FILE...")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	results, err := HashAll(ctx, flag.Args(), *workers)
	if encErr := json.NewEncoder(os.Stdout).Encode(results); encErr != nil {
		return encErr
	}
	if err != nil {
		return err
	}
	for _, r := range results {
		if r.Error != "" {
			return fmt.Errorf("one or more files failed")
		}
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
