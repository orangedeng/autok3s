# AutoK3s

A CLI and web UI for provisioning and managing K3s clusters (lightweight Kubernetes)
across cloud providers. One Go binary ships two runtimes — the **CLI** (`autok3s
create/join/delete/ssh/…`, cobra-based) and the **UI server** (`autok3s serve`: a
rancher/apiserver HTTP API plus the embedded Vue UI from `ui/`).

## Project

Stack: Go, cobra for the CLI, rancher/apiserver for the serve runtime, sqlite
(`glebarez/sqlite`) for persisted state, k3d for local clusters, ginkgo for
integration tests. The web UI is a Vue 3 / Vite app in `ui/` (it has its own
AGENTS.md) built into `pkg/server/ui/static` and embedded via `//go:embed` in prod
builds. SDK forks live in `thirdparty/` and are pinned by `replace` directives in
`go.mod`.

The module path is `github.com/orangedeng/autok3s` — imports must use the module
path, regardless of where the repository is cloned.

Directories whose contents are not obvious from the name:

```
cmd/             cobra commands; cmd.go blank-imports every provider package to
                 register its factory — a provider not imported here does not exist
pkg/airgap/      air-gapped install support (download/distribute K3s artifacts)
pkg/cluster/     cluster discovery for the UI/CLI list (ListClusters) plus kube/,
                 the API server access layer (client, /readyz & /version probes,
                 node description, kubeconfig persistence)
pkg/common/      sqlite DB + models, logging, settings, bundled templates
pkg/k3s/         pure assembly of the K3s install/join/restart/uninstall shell
                 commands (no IO; table driven tests pin the behavior)
pkg/providers/   cloud providers; provider.go defines the Provider interface and the
                 factory registry (RegisterProvider / GetProvider); base/ holds the
                 lifecycle orchestration (create/join/delete/upgrade, flags, config
                 merging) that every provider embeds
pkg/server/      serve runtime: schema.go registers API schemas, store/ holds the
                 per-resource stores and websocket handlers, ui/ serves the web UI
pkg/types/       shared flag/cluster/SSH types; types/apis are the API schema structs
test/            ginkgo suites: integration/ (verify stage), e2e/ (currently disabled)
hack/            shell scripts behind every make target (make-rules/ + lib/)
thirdparty/      vendored SDK forks pinned by go.mod replace — never hand-edit
ui/              the web UI; has its own AGENTS.md
```

## Build and test

Everything goes through `make autok3s <stage>`. Stages run in order
`mod -> lint -> build -> package -> unit -> verify -> e2e`; appending `only` runs
just that stage:

- `make autok3s lint only` — golangci-lint over `cmd/`, `pkg/`, `test/`
  (`.golangci.yml` enables only `revive` + `gofmt` and skips `_test.go` files);
  rolls back to `go fmt` + `go vet` when golangci-lint is not installed.
- `make autok3s unit only` — `go test -tags=test -race -cover ./cmd/... ./pkg/...`
  with CGO_ENABLED=1.
- `make autok3s verify only` — ginkgo over `test/integration`.
- `make ui` — builds `ui/` into `pkg/server/ui/static`; `AUTOK3S_SKIP_UI=true` makes
  it a no-op when the assets already exist.
- `make autok3s build` runs `go generate`, builds the UI in the background, then
  compiles with `-tags netgo,prod` for each platform. The `prod` tag embeds the UI
  (`pkg/server/ui/ui_prod.go`); without it the server serves `pkg/server/ui/static`
  from disk or proxies to a vite dev server on :5173 (`AUTOK3S_UI_MODE` /
  `AUTOK3S_UI_INDEX` env vars). Run `make ui` after UI changes, or the embedded
  assets go stale.

## Code style

- gofmt formatting; revive is the only lint gate — match the surrounding file rather
  than restyling it.
- Build tags are load-bearing (`prod`, `test`, plus `netgo` on release builds). A
  plain `go build` compiles the non-prod variants only; don't draw conclusions about
  `pkg/server/ui` from it.
- **Comments explain what the code does and why — never the change history.** Never
  write comments like "previously X, now Y", "renamed from…", or references to a
  commit, PR, or past bug. The same applies to test comments: state the invariant
  being verified, not the bug that motivated it.
- Generated code (`go generate` from `main.go`) and files under `thirdparty/` are
  not hand-written territory; regenerate or leave them alone.

## Adding a provider or an API resource

- **New provider**: implement `providers.Provider` in `pkg/providers/<name>/`, call
  `providers.RegisterProvider` from the package's `init()`, blank-import the package
  in `cmd/cmd.go`, and put its option structs under `pkg/types/<name>/`. Flags are
  returned as `types.Flag` slices from the `Get*Flags` methods — both the CLI and
  the UI render from them, so a flag added there shows up in both.
- **New API resource**: struct in `pkg/types/apis`, registered in
  `pkg/server/schema.go` via `MustImportAndCustomize` (collection/resource methods,
  actions, formatters), with its store under `pkg/server/store/<resource>/`. The UI
  is driven by what is registered there — changing methods or actions changes the UI.

## Security-sensitive changes

A mistake in these leaks data rather than breaking a build, so review them
deliberately and say in the PR what you checked:

- **Credentials** (cloud provider secrets, SSH keys) are persisted in the sqlite
  store (`pkg/server/store/credential`, `pkg/server/store/sshkey`) and rendered back
  through formatters — never log a secret or echo one in an API response.
- **Websocket handlers** (`pkg/server/store/websocket/`: kubectl, ssh) stream live
  sessions to cluster nodes; keep their input handling strict.
- **`pkg/server/proxy`** forwards `/meta/proxy` and `/k8s/proxy` requests to other
  hosts — an SSRF surface; do not widen the set of reachable targets.
- **Kubeconfig download** (`download-kubeconfig` action) outputs cluster admin
  credentials; treat its output handling as secret handling.

## Tests

- Unit tests live next to the code as `*_test.go` and run with the `test` build tag
  (`make autok3s unit only`).
- Integration tests are ginkgo suites under `test/integration` and exercise the
  compiled binary; `test/e2e` is currently disabled in the make script.
- No network and no real cloud credentials in unit tests — mock at the provider SDK
  boundary (`pkg/k3s/command_test.go`, `pkg/cluster/kube` and
  `pkg/providers/aws/template_test.go` show the pattern).
- Add or update tests for any behavior change. A bug fix without a regression test
  is incomplete.

## Git and pull requests

- Commit messages follow **Conventional Commits** with a scope drawn from the area
  touched: `feat(server): …`, `fix(k3d): …`, `refactor(cluster): …`, `docs: …`,
  `chore(deps): …`. The subject is lowercase, imperative, and says what changed in
  the product — not what files moved.
- Focus commit messages on the problem, solution, and relevant context; put
  validation details in the PR description.
- **Do not add `Co-Authored-By` or other agent-attribution trailers.**

## Notes

If something here contradicts what the code actually does, the code wins — and fixing
this file is part of the change.

