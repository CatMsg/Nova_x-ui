//go:build !windows

package xray

import (
	"os"
	"path/filepath"
	"testing"

	xuilogger "github.com/CatMsg/Nova_x-ui/v3/logger"
	"github.com/op/go-logging"
)

func TestEnsureBinaryAvailableSkipsProvisionWhenBinaryExists(t *testing.T) {
	initProvisionTestLogger(t)

	binDir := filepath.Join(t.TempDir(), "bin")
	t.Setenv("XUI_BIN_FOLDER", binDir)

	binPath := GetBinaryPath()
	if err := os.MkdirAll(filepath.Dir(binPath), 0o755); err != nil {
		t.Fatalf("mkdir binary dir: %v", err)
	}
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}

	calls := 0
	oldProvision := provisionXrayBinary
	provisionXrayBinary = func() error {
		calls++
		return nil
	}
	t.Cleanup(func() {
		provisionXrayBinary = oldProvision
	})

	if err := EnsureBinaryAvailable(); err != nil {
		t.Fatalf("EnsureBinaryAvailable: %v", err)
	}
	if calls != 0 {
		t.Fatalf("provision was called %d times, want 0", calls)
	}
}

func TestEnsureBinaryAvailableProvisionsMissingBinary(t *testing.T) {
	initProvisionTestLogger(t)

	binDir := filepath.Join(t.TempDir(), "bin")
	t.Setenv("XUI_BIN_FOLDER", binDir)

	binPath := GetBinaryPath()
	calls := 0
	oldProvision := provisionXrayBinary
	provisionXrayBinary = func() error {
		calls++
		if err := os.MkdirAll(filepath.Dir(binPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(binPath, []byte("#!/bin/sh\nexit 0\n"), 0o755)
	}
	t.Cleanup(func() {
		provisionXrayBinary = oldProvision
	})

	if err := EnsureBinaryAvailable(); err != nil {
		t.Fatalf("EnsureBinaryAvailable: %v", err)
	}
	if calls != 1 {
		t.Fatalf("provision was called %d times, want 1", calls)
	}
	if _, err := os.Stat(binPath); err != nil {
		t.Fatalf("binary was not created: %v", err)
	}
}

func initProvisionTestLogger(t *testing.T) {
	t.Helper()
	xuilogger.InitLogger(logging.ERROR)
}
