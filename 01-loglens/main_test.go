package main

import (
	"errors"
	"strings"
	"testing"
)

type brokenReader struct{}

func (brokenReader) Read([]byte) (int, error) { return 0, errors.New("disk failure") }

func TestAnalyze(t *testing.T) {
	s, err := Analyze(strings.NewReader("INFO started\nERROR disk full\nWARN\nunknown message\n\nINFO done"))
	if err != nil || s.Lines != 6 || s.Malformed != 3 || s.Levels["INFO"] != 2 || s.Levels["ERROR"] != 1 {
		t.Fatalf("summary=%+v err=%v", s, err)
	}
}

func TestAnalyzeReadFailure(t *testing.T) {
	if _, err := Analyze(brokenReader{}); err == nil {
		t.Fatal("expected reader error")
	}
}

func TestAnalyzeLineLimit(t *testing.T) {
	if _, err := Analyze(strings.NewReader(strings.Repeat("x", 1024*1024+1))); err == nil {
		t.Fatal("expected oversized line error")
	}
}
