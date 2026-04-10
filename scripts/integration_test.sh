#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
BIN_PATH="$TMP_DIR/version"
TEST_REPO="$TMP_DIR/repo"
ORIGIN_REPO="$TMP_DIR/origin.git"

cleanup() {
  rm -rf "$TMP_DIR"
}

assert_eq() {
  local actual="$1"
  local expected="$2"

  if [[ "$actual" != "$expected" ]]; then
    echo "expected: $expected"
    echo "actual:   $actual"
    exit 1
  fi
}

assert_contains() {
  local haystack="$1"
  local needle="$2"

  if [[ "$haystack" != *"$needle"* ]]; then
    echo "missing expected text: $needle"
    echo "$haystack"
    exit 1
  fi
}

trap cleanup EXIT

mkdir -p "$TEST_REPO"
git init --bare "$ORIGIN_REPO" >/dev/null

GOCACHE="$ROOT_DIR/.cache/go-build" go -C "$ROOT_DIR/src" build -o "$BIN_PATH" .

cd "$TEST_REPO"
git init >/dev/null
git config user.name "CI"
git config user.email "ci@example.com"
git branch -M main
git remote add origin "$ORIGIN_REPO"

printf '{\n  "name": "demo",\n  "version": "1.0.0"\n}\n' > package.json
printf 'name: demo\nversion: 1.0.0\n' > meta.yml

git add package.json meta.yml
git commit -m "initial" >/dev/null
git tag v1.2.3
git push -u origin main >/dev/null
git push origin v1.2.3 >/dev/null

current_output="$("$BIN_PATH" current)"
assert_eq "$current_output" "1.0.0"

resolve_output="$("$BIN_PATH" resolve)"
assert_contains "$resolve_output" "git repo: true"
assert_contains "$resolve_output" "local version: 1.0.0"
assert_contains "$resolve_output" "git version: 1.2.3"
assert_contains "$resolve_output" "resolved version: 1.2.3"
assert_contains "$(cat package.json)" '"version": "1.2.3"'
assert_contains "$(cat meta.yml)" 'version: 1.2.3'

patch_output="$(printf 'y\nn\n' | "$BIN_PATH" patch)"
assert_contains "$patch_output" "v1.2.4"
assert_contains "$patch_output" "create tag v1.2.4? [y/N]: "
assert_contains "$patch_output" "push tag v1.2.4? [y/N]: "
assert_eq "$(git tag --list v1.2.4)" "v1.2.4"
assert_contains "$(cat package.json)" '"version": "1.2.4"'
assert_contains "$(cat meta.yml)" 'version: 1.2.4'

force_output="$("$BIN_PATH" minor -f)"
assert_contains "$force_output" "v1.3.0"
assert_contains "$force_output" "pushed v1.3.0"
assert_eq "$(git tag --list v1.3.0)" "v1.3.0"
assert_eq "$(git --git-dir="$ORIGIN_REPO" tag --list v1.3.0)" "v1.3.0"

echo "integration test passed"
