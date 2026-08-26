# Linux/amd64 PGO pilot

This is an exploratory Linux/amd64 measurement harness, not a release build path.

It is Workstream 7's final research gate. Do not set `THREADFIN_PGO_FINAL_COMMIT` or run capture until accepted Workstreams 1–6 are complete, rejected candidates are reverted, their verification passes, no modernization change is pending, and `HEAD` is the reviewed final clean modernization commit.

Setting `THREADFIN_PGO_FINAL_COMMIT` is the operator's explicit assertion that this prerequisite is satisfied. The script verifies its 40-character value against `HEAD` for capture, both sessions, and comparison; the ignored run root records the same commit.

Run capture and sessions only on an otherwise idle, fixed-frequency or consistently governed host. Record kernel, CPU model, logical CPU count, Go build metadata, binary/profile/fixture hashes, guide start, commit, and run order in local evidence.

Run the exact sequence:

```bash
export THREADFIN_PGO_FINAL_COMMIT=$(git rev-parse HEAD)
RUN_ROOT=$(tools/pgo/run-linux-amd64.sh capture)
SESSION_1=$(tools/pgo/run-linux-amd64.sh session "$RUN_ROOT" session-1)
sleep 600
SESSION_2=$(tools/pgo/run-linux-amd64.sh session "$RUN_ROOT" session-2)
SUMMARY=$(tools/pgo/run-linux-amd64.sh compare "$RUN_ROOT" "$SESSION_1" "$SESSION_2")
printf '%s\n' "$SUMMARY"
```

The ten-minute gap is a manual thermal/load separation, not proof of independence.

Paired medians, MAD, and direction counts are descriptive. The practical thresholds are a screen for a separate review, not statistical significance.

`NO_ADOPT` means no build change. `PILOT_SUPPORTS_SEPARATE_OPERATIONAL_REVIEW` also means no build change; it permits only a new design/review request.

The profile is local evidence, not a repository or release asset. This pilot does not approve its custody, distribution, or use.

A future operational proposal must define authenticated artifact storage, access control, ownership, retention, provenance validation, and release fallback before any build integration.

A prospective profile is stale when it is older than 90 days; Go version or `GOAMD64` changes; fixture schema/cardinality/concurrency changes; the workload mix changes; or `src/provider.go`, `src/m3u.go`, `src/data.go`, `src/xepg.go`, `src/webserver.go`, or `src/buffer.go` changes materially.

Every refresh must recapture and rerun both five-pair sessions. Missing, stale, inaccessible, or unrepresentative evidence means explicit `-pgo=off` remains the only recommendation.

`.pgo-pilot/` may contain binaries, profiles, generated logs, result JSON, session JSONL, and summaries. None may be committed.
