#!/usr/bin/env sh
set -eu

PREFIX="${PREFIX:-$HOME/.local}"
BINDIR="${BINDIR:-$PREFIX/bin}"

mkdir -p "$BINDIR"
go -C src build -o "$BINDIR/version" .

printf 'added version command at %s/version\n' "$BINDIR"

case ":$PATH:" in
  *:"$BINDIR":*)
    printf 'version is available in PATH\n'
    ;;
  *)
    printf 'add %s to PATH to use the version command everywhere\n' "$BINDIR"
    ;;
esac
