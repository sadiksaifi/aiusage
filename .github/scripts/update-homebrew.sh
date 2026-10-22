#!/bin/sh
set -eu

tag=$1
archive=$2
tap=$3
version=${tag#v}
if ! printf '%s\n' "$tag" | grep -Eq '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'; then
  echo "Expected a stable aiusage release tag, got: $tag" >&2
  exit 1
fi

# v0.1.0 and earlier lack installation-aware upgrades and can overwrite the Cellar.
if [ "$(printf '%s\n' 0.1.0 "$version" | sort -V | tail -1)" = 0.1.0 ]; then
  echo "Skipping $tag: releases through v0.1.0 cannot safely upgrade Homebrew installations."
  exit 0
fi

formula=$tap/Formula/aiusage.rb
if [ -f "$formula" ]; then
  current=$(awk -F '"' '/^  url / { print $2 }' "$formula")
  current=${current##*/v}
  current=${current%.tar.gz}
  if ! printf '%s\n' "$current" | grep -Eq '^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'; then
    echo "Cannot determine the current Homebrew formula version." >&2
    exit 1
  fi
  if [ "$(printf '%s\n' "$current" "$version" | sort -V | tail -1)" != "$version" ]; then
    echo "Skipping $tag: the tap already has aiusage $current."
    exit 0
  fi
fi

checksum=$(shasum -a 256 "$archive")
checksum=${checksum%% *}
mkdir -p "$tap/Formula"
cat > "$formula" <<EOF
class Aiusage < Formula
  desc "Claude Code and Codex usage, with local reset times and countdowns"
  homepage "https://github.com/sadiksaifi/aiusage"
  url "https://github.com/sadiksaifi/aiusage/archive/refs/tags/$tag.tar.gz"
  sha256 "$checksum"

  depends_on "go" => :build

  def install
    ENV["CGO_ENABLED"] = "0"
    ldflags = "-s -w -X github.com/sadiksaifi/aiusage/internal/app.version=v#{version}"
    system "go", "build", *std_go_args(ldflags:), "./cmd/aiusage"
  end

  test do
    assert_match "v#{version}", shell_output("#{bin}/aiusage --version")
    assert_match "Claude Code and Codex", shell_output("#{bin}/aiusage --help")
  end
end
EOF
