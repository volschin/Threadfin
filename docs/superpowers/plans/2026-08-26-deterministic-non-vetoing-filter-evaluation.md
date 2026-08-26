# Deterministic Non-vetoing Filter Evaluation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build stream filters in ascending numeric ID order and let later filters accept streams declined by earlier conditions.

**Architecture:** Sort only `Settings.Filter` keys consumed by `createFilterRules`, decoding every entry into fresh zeroed values. Keep `filterThisStream`'s first-acceptance model, but continue after failed conditions and return `LiveEvent` only from an accepting filter.

**Tech Stack:** Go 1.27, standard `encoding/json`, `sort`, `strings`, and `testing` packages

**Spec:** `docs/superpowers/specs/2026-08-26-stream-reliability-fixes-design.md`

## Global Constraints

- Sort stream-filter IDs numerically ascending only in `createFilterRules`; do not change `src/xepg.go` or `decodeFilters`.
- Decode each map entry into fresh `FilterStruct` and `Filter` values; custom filters must not inherit `LiveEvent`.
- Failed base, exclude, or include checks decline only current filter; first passing filter still wins.
- Return `LiveEvent` only from accepting filter; return `false, false` when none accepts.
- Preserve conditions-only custom filters, genuinely empty-rule skipping, case behavior, group-title equality, condition syntax, channel ordering, category assignment, persisted JSON, and error behavior.
- Preserve current `strings.SplitSeq` condition parsing and its separator compatibility tests.
- Do not use `t.Parallel`; tests mutate package globals.
- Do not apply upstream PR #664 wholesale or revert merged Go 1.27 modernization.
- Run all tests and builds with checked-in vendor tree.

## File Structure

- Modify `src/data.go`: deterministic, fresh-state `createFilterRules` construction.
- Modify `src/m3u.go`: non-vetoing evaluation and accepted-filter `LiveEvent` return.
- Create `src/filter_test.go`: focused construction/evaluation and repeated database-build coverage.
- Read-only context: `src/struct-system.go` defines `Filter`, `FilterStruct`, `SettingsStruct`, and `DataStruct`; `src/persistence_test.go` provides `restorePersistentState`.

---

### Task 1: Build filters in deterministic ID order with fresh state

**Files:**
- Modify: `src/data.go:774-819`
- Create: `src/filter_test.go`

**Interfaces:**
- Consumes: `Settings.Filter map[int64]interface{}`, `mapToJSON(interface{}) string`, `FilterStruct`, and `Filter`.
- Produces: deterministic `Data.Filter []Filter` in ascending numeric ID order.

- [ ] **Step 1: Write failing construction test**

```go
package src

import (
	"reflect"
	"testing"
)

func TestCreateFilterRulesSortsIDsAndResetsPerRuleState(t *testing.T) {
	restorePersistentState(t)

	Settings = SettingsStruct{}
	Data = DataStruct{}
	Settings.Filter = map[int64]interface{}{
		30: FilterStruct{Type: "custom-filter", Filter: "third", CaseSensitive: true},
		10: FilterStruct{Type: "group-title", Filter: "News", LiveEvent: true},
		20: FilterStruct{Type: "custom-filter", Filter: "second"},
	}

	want := []Filter{
		{LiveEvent: true, Rule: "News", Type: "group-title"},
		{Rule: "second", Type: "custom-filter"},
		{CaseSensitive: true, Rule: "third", Type: "custom-filter"},
	}

	for build := 0; build < 32; build++ {
		if err := createFilterRules(); err != nil {
			t.Fatalf("createFilterRules() build %d error = %v", build, err)
		}
		if !reflect.DeepEqual(Data.Filter, want) {
			t.Fatalf("createFilterRules() build %d filters = %#v, want %#v", build, Data.Filter, want)
		}
	}
}
```

This fails regardless of first map order: output is either unsorted or custom filters inherit earlier `LiveEvent` state.

- [ ] **Step 2: Run test and verify RED**

```sh
go test -count=1 -mod=vendor ./src -run '^TestCreateFilterRulesSortsIDsAndResetsPerRuleState$' -v
```

Expected: FAIL with nonascending `Data.Filter` or inherited `LiveEvent: true`.

- [ ] **Step 3: Implement sorted IDs and fresh decoded values**

```go
func createFilterRules() (err error) {
	Data.Filter = nil

	filterIDs := make([]int64, 0, len(Settings.Filter))
	for id := range Settings.Filter {
		filterIDs = append(filterIDs, id)
	}
	sort.Slice(filterIDs, func(i, j int) bool {
		return filterIDs[i] < filterIDs[j]
	})

	for _, id := range filterIDs {
		var dataFilter Filter
		var filter FilterStruct
		var exclude, include string

		err = json.Unmarshal([]byte(mapToJSON(Settings.Filter[id])), &filter)
		if err != nil {
			return
		}

		switch filter.Type {
		case "custom-filter":
			dataFilter.CaseSensitive = filter.CaseSensitive
			dataFilter.Rule = filter.Filter
			dataFilter.Type = filter.Type
			Data.Filter = append(Data.Filter, dataFilter)

		case "group-title":
			if len(filter.Include) > 0 {
				include = fmt.Sprintf(" {%s}", filter.Include)
			}
			if len(filter.Exclude) > 0 {
				exclude = fmt.Sprintf(" !{%s}", filter.Exclude)
			}
			dataFilter.CaseSensitive = filter.CaseSensitive
			dataFilter.LiveEvent = filter.LiveEvent
			dataFilter.Rule = fmt.Sprintf("%s%s%s", filter.Filter, include, exclude)
			dataFilter.Type = filter.Type
			Data.Filter = append(Data.Filter, dataFilter)
		}
	}

	return
}
```

`sort` is already imported in `src/data.go`. Do not sort or call `decodeFilters` from XEPG.

- [ ] **Step 4: Run test and verify GREEN**

Run same command. Expected: PASS for all 32 constructions.

- [ ] **Step 5: Commit task only when explicitly requested**

```sh
git add src/data.go src/filter_test.go
git commit -m "fix: order stream filters deterministically"
```

---

### Task 2: Continue after condition failures and bind `LiveEvent` to acceptance

**Files:**
- Modify: `src/m3u.go:47-148`
- Modify: `src/filter_test.go`
- Preserve: `src/m3u.go:151-185`

**Interfaces:**
- Consumes: `Data.Filter []Filter` and stream maps containing `group-title`, `name`, and `_values`.
- Produces: `filterThisStream(s interface{}) (status bool, liveEvent bool)` with first-acceptance semantics and `false, false` exhaustion.

- [ ] **Step 1: Add stream fixture helper and table-driven failing tests**

```go
func filterTestStream(group, name, values string) map[string]string {
	return map[string]string{
		"group-title": group,
		"name":        name,
		"_values":     values,
	}
}

func TestFilterThisStreamUsesFirstAcceptingFilter(t *testing.T) {
	restorePersistentState(t)

	tests := []struct {
		name          string
		filters       []Filter
		stream        map[string]string
		wantStatus    bool
		wantLiveEvent bool
	}{
		{
			name: "include condition failure falls through",
			filters: []Filter{
				{Type: "custom-filter", Rule: "{ota}", LiveEvent: true},
				{Type: "group-title", Rule: "ErsatzTV"},
			},
			stream: filterTestStream("ErsatzTV", "Ersatz News", `tvg-id="ersatz-1"`),
			wantStatus: true,
		},
		{
			name: "exclude condition failure falls through",
			filters: []Filter{
				{Type: "custom-filter", Rule: "!{blocked}", LiveEvent: true},
				{Type: "group-title", Rule: "ErsatzTV"},
			},
			stream: filterTestStream("ErsatzTV", "Ersatz News", `tvg-id="ersatz-1" blocked`),
			wantStatus: true,
		},
		{
			name: "declined filter cannot leak live event",
			filters: []Filter{
				{Type: "custom-filter", Rule: "{missing}", LiveEvent: true},
			},
			stream: filterTestStream("ErsatzTV", "Ersatz News", `tvg-id="ersatz-1"`),
		},
		{
			name: "first passing filter wins",
			filters: []Filter{
				{Type: "group-title", Rule: "ErsatzTV", LiveEvent: true},
				{Type: "group-title", Rule: "ErsatzTV"},
			},
			stream: filterTestStream("ErsatzTV", "Ersatz News", `tvg-id="ersatz-1"`),
			wantStatus: true, wantLiveEvent: true,
		},
		{
			name: "later accepting filter supplies live event",
			filters: []Filter{
				{Type: "custom-filter", Rule: "{missing}"},
				{Type: "group-title", Rule: "ErsatzTV", LiveEvent: true},
			},
			stream: filterTestStream("ErsatzTV", "Ersatz News", `tvg-id="ersatz-1"`),
			wantStatus: true, wantLiveEvent: true,
		},
		{
			name: "conditions-only include accepts matching stream",
			filters: []Filter{{Type: "custom-filter", Rule: "{ersatz}", LiveEvent: true}},
			stream: filterTestStream("ErsatzTV", "Ersatz News", `tvg-id="ersatz-1"`),
			wantStatus: true, wantLiveEvent: true,
		},
		{
			name: "conditions-only include declines nonmatching stream",
			filters: []Filter{{Type: "custom-filter", Rule: "{ota}"}},
			stream: filterTestStream("ErsatzTV", "Ersatz News", `tvg-id="ersatz-1"`),
		},
		{
			name: "conditions-only exclude accepts stream without excluded value",
			filters: []Filter{{Type: "custom-filter", Rule: "!{blocked}", LiveEvent: true}},
			stream: filterTestStream("ErsatzTV", "Ersatz News", `tvg-id="ersatz-1"`),
			wantStatus: true, wantLiveEvent: true,
		},
		{
			name: "genuinely empty custom filter remains ignored",
			filters: []Filter{{Type: "custom-filter", Rule: "", LiveEvent: true}},
			stream: filterTestStream("ErsatzTV", "Ersatz News", `tvg-id="ersatz-1"`),
		},
		{
			name: "case-insensitive custom rule still matches",
			filters: []Filter{{Type: "custom-filter", Rule: "SPORTS"}},
			stream: filterTestStream("Sports", "Sports News", `tvg-id="sports-1"`),
			wantStatus: true,
		},
		{
			name: "case-sensitive custom rule still declines different case",
			filters: []Filter{{Type: "custom-filter", Rule: "SPORTS", CaseSensitive: true}},
			stream: filterTestStream("Sports", "Sports News", `tvg-id="sports-1"`),
		},
		{
			name: "case-insensitive group rule still matches",
			filters: []Filter{{Type: "group-title", Rule: "SPORTS"}},
			stream: filterTestStream("Sports", "Sports News", `tvg-id="sports-1"`),
			wantStatus: true,
		},
		{
			name: "case-sensitive group rule still declines different case",
			filters: []Filter{{Type: "group-title", Rule: "SPORTS", CaseSensitive: true}},
			stream: filterTestStream("Sports", "Sports News", `tvg-id="sports-1"`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			Data.Filter = tc.filters
			gotStatus, gotLiveEvent := filterThisStream(tc.stream)
			if gotStatus != tc.wantStatus || gotLiveEvent != tc.wantLiveEvent {
				t.Fatalf("filterThisStream() = (%t, %t), want (%t, %t)", gotStatus, gotLiveEvent, tc.wantStatus, tc.wantLiveEvent)
			}
		})
	}
}
```

- [ ] **Step 2: Run evaluation test and verify RED**

```sh
go test -count=1 -mod=vendor ./src -run '^TestFilterThisStreamUsesFirstAcceptingFilter$' -v
```

Expected: include/exclude fallthrough and declined-`LiveEvent` cases fail.

- [ ] **Step 3: Implement minimal non-vetoing evaluation**

Remove early `liveEvent = filter.LiveEvent` assignment. Keep existing base matching and condition extraction. In matching branch:

```go
if match {
	if len(exclude) > 0 && !checkConditions(search, exclude, "exclude") {
		continue
	}
	if len(include) > 0 && !checkConditions(search, include, "include") {
		continue
	}
	return true, filter.LiveEvent
}
```

At function exhaustion, return `false, false`. Do not modify `checkConditions`; `strings.SplitSeq` and current empty-element behavior stay unchanged.

- [ ] **Step 4: Run evaluation and separator tests; verify GREEN**

```sh
go test -count=1 -mod=vendor ./src -run '^(TestFilterThisStreamUsesFirstAcceptingFilter|TestCheckConditionsSeparatorCompatibility)$' -v
```

Expected: PASS.

- [ ] **Step 5: Commit task only when explicitly requested**

```sh
git add src/m3u.go src/filter_test.go
git commit -m "fix: allow later stream filters to accept"
```

---

### Task 3: Prove repeated database classification is stable

**Files:**
- Modify: `src/filter_test.go`
- Exercise unchanged integration path: `src/data.go:822-1019`

**Interfaces:**
- Consumes: Task 1 deterministic `Data.Filter` and Task 2 non-vetoing `filterThisStream`.
- Produces: repeated active/inactive classification regression coverage through `buildDatabaseDVR()`.

- [ ] **Step 1: Expand test imports**

Add `os` and `path/filepath` to `src/filter_test.go`.

- [ ] **Step 2: Write repeated-build integration test**

```go
func TestBuildDatabaseDVRProducesStableFilterSets(t *testing.T) {
	restorePersistentState(t)

	root := t.TempDir()
	System = SystemStruct{}
	System.Flag.Info = true
	System.Folder.Data = root + string(os.PathSeparator)
	System.File.Settings = filepath.Join(root, "settings.json")
	System.UnfilteredChannelLimit = 100
	System.PlexChannelLimit = 100

	Settings = SettingsStruct{}
	Settings.Files.M3U = map[string]interface{}{
		"provider": map[string]interface{}{"name": "Filter Fixture"},
	}
	Settings.Filter = map[int64]interface{}{
		10: FilterStruct{Type: "custom-filter", Filter: "{ota-}"},
		20: FilterStruct{Type: "group-title", Filter: "ErsatzTV"},
	}

	playlist := `#EXTM3U
#EXTINF:-1 tvg-id="ota-16" tvg-name="OTA 16" group-title="OTA",OTA 16
https://example.test/ota
#EXTINF:-1 tvg-id="ersatz-1" tvg-name="Ersatz News" group-title="ErsatzTV",Ersatz News
https://example.test/ersatz
`
	if err := os.WriteFile(filepath.Join(root, "provider.m3u"), []byte(playlist), 0600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	wantActive := []string{"Ersatz News [ErsatzTV]", "OTA 16 [OTA]"}
	for build := 0; build < 32; build++ {
		if err := buildDatabaseDVR(); err != nil {
			t.Fatalf("buildDatabaseDVR() build %d error = %v", build, err)
		}
		if !reflect.DeepEqual(Data.StreamPreviewUI.Active, wantActive) {
			t.Fatalf("buildDatabaseDVR() build %d active = %#v, want %#v", build, Data.StreamPreviewUI.Active, wantActive)
		}
		if len(Data.StreamPreviewUI.Inactive) != 0 {
			t.Fatalf("buildDatabaseDVR() build %d inactive = %#v, want empty", build, Data.StreamPreviewUI.Inactive)
		}
	}
}
```

- [ ] **Step 3: Run integration test**

```sh
go test -count=1 -mod=vendor ./src -run '^TestBuildDatabaseDVRProducesStableFilterSets$' -v
```

Expected after Tasks 1-2: PASS on all 32 builds. If fixture assumptions differ from current provider loader, adjust only fixture initialization to match existing test helpers; do not weaken active/inactive assertions.

- [ ] **Step 4: Commit task only when explicitly requested**

```sh
git add src/filter_test.go
git commit -m "test: prove stable stream filter classification"
```

---

### Task 4: Verify filter workstream

**Files:**
- Verify: `src/data.go`
- Verify: `src/m3u.go`
- Verify: `src/filter_test.go`

**Interfaces:**
- Consumes: complete filter workstream.
- Produces: formatted, tested, vetted, cross-built change with XEPG and condition parser untouched.

- [ ] **Step 1: Format changed files**

```sh
gofmt -w src/data.go src/m3u.go src/filter_test.go
```

- [ ] **Step 2: Run focused tests**

```sh
go test -count=1 -mod=vendor ./src -run '^(TestCreateFilterRulesSortsIDsAndResetsPerRuleState|TestFilterThisStreamUsesFirstAcceptingFilter|TestBuildDatabaseDVRProducesStableFilterSets|TestCheckConditionsSeparatorCompatibility)$' -v
```

Expected: PASS.

- [ ] **Step 3: Run repository verification**

```sh
go test -count=1 -mod=vendor ./...
go test -mod=vendor -run='^$' -bench='^$' ./...
go vet -mod=vendor ./...
```

Expected: all commands exit 0.

- [ ] **Step 4: Run required Linux builds**

```sh
env GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags="-s -w" -o /tmp/Threadfin_linux_amd64 .
env GOOS=linux GOARCH=arm64 go build -mod=vendor -ldflags="-s -w" -o /tmp/Threadfin_linux_arm64 .
```

Expected: both builds exit 0.

- [ ] **Step 5: Check scope and diff**

```sh
git diff --check -- src/data.go src/m3u.go src/filter_test.go
git diff -- src/data.go src/m3u.go src/filter_test.go
git status --short
```

Confirm `src/xepg.go` is unchanged, `checkConditions` still uses `strings.SplitSeq`, and no schema/vendor/generated/UI files changed.

- [ ] **Step 6: Commit workstream only when explicitly requested**

```sh
git add src/data.go src/m3u.go src/filter_test.go
git commit -m "fix: make stream filters deterministic and non-vetoing"
```
