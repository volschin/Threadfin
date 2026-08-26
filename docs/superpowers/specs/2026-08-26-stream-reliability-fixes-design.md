# Stream Delivery and Reliability Fixes Design

**Status:** Approved for implementation planning
**Date:** 2026-08-26
**Target:** M3U HTTP delivery, third-party buffer failover, and playlist filter evaluation

## Context

An audit of the upstream Threadfin branches and open pull requests identified three defects that remain present in this fork:

1. An authenticated request for the complete M3U playlist skips the existing-file path because authentication is supplied through query parameters. A full playlist rebuild writes the playlist to disk and returns an empty string, so the request can receive an empty successful response.
2. A third-party buffer restarts segment numbering at `1.ts` after switching to a backup source. Connected clients remember delivered segment filenames and suppress reused names, which can cause a blackout proportional to the number of segments already watched.
3. Filter rules are constructed by ranging over a Go map, so their evaluation order changes nondeterministically. A filter whose include or exclude condition rejects a stream also terminates evaluation instead of allowing a later filter to accept it, causing silent channel loss.

The upstream changes are design inputs, not patches to cherry-pick. This fork has additional error handling and refactoring that must be preserved.

## Goals

1. Return the complete cached or newly generated M3U file for authenticated requests while preserving group-filtered playlist behavior.
2. Keep third-party buffer segment filenames monotonic across primary-to-backup and backup-to-backup transitions for the lifetime of a connected stream.
3. Make playlist filter evaluation deterministic and prevent one filter's failed conditions from vetoing later filters.
4. Add focused regression tests for each defect before changing production behavior.
5. Preserve existing authentication, playlist, buffering, configuration, and persisted-data contracts outside these defects.

## Non-goals

- Changing authentication methods, credential formats, or authorization policy.
- Changing M3U contents, channel ordering, stream URLs, or `group-title` syntax.
- Redesigning the buffering architecture, segment size policy, backup selection order, or client lifecycle.
- Adding configurable segment counters or filter-order settings.
- Delaying XEPG cleanup or adding missing-channel persistence counters.
- Adopting unrelated upstream FFmpeg, UDPxy, EPG, or channel-limit changes.
- Changing TypeScript, browser assets, dependencies, versions, or release metadata.

## Design Principles

### Preserve functional query semantics

Authentication parameters do not change playlist contents. `group-title` does. The M3U handler must therefore decide between complete-file and filtered-response behavior from `group-title`, not from the total number of query parameters.

### Never reuse a delivered segment name

The buffer directory is only a partial record because previously delivered files may have been removed. `OldSegments` is only a partial record because the newest files may not have been delivered. Backup numbering must consider both sources before choosing the next filename.

### A filter may accept, decline, or not match

A matched filter whose conditions fail declines that stream for that filter only. Evaluation continues until another filter accepts the stream or all filters have been exhausted. The first accepting filter still wins.

## Workstream 1: Authenticated Complete-M3U Delivery

### Current failure

The M3U handler authenticates the request and then checks `len(r.URL.Query()) == 0` before serving `threadfin.m3u`. Requests using `username` and `password` therefore skip an existing cached file.

For a complete playlist, `buildM3U` streams output to `threadfin.m3u` and returns an empty response string. The handler subsequently writes that empty string to the HTTP response instead of serving the file it just created.

### Required behavior

After successful authentication, the handler reads the `group-title` value once:

- If `group-title` is empty, the request is for the complete playlist.
  - If `threadfin.m3u` already exists as a regular readable file, serve it immediately.
  - Otherwise call `buildM3U` with no groups.
  - After a successful build, serve the generated `threadfin.m3u` file.
  - Authentication and unknown non-functional query parameters must not force a rebuild.
- If `group-title` is non-empty, split it using the existing comma-separated behavior, build the filtered playlist in memory, and return that content. The complete cached file must not be served for this request.

Authentication remains the first gate. An unauthenticated or incorrectly authenticated request must never receive either the cached file or generated content.

### Error behavior

- A missing complete playlist triggers generation; other file inspection errors are returned through the existing HTTP error path rather than being treated as absence.
- A `buildM3U` failure must not produce an empty successful response or serve a partial file.
- If generation reports success but the resulting file cannot be opened or served, the request must fail through the existing HTTP error path.
- Filtered-playlist generation errors retain the existing error-reporting policy but must not write successful empty content.

The implementation must preserve the existing successful response headers and download filename behavior for each path unless a regression test demonstrates that the current headers are internally inconsistent.

### Tests

Handler-level tests cover:

1. A complete request without authentication parameters serves an existing file without rebuilding it.
2. A complete request with valid `username` and `password` parameters serves the same existing file.
3. Other non-functional query parameters do not force a complete-playlist rebuild.
4. A `group-title` request returns only the requested groups and does not serve the complete cached file.
5. A missing complete file is generated and the generated bytes are returned in the same request.
6. Failed authentication cannot read an existing cached file.
7. Build, file inspection, and post-build file access failures do not return HTTP 200 with an empty body.

Tests must use temporary data directories and must not depend on production configuration or credentials.

## Workstream 2: Monotonic Segment Numbering Across Failover

### Current failure

`thirdPartyBuffer` removes the buffer directory and initializes `tmpSegment` to 1 for every invocation, including backup invocations. `getBufTmpFiles` tracks delivered segments by basename in `stream.OldSegments` and skips any basename already present there. Reissuing `1.ts`, `2.ts`, and subsequent names after failover therefore suppresses valid backup data for an already connected client.

### Required behavior

Primary and backup startup have distinct policies:

- A new primary buffer retains the existing behavior: clear its buffer directory and start at `1.ts`.
- A backup buffer must not clear the existing buffer directory.
- Before creating the first backup segment, determine the highest valid numeric `.ts` basename found in:
  - the retained buffer directory; and
  - `stream.OldSegments`.
- Start the backup at one greater than that maximum. If neither source contains a valid segment name, start at `1.ts`.
- Apply the same rule to every later backup transition, so names remain monotonic across backup 1, backup 2, and backup 3.

Only basenames matching the existing numeric segment convention are considered. Directories, unrelated files, malformed names, negative values, and numeric overflow must not influence the counter.

If the highest valid segment number cannot be incremented without overflowing the counter type, fail the backup attempt through the existing fallback path rather than wrapping or reusing a filename.

The invocation's chosen first segment is recorded as `startSegment`. Existing logic that currently uses literal segment `1` to detect “no segment completed yet” or the first completed segment must compare against `startSegment` instead. This includes the startup timeout and the one-time buffering-status transition.

### Directory and error behavior

- If the backup directory does not exist, create it and derive the next number from `OldSegments`.
- If an existing backup directory cannot be read, fail that backup attempt through the existing buffer error/fallback path. Do not clear it and restart numbering at 1 because unseen filenames could then be reused.
- Errors creating the selected segment retain the existing process termination, client notification, and next-backup behavior.
- Segment-number discovery must not modify or truncate retained files.

No persistent counter is added. Segment numbering only needs to remain unique for the lifetime of the active stream and its connected clients.

### Tests

Focused buffer/VFS tests cover:

1. A primary invocation starts at `1.ts` after clearing stale files.
2. A backup with retained `1.ts` through `5.ts` starts at `6.ts`.
3. A backup whose retained directory starts later than `OldSegments` uses the directory maximum.
4. A backup whose delivered history is later than the retained directory uses the `OldSegments` maximum.
5. A missing backup directory still continues after the maximum delivered segment.
6. Backup 2 continues after files produced by backup 1.
7. Malformed and unrelated directory entries are ignored without lowering or resetting the counter.
8. An unreadable existing directory fails instead of reusing a segment name.
9. Startup timeout and first-segment status behavior work when `startSegment` is greater than 1.
10. `getBufTmpFiles` returns newly completed backup segments to a client whose `OldSegments` contains the primary segment names.

Tests must exercise the configured virtual filesystem boundary without spawning a real FFmpeg or VLC process where a narrower deterministic test is sufficient.

## Workstream 3: Deterministic, Non-vetoing Filter Evaluation

### Current failure

`Settings.Filter` is a map keyed by the filter IDs assigned by the UI. `createFilterRules` ranges over that map directly, creating `Data.Filter` in a nondeterministic order.

When a filter's base rule matches but an include or exclude condition fails, `filterThisStream` returns `false` immediately. A stream is therefore rejected before a later applicable filter is evaluated. Conditions-only custom filters amplify this problem: after their condition expression is extracted, their base rule is empty, and an empty substring matches every stream.

### Required behavior

`createFilterRules` must:

1. Collect the keys from `Settings.Filter`.
2. Sort them in ascending numeric order.
3. Decode and append rules in that order.

This order matches the stable ID order already represented by the configuration. No new priority field or persisted format is introduced.

`filterThisStream` must preserve the existing first-acceptance model:

- A filter whose base rule does not match proceeds to the next filter.
- A matching filter whose exclude condition fails proceeds to the next filter.
- A matching filter whose include condition fails proceeds to the next filter.
- A matching filter whose conditions pass immediately accepts the stream and returns that filter's `LiveEvent` value.
- If no filter accepts the stream, the stream remains inactive.

A custom filter consisting only of include or exclude conditions remains valid: after extracting its conditions, its empty base rule intentionally considers every stream, and the conditions determine whether that filter accepts it. A genuinely empty configured filter remains ignored according to existing behavior.

The change must not alter case-sensitivity, group-title matching, condition syntax, channel ordering, category assignment, or filter persistence.

### Tests

Table-driven tests cover:

1. Repeated construction from the same map always produces ascending filter-ID order.
2. A stream rejected by an earlier filter's include condition can be accepted by a later filter.
3. A stream rejected by an earlier filter's exclude condition can be accepted by a later filter.
4. The first filter whose rule and conditions pass still wins.
5. A conditions-only custom filter accepts matching streams and declines non-matching streams without vetoing later filters.
6. A genuinely empty custom filter remains ignored.
7. Case-sensitive and case-insensitive rules preserve their existing behavior.
8. `LiveEvent` comes from the filter that accepts the stream, not from an earlier declined filter.
9. Repeated database builds with identical settings and input produce identical active and inactive stream sets.

## Implementation Boundaries

Expected production changes are limited to:

- `src/webserver.go` for complete versus filtered M3U response selection;
- `src/buffer.go` for failover segment-number discovery and `startSegment` comparisons;
- `src/data.go` for deterministic filter construction; and
- `src/m3u.go` for condition fallthrough.

Tests may be added to existing colocated test files or new `_test.go` files in `src/`. A small unexported helper is acceptable where it isolates segment-name parsing or makes handler behavior testable, but no broader abstraction is required.

No dependency, configuration-schema, generated asset, or persisted-data change is expected.

## Verification

Format changed Go files and run:

```sh
gofmt -w <changed-go-files>
go test -count=1 -mod=vendor ./...
go vet -mod=vendor ./...
git diff --check
```

The changes are platform-neutral. Run the repository-required build checks for Linux `amd64` and `arm64`; no additional platform-specific behavior is introduced.

## Commit Strategy

Implement the workstreams as three independently reviewable production commits, each preceded or accompanied by its focused regression tests:

1. Fix authenticated complete-M3U delivery.
2. Preserve segment numbering across backup failover.
3. Make filter evaluation deterministic and non-vetoing.

Do not combine unrelated cleanup or upstream changes with these commits. Each commit must pass the complete Go test and vet gates.

## Acceptance Criteria

- A valid authenticated complete-M3U request returns the cached or newly generated playlist bytes, never an empty successful body caused by disk-streaming generation.
- A `group-title` request continues to return only the requested groups even when a complete cached playlist exists.
- No segment basename previously delivered to a connected client is reused after any backup transition.
- Backup data becomes eligible for delivery as soon as its first segment is complete; blackout duration does not scale with viewing duration.
- Identical filters and playlist input produce identical active and inactive stream sets across repeated builds.
- A filter whose conditions fail cannot prevent a later filter from accepting the stream.
- Existing authentication failures, backup ordering, filter syntax, and persisted formats remain unchanged.
- Full tests, vet, formatting, diff checks, and required Linux builds pass with the checked-in vendor tree.

## References

- [Threadfin upstream branch `branch-1.2.40`](https://github.com/Threadfin/Threadfin/tree/branch-1.2.40)
- [Upstream PR #662: continue segment numbering across backup switches](https://github.com/Threadfin/Threadfin/pull/662)
- [Upstream PR #664: deterministic filter evaluation and condition fallthrough](https://github.com/Threadfin/Threadfin/pull/664)
