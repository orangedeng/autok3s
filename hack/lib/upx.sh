#!/usr/bin/env bash

function autok3s::upx::validate() {
  if [[ -n "$(command -v upx)" ]]; then
    return 0
  fi
  return 1
}

function autok3s::upx::run() {
  if ! autok3s::upx::validate; then
    autok3s::log::warn "upx hasn't been installed, skip compressing binaries"
    return
  fi

  autok3s::log::info "compressing binaries"
  for file in `ls ${ROOT_DIR}/dist`; do
    if [[ $file == autok3s_darwin_* ]]; then
      # Historical issues from the upstream project (cnrancher/autok3s
      # issues #619, #516): darwin binaries don't work well with upx.
      autok3s::log::info "darwin binary doesn't work well with upx, skipping compressing this binary."
      continue
    fi
    autok3s::log::info "compressing $file"
    # Historical issue from the upstream project (cnrancher/autok3s
    # issue #456): upx -5 for better compatibility.
    upx -5 ${ROOT_DIR}/dist/$file;
    autok3s::log::info "compress $file done"
  done
}

