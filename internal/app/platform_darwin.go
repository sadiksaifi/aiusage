package app

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
)

// security is part of macOS. No user-installed credential helper is needed.
func claudeKeychain(ctx context.Context, configDir string) ([]byte, error) {
	service := "Claude Code-credentials"
	if os.Getenv("CLAUDE_CONFIG_DIR") != "" {
		hash := sha256.Sum256([]byte(configDir))
		service += fmt.Sprintf("-%x", hash)[:9]
	}
	data, err := exec.CommandContext(ctx, "/usr/bin/security", "find-generic-password", "-s", service, "-w").Output()
	if err != nil {
		return nil, fmt.Errorf("macOS Keychain entry for Claude unavailable or locked")
	}
	return data, nil
}
