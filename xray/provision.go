package xray

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/CatMsg/Nova_x-ui/v3/config"
	"github.com/CatMsg/Nova_x-ui/v3/logger"
)

var provisionMu sync.Mutex

// provisionXrayBinary installs the local Xray binary when it is missing.
// The default implementation reuses the repo-local provisioner script so
// local development and source checkouts keep the same asset-selection logic.
var provisionXrayBinary = defaultProvisionXrayBinary

// EnsureBinaryAvailable makes sure the Xray binary referenced by GetBinaryPath
// exists and is executable. If it is missing, the function provisions it once
// before returning. Callers can use this as a lightweight preflight check
// before attempting to start or restart Xray.
func EnsureBinaryAvailable() error {
	if binaryAvailable() {
		return nil
	}

	provisionMu.Lock()
	defer provisionMu.Unlock()

	if binaryAvailable() {
		return nil
	}

	logger.Infof("xray binary missing at %s, provisioning automatically", GetBinaryPath())
	if err := provisionXrayBinary(); err != nil {
		return err
	}

	if !binaryAvailable() {
		return fmt.Errorf("xray binary is still missing after provisioning: %s", GetBinaryPath())
	}

	logger.Infof("xray binary provisioned at %s", GetBinaryPath())
	return nil
}

func binaryAvailable() bool {
	info, err := os.Stat(GetBinaryPath())
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode().Perm()&0o111 != 0
}

func defaultProvisionXrayBinary() error {
	scriptPath := filepath.Join(filepath.Dir(config.GetBinFolderPath()), "scripts", "provision_xray.py")
	if _, err := os.Stat(scriptPath); err != nil {
		return fmt.Errorf("provision script not found at %s: %w", scriptPath, err)
	}

	interpreters := []string{"python3", "python"}
	var lastErr error
	for _, interpreter := range interpreters {
		cmd := exec.Command(interpreter, scriptPath)
		cmd.Env = os.Environ()
		output, err := cmd.CombinedOutput()
		if err == nil {
			return nil
		}
		if errors.Is(err, exec.ErrNotFound) {
			lastErr = err
			continue
		}
		return fmt.Errorf("%s %s failed: %w: %s", interpreter, scriptPath, err, strings.TrimSpace(string(output)))
	}

	if lastErr != nil {
		return fmt.Errorf("no python interpreter available to run %s: %w", scriptPath, lastErr)
	}

	return fmt.Errorf("failed to provision xray binary with %s", scriptPath)
}
