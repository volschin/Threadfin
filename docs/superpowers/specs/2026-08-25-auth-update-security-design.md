# Authentication and Update Integrity Design

## Goal

Close the two high-risk findings from the 2026-08-25 security review without forcing password resets or removing Threadfin's self-update feature.

## Scope

- Replace fast password verifiers with Argon2id for new, changed, and successfully authenticated legacy credentials.
- Authenticate every self-update artifact with an Ed25519 signature and verify its SHA-256 digest before replacing the running executable.
- Generate and publish the required integrity metadata in the release workflow.
- Preserve existing usernames, authorization data, API behavior, update URLs, and supported operating systems.

Cookie authentication redesign, general request-size limits, updater key rotation, and unrelated authentication concurrency changes are outside this change.

## Password Storage

### New verifier format

Passwords are stored in the existing per-user `_password` field as a self-describing PHC string:

```text
$argon2id$v=19$m=19456,t=2,p=1$<base64-salt>$<base64-hash>
```

The implementation uses `golang.org/x/crypto/argon2.IDKey` with:

- 19,456 KiB memory
- 2 iterations
- 1 thread
- 16 random salt bytes from `crypto/rand`
- 32 derived-key bytes

These parameters meet OWASP's current minimum Argon2id recommendation while remaining practical on the low-resource systems Threadfin supports. PHC parsing is strict and rejects unsupported versions, zero or excessive parameters, invalid base64, incorrect salt length, and incorrect key length before invoking Argon2.

### Compatibility and migration

The `_password` value is authoritative:

- Values beginning with `$argon2id$` are verified as Argon2id.
- Existing values without that prefix are verified with the legacy HMAC-SHA256 behavior.
- A successful legacy verification immediately creates an Argon2id verifier and saves it before issuing a session token.
- A failed database save fails authentication and does not issue a token, so a login cannot appear migrated when it was not persisted.
- Incorrect credentials never modify the database.

New users and password changes write Argon2id directly. Existing `_username`, `_salt`, `_id`, authorization data, and username lookup behavior remain unchanged. The unused top-level `hash` metadata is set to `argon2id` for newly created databases but is not used to select a verifier for existing databases.

### Error behavior

Malformed Argon2id strings are treated as non-matching credentials, not as legacy values and not as a reason to panic. Random-source failures and persistence failures are returned to the caller. Comparisons of derived keys use constant-time comparison.

## Authenticated Self-Updates

### Release artifact contract

For every release binary named `<asset>`, the release publishes:

```text
<asset>
<asset>.sha256
<asset>.sha256.sig
```

`<asset>.sha256` contains exactly the lowercase hexadecimal SHA-256 digest followed by a newline. `<asset>.sha256.sig` is the raw 64-byte Ed25519 signature over the exact bytes of the checksum file.

The updater derives both sidecar URLs from the selected binary or ZIP URL. URL query parameters are preserved, and the sidecar suffix is added to the URL path. GitHub releases and custom update servers use the same contract. Missing or unsigned sidecars fail closed.

### Verification and replacement order

The updater performs these operations in order:

1. Fetch the checksum and signature sidecars with an HTTP client timeout and small response-size limits.
2. Verify the detached signature using the public key embedded in the binary.
3. Strictly parse the signed SHA-256 digest.
4. Download the candidate artifact to a temporary file in the executable directory while computing SHA-256.
5. Compare the downloaded digest in constant time with the signed digest.
6. For ZIP updates, extract the verified archive into a temporary directory and resolve the expected binary.
7. Only after every verification succeeds, begin the existing executable replacement and restart sequence.

Every failure before step 7 removes temporary material and leaves the current executable untouched. Signature verification occurs before trusting the checksum, and checksum verification occurs before trusting the downloaded artifact.

### Signing key management

An Ed25519 key pair is generated once:

- Only the raw public key is base64-encoded and committed in the updater package.
- The PEM private key is stored in the GitHub Actions secret `THREADFIN_UPDATE_SIGNING_KEY`.
- An owner-only recovery copy is stored outside the repository at `~/.config/threadfin/update-signing-key.pem` with mode `0600`.
- The private key is never printed, passed as a command-line argument, written into the repository, or included in build artifacts.

The release job reconstructs the private key in a mode-`0600` temporary file, signs each checksum, and deletes the temporary file through a shell trap. A missing signing secret fails the release rather than publishing unsigned update artifacts.

## Tests

Authentication tests cover:

- New users receive an Argon2id password verifier.
- Correct and incorrect Argon2id passwords are distinguished.
- A successful legacy login persists an Argon2id replacement and still authenticates after reloading the database.
- An incorrect legacy password does not migrate the verifier.
- Malformed and excessive PHC parameters fail safely.
- Password changes write Argon2id.

Updater tests use generated test keys and local HTTP servers to cover:

- Valid signature plus matching digest succeeds.
- Missing or invalid signatures fail.
- A validly signed but malformed checksum fails.
- A signed checksum mismatch fails and removes the candidate.
- Query-bearing update URLs produce correct sidecar URLs.
- Verification failures occur before executable replacement helpers are called.

Project verification includes unit tests, race tests, vet, `govulncheck`, workflow linting, vendoring reproducibility, all existing cross-build targets, and standard container builds.

## Acceptance Criteria

- No plaintext password is required outside the request performing authentication or a password change.
- No new or changed password is stored with the legacy fast hash.
- Existing valid credentials migrate without a forced reset.
- A self-update cannot replace the executable unless its checksum metadata is signed by the embedded key and its bytes match that signed digest.
- Release workflows cannot publish update binaries without matching signed sidecars.
- The signing private key remains outside Git and build artifacts.

## References

- OWASP Password Storage Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html
- Go Argon2 package: https://pkg.go.dev/golang.org/x/crypto/argon2
- Go Ed25519 package: https://pkg.go.dev/crypto/ed25519
