#!/usr/bin/env bash

# -----------------------------------------------------------------------------
# Build the bundled ui (source in ui/) and place the static assets in
# pkg/server/ui/static, ready to be embedded by `go build -tags prod` or
# served directly by non-prod builds.
#
# Usage:
#   make ui            # build the ui (npm ci + vite build)
#   AUTOK3S_SKIP_UI=true make ui   # no-op if assets already exist

source "$(cd "$(dirname "${BASH_SOURCE[0]}")/../lib" && pwd -P)/init.sh"

autok3s::ui::build
