#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
  printf 'usage: %s <source-revision> [output-directory]\n' "$0" >&2
  exit 64
fi

source_revision="$1"
output_dir="${2:-${TMPDIR:-/tmp}/goreecloud-search-native-artifacts}"
script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
native_dir="$(CDPATH= cd -- "${script_dir}/.." && pwd)"
repo_root="$(CDPATH= cd -- "${native_dir}/.." && pwd)"

if [[ ! "$source_revision" =~ ^[0-9a-f]{40}$ ]]; then
  printf 'source revision must be a canonical lowercase 40-character Git SHA\n' >&2
  exit 65
fi

actual_revision="$(git -C "$repo_root" rev-parse HEAD)"
if [ "$actual_revision" != "$source_revision" ]; then
  printf 'checked-out revision %s does not match requested source revision %s\n' \
    "$actual_revision" "$source_revision" >&2
  exit 66
fi

if [ -n "$(git -C "$repo_root" status --porcelain)" ]; then
  printf 'refusing to build provenance artifact from a modified working tree\n' >&2
  exit 67
fi

rm -rf "$output_dir"
mkdir -p "$output_dir"

go_version="$(go env GOVERSION)"
module_path="$(cd "$native_dir" && go list -m)"
if [ "$module_path" != 'github.com/GoreeCloud/goreecloud-search/native' ]; then
  printf 'unexpected native module identity: %s\n' "$module_path" >&2
  exit 68
fi

build_target() {
  local arch="$1"
  local artifact_name="goreecloud-search-linux-${arch}"
  local stage_dir="${output_dir}/${artifact_name}"
  local repeat_dir="${output_dir}/.repeat-${arch}"
  local binary="${stage_dir}/goreecloud-search"
  local repeat_binary="${repeat_dir}/goreecloud-search"
  local package="${output_dir}/${artifact_name}.tar.gz"

  mkdir -p "$stage_dir" "$repeat_dir"

  (
    cd "$native_dir"
    CGO_ENABLED=0 GOOS=linux GOARCH="$arch" \
      go build -trimpath -buildvcs=true -ldflags='-s -w' \
      -o "$binary" ./cmd/searchd
    CGO_ENABLED=0 GOOS=linux GOARCH="$arch" \
      go build -trimpath -buildvcs=true -ldflags='-s -w' \
      -o "$repeat_binary" ./cmd/searchd
  )

  cmp "$binary" "$repeat_binary"
  rm -rf "$repeat_dir"

  go version -m "$binary" > "${stage_dir}/BUILDINFO.txt"
  grep -Fq $'vcs.revision\t'"${source_revision}" "${stage_dir}/BUILDINFO.txt"
  grep -Fq $'vcs.modified\tfalse' "${stage_dir}/BUILDINFO.txt"

  cp "${repo_root}/LICENSE" "${stage_dir}/LICENSE"

  SOURCE_REVISION="$source_revision" \
  GO_VERSION="$go_version" \
  GO_ARCH="$arch" \
  BINARY_PATH="$binary" \
  METADATA_PATH="${stage_dir}/ARTIFACT-METADATA.json" \
  python3 - <<'PY'
import hashlib
import json
import os
from pathlib import Path

binary = Path(os.environ['BINARY_PATH'])
metadata = {
    'schema_version': 1,
    'product': 'GoreeCloud Search',
    'component': 'native-searchd',
    'source_repository': 'GoreeCloud/goreecloud-search',
    'source_revision': os.environ['SOURCE_REVISION'],
    'release_lifecycle': 'development',
    'artifact_scope': 'ci-development-candidate',
    'production_approved': False,
    'target_environment_validated': False,
    'go_version': os.environ['GO_VERSION'],
    'target': {
        'os': 'linux',
        'arch': os.environ['GO_ARCH'],
    },
    'binary': {
        'name': 'goreecloud-search',
        'sha256': hashlib.sha256(binary.read_bytes()).hexdigest(),
    },
}
Path(os.environ['METADATA_PATH']).write_text(
    json.dumps(metadata, indent=2, sort_keys=True) + '\n',
    encoding='utf-8',
)
PY

  tar --sort=name \
    --mtime='UTC 1970-01-01' \
    --owner=0 --group=0 --numeric-owner \
    -C "$output_dir" -cf - "$artifact_name" \
    | gzip -n > "$package"
}

build_target amd64
build_target arm64

(
  cd "$output_dir"
  sha256sum \
    goreecloud-search-linux-amd64.tar.gz \
    goreecloud-search-linux-arm64.tar.gz \
    > SHA256SUMS
  sha256sum --check SHA256SUMS
)

SOURCE_REVISION="$source_revision" \
GO_VERSION="$go_version" \
OUTPUT_DIR="$output_dir" \
python3 - <<'PY'
import hashlib
import json
import os
from pathlib import Path

root = Path(os.environ['OUTPUT_DIR'])
artifacts = []
for arch in ('amd64', 'arm64'):
    package = root / f'goreecloud-search-linux-{arch}.tar.gz'
    artifacts.append({
        'os': 'linux',
        'arch': arch,
        'file': package.name,
        'sha256': hashlib.sha256(package.read_bytes()).hexdigest(),
    })

manifest = {
    'schema_version': 1,
    'product': 'GoreeCloud Search',
    'component': 'native-searchd',
    'source_repository': 'GoreeCloud/goreecloud-search',
    'source_revision': os.environ['SOURCE_REVISION'],
    'release_lifecycle': 'development',
    'artifact_scope': 'ci-development-candidate',
    'production_approved': False,
    'release_candidate_declared': False,
    'target_environment_validated': False,
    'live_provider_acceptance_validated': False,
    'platform_conformance': 'nonconformant',
    'go_version': os.environ['GO_VERSION'],
    'artifacts': artifacts,
}
(root / 'artifact-provenance.json').write_text(
    json.dumps(manifest, indent=2, sort_keys=True) + '\n',
    encoding='utf-8',
)
PY

printf 'Built GoreeCloud Search native Development artifacts for %s in %s\n' \
  "$source_revision" "$output_dir"
