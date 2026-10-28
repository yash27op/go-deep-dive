// Loglens processes arbitrarily long input streams without storing every line.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type Summary struct {
	Lines     int            `json:"lines"`
	Levels    map[string]int `json:"levels"`
	Malformed int            `json:"malformed"`
}

// Analyze accepts an interface, so callers can supply files, stdin, or test data.
// Format: LEVEL message; supported levels are DEBUG, INFO, WARN, and ERROR.
func Analyze(r io.Reader) (Summary, error) {
	s := Summary{Levels: make(map[string]int)}
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	for scanner.Scan() {
		s.Lines++
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			s.Malformed++
			continue
		}
		switch fields[0] {
		case "DEBUG", "INFO", "WARN", "ERROR":
			s.Levels[fields[0]]++
		default:
			s.Malformed++
		}
	}
	return s, scanner.Err()
}

func run() error {
	path := flag.String("file", "", "log file (default: stdin)")
	flag.Parse()
	var r io.Reader = os.Stdin
	if *path != "" {
		f, err := os.Open(*path)
		if err != nil {
			return err
		}
		defer f.Close()
		r = f
	}
	s, err := Analyze(r)
	if err != nil {
		return fmt.Errorf("read logs: %w", err)
	}
	return json.NewEncoder(os.Stdout).Encode(s)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
