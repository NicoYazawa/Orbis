#!/usr/bin/env sh
set -eu
script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
exec node "$script_dir/orbis.mjs" stop --profiles=app,rag,obs,ollama,search "$@"
