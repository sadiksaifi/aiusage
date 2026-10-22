package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/google/go-github/v86/github"
	"github.com/stretchr/testify/require"
)

func TestUpgradeDoesNotReplaceEqualOrNewerVersions(t *testing.T) {
	for _, current := range []string{"v0.1.0", "v0.2.0"} {
		t.Run(current, func(t *testing.T) {
			source := &upgradeSource{}
			updater, err := selfupdate.NewUpdater(selfupdate.Config{
				Source: source, Validator: &selfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"},
			})
			require.NoError(t, err)
			var output bytes.Buffer
			err = upgradeBinary(t.Context(), &output, current, filepath.Join(t.TempDir(), "aiusage"), updater)
			require.NoError(t, err)
			require.False(t, source.downloaded, "an equal or newer version must never be replaced")
			require.Contains(t, output.String(), "already up to date")
		})
	}
}

// Only the release-service boundary is replaced; selection uses the real updater.
type upgradeSource struct{ downloaded bool }

func (*upgradeSource) ListReleases(context.Context, selfupdate.Repository) ([]selfupdate.SourceRelease, error) {
	return []selfupdate.SourceRelease{selfupdate.NewGitHubRelease(&github.RepositoryRelease{
		TagName: new("v0.1.0"),
		Assets: []*github.ReleaseAsset{
			{ID: new(int64(1)), Name: new(fmt.Sprintf("aiusage_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH))},
			{ID: new(int64(2)), Name: new("checksums.txt")},
		},
	})}, nil
}

func (s *upgradeSource) DownloadReleaseAsset(context.Context, *selfupdate.Release, int64) (io.ReadCloser, error) {
	s.downloaded = true
	return nil, fmt.Errorf("unexpected download")
}
