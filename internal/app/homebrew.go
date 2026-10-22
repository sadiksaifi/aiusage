package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// The caller resolves symlinks before classifying the executable.
func isHomebrewPath(path string) bool {
	bin := filepath.Dir(path)
	version := filepath.Dir(bin)
	formula := filepath.Dir(version)
	return filepath.Base(path) == "aiusage" && filepath.Base(bin) == "bin" &&
		filepath.Base(formula) == "aiusage" && filepath.Base(filepath.Dir(formula)) == "Cellar"
}

func upgradeHomebrew(ctx context.Context, out io.Writer, path string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	// Use the brew that owns this Cellar, not a different installation on PATH.
	prefix := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(path)))))
	brew := filepath.Join(prefix, "bin", "brew")
	if _, err := fmt.Fprintln(out, "Upgrading aiusage with Homebrew..."); err != nil {
		return err
	}
	for _, args := range [][]string{{"update"}, {"upgrade", "sadiksaifi/tap/aiusage"}} {
		cmd := exec.CommandContext(ctx, brew, args...)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, out, out
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("upgrade with Homebrew failed (%s); run %s upgrade sadiksaifi/tap/aiusage manually: %w", args[0], brew, err)
		}
	}
	return nil
}
