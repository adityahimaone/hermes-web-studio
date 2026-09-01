package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestStartupLogDoesNotLogGatewayURL(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	mainSource, err := os.ReadFile(filepath.Join(filepath.Dir(filename), "main.go"))
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	if strings.Contains(string(mainSource), `"gateway", cfg.GatewayBaseURL`) {
		t.Fatal("startup log must not include raw cfg.GatewayBaseURL")
	}
}
