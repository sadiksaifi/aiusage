#!/bin/sh
set -eu

repo=https://github.com/sadiksaifi/aiusage
install_dir=${AIUSAGE_INSTALL_DIR:-"$HOME/.local/bin"}

case "$(uname -s)" in
  Darwin) os=darwin ;;
  Linux) os=linux ;;
  *) echo "aiusage supports macOS and Linux only." >&2; exit 1 ;;
esac
case "$(uname -m)" in
  arm64|aarch64) arch=arm64 ;;
  x86_64|amd64) arch=amd64 ;;
  *) echo "aiusage supports arm64 and x86_64 only." >&2; exit 1 ;;
esac

download() {
  curl --proto '=https' --tlsv1.2 -fsSL --retry 3 "$@"
}

# Resolve once so the archive and checksum always come from the same release.
release=$(download -o /dev/null -w '%{url_effective}' "$repo/releases/latest")
tag=${release##*/}
if ! printf '%s\n' "$tag" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$'; then
  echo "Cannot resolve the latest stable aiusage release." >&2
  exit 1
fi

archive=aiusage_${os}_${arch}.tar.gz
base=$repo/releases/download/$tag
tmp=$(mktemp -d)
stage=
trap 'rm -rf "$tmp"; if [ -n "$stage" ]; then rm -f "$stage"; fi' 0
trap 'exit 130' INT
trap 'exit 143' TERM

download "$base/$archive" -o "$tmp/$archive"
download "$base/checksums.txt" -o "$tmp/checksums.txt"
awk -v name="$archive" '$2 == name { print; found++ } END { if (found != 1) exit 1 }' \
  "$tmp/checksums.txt" > "$tmp/selected-checksum.txt"
if command -v sha256sum >/dev/null 2>&1; then
  (cd "$tmp" && sha256sum -c selected-checksum.txt)
else
  (cd "$tmp" && shasum -a 256 -c selected-checksum.txt)
fi
tar -xzf "$tmp/$archive" -C "$tmp" aiusage
mkdir -p "$install_dir"
stage=$(mktemp "$install_dir/.aiusage.XXXXXX")
install -m 755 "$tmp/aiusage" "$stage"
mv -f "$stage" "$install_dir/aiusage"
echo "Installed aiusage $tag to $install_dir/aiusage"
case ":$PATH:" in
  *":$install_dir:"*) ;;
  *) printf 'Add it to your PATH: export PATH="%s:%s"\n' "$install_dir" "\$PATH" ;;
esac
