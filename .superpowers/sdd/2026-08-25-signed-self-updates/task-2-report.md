# Task 2 Report: Update Integrity Primitive

## Scope

- Added the embedded Ed25519 public-key parser, sidecar URL construction, bounded metadata retrieval, signed checksum verification, and verified artifact download.
- Added the task-specified URL, signature, checksum-format, successful-download, and mismatch-cleanup tests.
- Reused the tracked `update-signing-public-key.pem` from Task 1 without modifying it. No private signing key was accessed.

## TDD Evidence

### RED 1: URL and signature tests

Command:

```text
GOTOOLCHAIN=go1.27.0 go test -mod=vendor ./src/internal/up2date/client
```

Exit status: `1`

Relevant output:

```text
src/internal/up2date/client/integrity_test.go:34:14: undefined: sidecarURL
src/internal/up2date/client/integrity_test.go:54:14: undefined: fetchExpectedChecksum
src/internal/up2date/client/integrity_test.go:73:15: undefined: fetchExpectedChecksum
src/internal/up2date/client/integrity_test.go:94:15: undefined: fetchExpectedChecksum
src/internal/up2date/client/integrity_test.go:108:15: undefined: fetchExpectedChecksum
FAIL threadfin/src/internal/up2date/client [build failed]
```

The compiler also reported the imports reserved for the second task-specified test batch as unused.

### RED 2: verified-download tests

Command:

```text
GOTOOLCHAIN=go1.27.0 go test -mod=vendor ./src/internal/up2date/client
```

Exit status: `1`

Relevant output:

```text
src/internal/up2date/client/integrity_test.go:120:12: undefined: downloadVerified
src/internal/up2date/client/integrity_test.go:139:12: undefined: downloadVerified
FAIL threadfin/src/internal/up2date/client [build failed]
```

The compiler also reported the `subtle` and `os` imports reserved for the pending implementation as unused.

### GREEN: focused package

Commands:

```text
gofmt -w src/internal/up2date/client/integrity.go src/internal/up2date/client/integrity_test.go
GOTOOLCHAIN=go1.27.0 go test -mod=vendor ./src/internal/up2date/client
```

Exit status: `0`

Output:

```text
ok  threadfin/src/internal/up2date/client  0.008s
```

### GREEN: full repository

Command:

```text
GOTOOLCHAIN=go1.27.0 go test -mod=vendor ./...
```

Exit status: `0`

All packages passed; packages without tests were reported as such.

## Self-review

- Sidecar suffixes are applied to the URL path while preserving queries, and only HTTP(S) schemes are accepted.
- Metadata responses require HTTP 200 and are bounded to 1024 bytes for checksums and 64 bytes for signatures.
- Signatures cover the exact checksum bytes before checksum parsing.
- Checksums require exactly 64 lowercase hexadecimal characters followed by one newline.
- Artifact downloads require HTTP 200, use exclusive creation with mode `0600`, stream through SHA-256, compare in constant time, and remove the destination on copy, close, or checksum failure.
- No unrelated source, plan, or public-key changes were made.
