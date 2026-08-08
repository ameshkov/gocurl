# AGENTS.md

Single-binary Go CLI (`gocurl`, a curl-like tool with ECH, encrypted DNS,
OHTTP, and WebSocket extras). Module `github.com/ameshkov/gocurl`, Go 1.26.5 —
the local toolchain must match `go.mod`. All code is under `internal/`; there is
no exported API or second module.

## Commands

- `make init` — installs the git pre-commit hook (it is NOT installed by
  default on clone). The hook blocks commits containing `FIXME` or `TODO!!` in
  staged non-doc files and then runs `make check`.
- `make check` — the CI quality gate: `make lint` then `make test`. This is what
  the pre-commit hook and CI run; mirror it before finishing work.
- `make lint` — `npx markdownlint-cli .` (needs Node.js; config in
  `.markdownlint.json`) then `golangci-lint run`.
- `make test` — `go test -race -v -bench=. ./...` (runs benchmarks too, slow).
  For a fast iteration loop use `go test ./internal/...` or
  `go test -run <TestName> ./internal/cmd`.
- `make build` / `make release` — release requires `GOOS`, `GOARCH`, `VERSION`
  env vars (see CI `build.yaml` for the full cross-compile matrix).
- `golangci-lint run --fix` applies the configured formatters.

## Formatting (lint fails otherwise)

- `.golangci.yml` (golangci-lint v2) enables the `gofumpt`, `goimports`, and
  `golines` formatters with `max-len: 120`. Keep lines <=120 chars and
  gofumpt-formatted; there is no standalone gofmt step.
- AdGuard-team doc-comment style (`// Package name ...`, sentence-per-paragraph
  style) is used throughout; match existing files.
- Any markdown you touch (including this file) must pass markdownlint: dashed
  list markers, 4-space nested indentation, asterisk emphasis; line-length and
  bare-URL checks are disabled.

## Architecture

- Entry flow: `main.go` -> `internal/cmd.Main()` -> `config.ParseConfig` ->
  `output.NewOutput` -> `cmd.Run(cfg, out)`. `Run` is extracted purely for
  testability — tests call it directly, they never exec the binary.
- Two config layers: raw go-flags tags live in `internal/config/options.go`
  (`Options`); `ParseConfig` validates them into the typed `Config` in
  `internal/config/config.go`. To add a CLI flag: add a tagged field in
  `Options`, plumb it in `ParseConfig`, and update the "All command-line
  arguments" block in `README.md`, which mirrors the `--help` text generated
  from those tags.
- `internal/client` — HTTP transport, dialer, proxy, plus feature packages:
  `ohttp`, `websocket`, `splittls`, `cfcrypto` (ECH), `connectto`.
  `internal/resolve` owns DNS resolution and ECH-config discovery;
  `internal/output` writes results; `internal/appversion` holds the linker-set
  version.
- `cmd.Run` hand-rolls response body-readability with a large conditional, and
  special-cases WebSocket upgrade. Read that function before touching response
  handling.

## Testing

- Command-level tests (`internal/cmd/cmd_test.go`) spin up a local
  `httptest.NewServer(httpbin.New().Handler())` and call
  `cmd.Run(cfg, out)` with `config.ParseConfig(args)` plus a buffer-backed
  `output.NewOutputWithWriters(...)`. New tests should be fully local and
  hermetic.
- Existing network-dependent exceptions (flaky/offline): the OHTTP tests in
  `cmd_test.go` hit `httpbin.agrd.workers.dev`, and `resolve_test.go` does real
  DNS lookups for `example.org` / `cloudflare-ech.com` (its header has a TODO to
  remove that dependency). Don't be surprised when these fail without network.

## Releases

- Version is injected at link time via
  `-ldflags "-X github.com/ameshkov/gocurl/internal/appversion.version=$(VERSION)"`
  (Makefile and Dockerfile); `appversion.Version()` falls back to `v0.0` when
  unset.
- Version bumps are dedicated commits ("Bump version to vX.Y.Z") with a
  CHANGELOG.md entry (Keep a Changelog; update the `[Unreleased]` block).
- Releasing = push a `v*` tag. CI builds the cross-OS archive matrix and uploads
  it as a GitHub release, and builds/pushes a multi-arch Docker image (on tags
  and `master`).

## Guidelines

- When bumping the Go version, update it in `go.mod` first (source of truth),
  then the `golang:...` image tag in `Dockerfile` and the version mention in the
  intro of this file. CI workflows use `go-version: stable` and need no edit.
- Add new guidelines to this "Guidelines" section automatically whenever you are
  asked to add a guideline, or infer from the conversation that a correction
  deserves to be codified as one. Keep each guideline concise and actionable.
- Whenever a task is user-facing or otherwise important, add a brief description
  of it to the `[Unreleased]` block of `CHANGELOG.md`, following the Keep a
  Changelog format (e.g. under `### Added` / `### Changed` / `### Fixed`).
