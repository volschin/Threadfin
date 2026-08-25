# Repository working agreements

## Scope and change discipline

- Follow YAGNI: make the smallest change that satisfies the current task and a concrete risk.
- Preserve unrelated local changes. Do not rewrite generated or vendored files unless the task requires it.
- Keep secrets, tokens, signing keys, and private configuration out of source, logs, and command arguments.

## Go development

- Use the Go version declared in `go.mod` and build with the checked-in vendor tree.
- Format changed Go files with `gofmt`.
- Before committing Go changes, run:

  ```sh
  go test -count=1 -mod=vendor ./...
  go vet -mod=vendor ./...
  ```

- When dependencies change, run `go mod tidy` and `go mod vendor`, then verify that `go.mod`, `go.sum`, and `vendor/` contain only the intended update.
- Keep platform-specific updater behavior covered on its target platform. At minimum, cross-build Windows changes with `GOOS=windows GOARCH=amd64` and Linux changes for `amd64` and `arm64`.

## Web assets

- TypeScript source lives in `ts/`; compiled browser assets live under `html/`.
- After changing TypeScript, run `tsc -p ./ts/tsconfig.json`.
- When browser assets change, regenerate `src/webUI.go` using the development-mode procedure documented in `README.md`, and include the generated file in the same change.

## Releases

- Keep `Version` and `APIVersion` in `threadfin.go` synchronized for application releases unless a task explicitly changes the API versioning policy.
- Release tags use the `vMAJOR.MINOR.PATCH` form. Never move or overwrite a published tag.
- Push the release commit to `main` and wait for its CI run before creating the tag.
- A release is complete only after the tag workflow succeeds and the GitHub release, signed binary assets, and expected GHCR image manifests are verified.
