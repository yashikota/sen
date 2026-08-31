#!/bin/sh
set -eu
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
export SEN_HOME="${E2E_SEN_HOME:-$(mktemp -d)}"
cd "$ROOT"
bin="${E2E_SEN_BIN:-}"
if [ ! -f "$SEN_HOME/workspace.toml" ]; then
  if [ -n "$bin" ]; then
    "$bin" init
  else
    go run . init
  fi
fi
if [ -n "$bin" ]; then
  exec "$bin" serve --addr 127.0.0.1:7730
fi
exec go run . serve --addr 127.0.0.1:7730
