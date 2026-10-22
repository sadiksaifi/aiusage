package app

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHomebrewUpgradeUsesOwningInstallation(t *testing.T) {
	for _, failure := range []string{"", "update", "upgrade", "missing"} {
		t.Run("failure="+failure, func(t *testing.T) {
			prefix := t.TempDir()
			path := filepath.Join(prefix, "Cellar", "aiusage", "0.1.0", "bin", "aiusage")
			require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
			require.NoError(t, os.WriteFile(path, []byte("managed binary"), 0o755))
			require.NoError(t, os.MkdirAll(filepath.Join(prefix, "bin"), 0o755))
			log := filepath.Join(prefix, "commands")
			script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> '" + log + "'\nprintf 'brew output\\n'\n"
			if failure != "" {
				script += "if [ \"$1\" = '" + failure + "' ]; then exit 1; fi\n"
			}
			if failure != "missing" {
				require.NoError(t, os.WriteFile(filepath.Join(prefix, "bin", "brew"), []byte(script), 0o755))
			}
			var output bytes.Buffer
			err := upgradeHomebrew(t.Context(), &output, path)
			if failure != "" {
				require.ErrorContains(t, err, "Homebrew")
			} else {
				require.NoError(t, err)
			}
			commands, err := os.ReadFile(log)
			switch failure {
			case "missing":
				require.ErrorIs(t, err, os.ErrNotExist)
			case "update":
				require.NoError(t, err)
				require.Equal(t, "update\n", string(commands))
			default:
				require.NoError(t, err)
				require.Equal(t, "update\nupgrade sadiksaifi/tap/aiusage\n", string(commands))
			}
			if failure != "missing" {
				require.Contains(t, output.String(), "brew output")
			}
			binary, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, "managed binary", string(binary))
		})
	}
}

func TestHomebrewDetection(t *testing.T) {
	for _, path := range []string{
		"/opt/homebrew/Cellar/aiusage/0.1.0/bin/aiusage",
		"/usr/local/Cellar/aiusage/0.2.0/bin/aiusage",
		"/home/linuxbrew/.linuxbrew/Cellar/aiusage/0.1.0/bin/aiusage",
		"/custom prefix/Cellar/aiusage/0.1.0/bin/aiusage",
	} {
		require.True(t, isHomebrewPath(path), path)
	}
	for _, path := range []string{
		"/home/user/.local/bin/aiusage",
		"/home/user/go/bin/aiusage",
		"/tmp/Cellar/other/0.1.0/bin/aiusage",
		"/tmp/Cellar/aiusage/bin/aiusage",
		"/tmp/Cellar/aiusage/0.1.0/share/aiusage",
	} {
		require.False(t, isHomebrewPath(path), path)
	}
}
