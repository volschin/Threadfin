#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

ROOT=$(git rev-parse --show-toplevel)
cd "$ROOT"

fail() { printf 'NO_ADOPT: %s\n' "$*" >&2; exit 1; }
need() { command -v "$1" >/dev/null 2>&1 || fail "missing command: $1"; }
sha() { sha256sum "$1" | cut -d' ' -f1; }
require_session_path() {
  local out=$1 supplied=$2 session_root resolved
  session_root=$(realpath -e -- "$out/sessions") || fail "missing sessions directory under run root: $out"
  resolved=$(realpath -e -- "$supplied") || fail "missing session: $supplied"
  case "$resolved" in
    "$session_root"/*) printf '%s\n' "$resolved" ;;
    *) fail "session path is outside run root: $supplied" ;;
  esac
}

preflight() {
  local final_commit
  test "$(uname -s)" = Linux || fail "pilot requires Linux"
  test "$(uname -m)" = x86_64 || fail "pilot requires amd64"
  test "$(go env GOOS)" = linux || fail "go env GOOS is not linux"
  test "$(go env GOARCH)" = amd64 || fail "go env GOARCH is not amd64"
  case "$(go version)" in *' go1.27.'*) ;; *) fail "Go 1.27 is required";; esac
  test -z "$(git status --porcelain)" || fail "working tree must be clean"
  final_commit=${THREADFIN_PGO_FINAL_COMMIT:-}
  [[ "$final_commit" =~ ^[0-9a-f]{40}$ ]] || fail "THREADFIN_PGO_FINAL_COMMIT must assert the reviewed final Workstreams 1-6 commit"
  test "$final_commit" = "$(git rev-parse HEAD)" || fail "HEAD does not match THREADFIN_PGO_FINAL_COMMIT"
  for command in go git ffmpeg sha256sum cmp grep realpath stat uname; do need "$command"; done
  cpus=$(getconf _NPROCESSORS_ONLN)
  test "$cpus" -ge 4 || fail "at least four logical CPUs are required"
}

build_one() {
  local pgo=$1 output=$2
  env -u GOFLAGS CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v1 \
    go build -buildvcs=false -mod=vendor -tags=pgopilot -trimpath \
      -ldflags='-s -w' -pgo="$pgo" -o "$output" .
}

validate_profile() {
  local binary=$1 profile=$2 report=$3
  test -s "$profile" || return 1
  go tool pprof -top -nodefraction=0 -nodecount=0 "$binary" "$profile" >"$report"
  grep -Eq 'getProviderData|MakeInterfaceFromM3U|buildDatabaseDVR' "$report" || return 1
  grep -Eq 'createXEPGMapping|createXEPGDatabase|mapping|createXMLTVFile|writeXEPGOutputFiles' "$report" || return 1
  grep -Eq 'Stream|bufferingStream|thirdPartyBuffer' "$report" || return 1
}

capture() {
  preflight
  local stamp out guide_start attempt result profile report
  stamp=$(date -u +%Y%m%dT%H%M%SZ)
  out="$ROOT/.pgo-pilot/$stamp"
  mkdir -p "$out"/{bin,capture,profile,sessions}
  git rev-parse HEAD >"$out/commit"
  env -u GOFLAGS GOOS=linux GOARCH=amd64 GOAMD64=v1 CGO_ENABLED=0 \
    go build -buildvcs=false -mod=vendor -trimpath -o "$out/bin/pgo-pilot" ./tools/pgo
  build_one off "$out/bin/threadfin-off-a"
  build_one off "$out/bin/threadfin-off-b"
  cmp "$out/bin/threadfin-off-a" "$out/bin/threadfin-off-b" || fail "off builds are not reproducible"
  guide_start=$(date -u -d 'tomorrow 00:00' +%Y-%m-%dT%H:%M:%SZ)

  for attempt in 1 2 3; do
    result="$out/capture/result-$attempt.json"
    profile="$out/profile/cpu-$attempt.pprof"
    report="$out/profile/top-$attempt.txt"
    if "$out/bin/pgo-pilot" run -variant capture \
      -binary "$out/bin/threadfin-off-a" -output "$result" -profile "$profile" \
      -pair 0 -sequence "$attempt" -stream-duration 90s -clients 32 \
      -sample-bytes 2097152 -guide-start "$guide_start" && \
      validate_profile "$out/bin/threadfin-off-a" "$profile" "$report"; then
        mv "$result" "$out/capture/result.json"
        mv "$profile" "$out/profile/cpu.pprof"
        mv "$report" "$out/profile/top.txt"
        break
    fi
    rm -f "$result" "$profile" "$profile.tmp"
  done
  test -s "$out/profile/cpu.pprof" || fail "representative profile capture was not practical after three attempts"
  build_one "$out/profile/cpu.pprof" "$out/bin/threadfin-pgo-a"
  build_one "$out/profile/cpu.pprof" "$out/bin/threadfin-pgo-b"
  cmp "$out/bin/threadfin-pgo-a" "$out/bin/threadfin-pgo-b" || fail "PGO builds are not reproducible"
  printf '%s\n' "$out"
}

session() {
  preflight
  local out=$1 name=$2 target tmp guide_start pair first second seq=0 result
  test -x "$out/bin/pgo-pilot" || fail "missing pilot tool under $out"
  test -x "$out/bin/threadfin-off-a" || fail "missing off binary under $out"
  test -x "$out/bin/threadfin-pgo-a" || fail "missing PGO binary under $out"
  test -s "$out/profile/cpu.pprof" || fail "missing capture profile under $out"
  test "$(cat "$out/commit")" = "$(git rev-parse HEAD)" || fail "run root was built from a different commit"
  target="$out/sessions/$name.jsonl"
  tmp="$target.tmp"
  test ! -e "$target" && test ! -e "$tmp" || fail "session already exists: $name"
  guide_start=$(date -u -d 'tomorrow 00:00' +%Y-%m-%dT%H:%M:%SZ)
  : >"$tmp"
  trap 'rm -f "$tmp"' RETURN
  for pair in 1 2 3 4 5; do
    if (( pair % 2 == 1 )); then first=off; second=pgo; else first=pgo; second=off; fi
    for variant in "$first" "$second"; do
      seq=$((seq+1))
      result="$out/sessions/$name-pair-$pair-$variant.json"
      "$out/bin/pgo-pilot" run -variant "$variant" \
        -binary "$out/bin/threadfin-$variant-a" -output "$result" \
        -pair "$pair" -sequence "$seq" -stream-duration 30s -clients 32 \
        -sample-bytes 2097152 -guide-start "$guide_start"
      cat "$result" >>"$tmp"
      rm -f "$result"
    done
  done
  mv "$tmp" "$target"
  trap - RETURN
  printf '%s\n' "$target"
}

compare() {
  preflight
  local out=$1 first second summary
  first=$(require_session_path "$out" "$2")
  second=$(require_session_path "$out" "$3")
  summary="$out/summary.json"
  "$out/bin/pgo-pilot" compare -session "$first" -session "$second" -output "$summary"
  printf '%s\n' "$summary"
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  case "${1:-}" in
    capture) test "$#" -eq 1 || fail 'usage: run-linux-amd64.sh capture'; capture ;;
    session) test "$#" -eq 3 || fail 'usage: run-linux-amd64.sh session RUN_ROOT NAME'; session "$2" "$3" ;;
    compare) test "$#" -eq 4 || fail 'usage: run-linux-amd64.sh compare RUN_ROOT SESSION_1 SESSION_2'; compare "$2" "$3" "$4" ;;
    *) fail 'usage: run-linux-amd64.sh capture|session|compare' ;;
  esac
fi
