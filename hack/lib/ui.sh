#!/usr/bin/env bash

# -----------------------------------------------------------------------------
# UI build helpers. These functions need the following variables:
#
#    AUTOK3S_SKIP_UI  -  Skip the ui build (default: false).

function autok3s::ui::build() {
  if [[ "${AUTOK3S_SKIP_UI:-}" == "true" ]]; then
    autok3s::log::info "skipping autok3s ui build (AUTOK3S_SKIP_UI=true)"
    if [[ ! -d pkg/server/ui/static ]]; then
      autok3s::log::fatal "pkg/server/ui/static is missing, remove AUTOK3S_SKIP_UI to build the ui first"
    fi
    return 0
  fi

  autok3s::log::info "building autok3s ui from source (ui/)..."
  if ! command -v npm >/dev/null 2>&1; then
    autok3s::log::fatal "npm is required to build the bundled ui"
  fi

  pushd "${ROOT_DIR}/ui" >/dev/null || return 1
  npm ci --no-audit --no-fund || npm install --no-audit --no-fund
  npm run build
  rm -rf "${ROOT_DIR}/pkg/server/ui/static"
  mv dist/static "${ROOT_DIR}/pkg/server/ui/static"
  popd >/dev/null || return 1

  autok3s::log::info "autok3s ui build done"
}
