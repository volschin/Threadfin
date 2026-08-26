# Evolutionary UI Modernization Implementation Plan

> **For Codex:** Execute this plan with subagent-driven development. Use one implementer and a fresh reviewer per task, record reports under `.superpowers/sdd/2026-08-26-ui-modernization/`, and do not advance past a failed review.

**Goal:** Deliver the approved pipeline-oriented Threadfin interface, preserve every legacy workflow and persisted contract, and add a narrowly authorized machine-readable configuration API suitable for a small LLM agent.

**Architecture:** Keep the current Go server, global WebSocket response, Bootstrap compatibility layer, and TypeScript compiler. Add a semantic application shell and small screen-focused TypeScript modules around the legacy renderers, then replace one screen at a time. New destinations use stable string keys; legacy backend `openMenu` numeric destinations are resolved through an explicit adapter so `System.WEB.Menu` order remains unchanged. The configuration API is a separate authenticated HTTP command surface with describe/get/plan/apply and an explicit, default-off permission.

**Tech stack:** Go 1.27 with vendored dependencies, TypeScript ES6 compiled by `tsc`, Bootstrap 5 compatibility, embedded static assets in `src/webUI.go`, Go tests, native headless Chrome/CDP via browser-use.

## Invariants

- Keep literal destinations Playlist, XMLTV, Filter, Mapping, Users, Settings, and Log.
- Do not reorder or prepend `menuItems` or `System.WEB.Menu`; backend responses still address legacy menu indexes.
- Do not change existing WebSocket commands or persisted values for migrated browser actions.
- New browser assets stay scoped below `.tf-app`; legacy pages remain usable throughout the migration.
- PMS mode explains XMLTV/Mapping applicability instead of hiding the destinations.
- Mapping draft state is client-side only; the existing `saveEpgMapping` transaction remains authoritative.
- The agent API is unavailable unless API authentication and the new config permission are both enabled. It never returns secrets or accepts arbitrary paths, users, authentication values, runtime/network settings, backup/restore, updates, or raw configuration maps.
- Every asset task includes TypeScript compilation and deliberate `src/webUI.go` regeneration after edits have settled.
- Browser fixtures bind to loopback, disable auto-update and SSDP, and use only temporary local M3U/XMLTV inputs.

## Task 1: Establish structural tests and the semantic application shell

**Files:**

- Modify: `src/html_build_test.go`
- Modify: `html/index.html`
- Create: `html/css/app-shell.css`
- Modify: `html/lang/en.json`

**Steps:**

1. Add failing structural tests asserting an enabled viewport meta element, `.tf-app`, semantic navigation and main landmarks, the new stylesheet, and dedicated Overview/Connections/Activity containers.
2. Run `go test -count=1 -mod=vendor ./src -run 'TestIndexHTML'` and retain the expected failure in the task report.
3. Reshape `index.html` around this stable compatibility skeleton while retaining all IDs consumed by legacy code:

   ```html
   <div class="tf-app">
     <aside class="tf-sidebar">
       <nav id="main-menu" aria-label="Primary"></nav>
     </aside>
     <main id="main-content" tabindex="-1">
       <section id="content"></section>
     </main>
   </div>
   ```

   Keep the server information modal and active/inactive stream table IDs until their callers are migrated.
4. Add scoped design tokens and responsive sidebar/drawer rules using the approved palette. Add `:focus-visible`, reduced-motion handling, 44px mobile targets, and no global element resets.
5. Add English labels and concise explanatory copy for the five navigation groups and three new screens.
6. Rerun the focused test and check `git diff --check`.

## Task 2: Add stable grouped navigation without breaking numeric legacy routing

**Files:**

- Create: `ts/navigation_ts.ts`
- Modify: `ts/base_ts.ts`
- Modify: `ts/menu_ts.ts`
- Modify: `ts/network_ts.ts`
- Modify: `html/index.html`
- Create: `src/ui_navigation_test.go`

**Steps:**

1. Add source-level contract tests that prove the seven legacy menu keys remain in their original order and that numeric backend destinations resolve only through `openLegacyMenu(index)`.
2. Define string-keyed navigation metadata independently from `menuItems`:

   ```ts
   type AppDestination = "overview" | "playlist" | "xmltv" | "filter" |
     "mapping" | "connections" | "users" | "activity" | "settings" |
     "log" | "logout";
   interface NavigationGroup { key: string; label: string; items: AppDestination[]; }
   ```

3. Render semantic buttons/links grouped as Overview, Sources, Lineup, Delivery, System, and Account. Preserve numeric DOM IDs on hidden or explicit legacy adapter targets, not on newly ordered items.
4. Route new destinations by string key and legacy destinations through the existing `ShowContent` entry point. Replace `document.getElementById(response.openMenu).click()` with `openLegacyMenu(response.openMenu)`.
5. Manage current-page state, `aria-current`, focus return, mobile navigation dismissal, and browser history without inventing public server routes.
6. Verify compile, focused Go tests, keyboard navigation, and direct numeric save-response routing.

## Task 3: Derive setup state and implement Overview

**Files:**

- Create: `ts/app_state_ts.ts`
- Create: `ts/overview_ts.ts`
- Modify: `ts/navigation_ts.ts`
- Modify: `ts/network_ts.ts`
- Modify: `html/index.html`
- Modify: `html/css/app-shell.css`
- Create: `src/ui_overview_test.go`

**Steps:**

1. Define pure, null-safe selectors over the existing `SERVER` response for playlist count, selected stream count, XMLTV applicability/readiness, active/unresolved mapping counts, output readiness, active streams, capacity, warning/error counts, and source state.
2. Cover empty, PMS, XEPG-ready, and XEPG-attention fixtures in tests. Do not infer downstream client configuration.
3. Render the signature signal path as an ordered list with explicit text status, explanation, and one context action per stage. Example state:

   ```ts
   { key: "mapping", status: "attention", summary: "8 need attention",
     action: { label: "Review 8 channels", destination: "mapping" } }
   ```

4. Render Current activity, Attention required, Output endpoints, and Recent source state below the path. Copy actions use the existing clipboard helper and announce success through a polite status region.
5. Make Overview the initial destination without changing a server menu index. Refresh only its live portions after `updateLog` responses.
6. Verify desktop/tablet/phone layouts and empty/PMS/XEPG content using real WebSocket fixture responses.

## Task 4: Implement Connections and Activity

**Files:**

- Create: `ts/connections_ts.ts`
- Create: `ts/activity_ts.ts`
- Modify: `ts/navigation_ts.ts`
- Modify: `ts/network_ts.ts`
- Modify: `html/css/app-shell.css`
- Create: `src/ui_operations_test.go`

**Steps:**

1. Render DVR, M3U, and XMLTV endpoint cards from existing response values, marking unavailable endpoints with a reason instead of an empty control. Do not invent the XMLTV-GZIP variant described as conditional in the design: the current response has no such field, so it remains absent until a verified backend contract exists.
2. Add client-neutral help and collapsed Plex/Jellyfin/Emby notes that never claim the client was configured successfully.
3. Move the current active/inactive stream presentation and source/client capacity summaries into Activity while keeping compatibility DOM IDs until legacy callers are removed.
4. Refresh Activity after live responses without replacing focused controls.
5. Retain the server-info affordance as a link to Connections for this transition release.
6. Test copy, conditional availability, capacity labels, responsive overflow, and current-state refresh.

## Task 5: Modernize Playlist, XMLTV, and reachable setup guidance

**Files:**

- Create: `ts/sources_ts.ts`
- Modify: `ts/menu_ts.ts`
- Modify: `ts/configuration_ts.ts`
- Modify: `html/configuration.html`
- Modify: `html/css/app-shell.css`
- Modify: `html/lang/en.json`
- Create: `src/ui_sources_test.go`

**Steps:**

1. Characterize existing source add/edit/update/delete commands and payloads with source-level tests.
2. Introduce the shared source-list renderer with purpose text, source type, availability, last update, counts, and explicit Edit/Update now/Delete actions.
3. Add actionable empty, progress, success, and field-level error states without altering command payloads.
4. Explain URL/local-path inputs in the forms and point successful source creation to the next workflow stage.
5. Update the configuration wizard copy and completion route to Overview. Add a regression test documenting and fixing the normal-startup reachability of the intended first-run experience if the server state model confirms it is unreachable.
6. Verify command/persistence parity through a disposable runtime configuration.

## Task 6: Modernize Filter while preserving rule semantics

**Files:**

- Create: `ts/filter_ts.ts`
- Modify: `ts/menu_ts.ts`
- Modify: `html/css/app-shell.css`
- Modify: `html/lang/en.json`
- Create: `src/ui_filter_test.go`

**Steps:**

1. Characterize Group Filter and Custom Filter payloads, include/exclude behavior, invalid-input behavior, and result counts available in the response.
2. Render imported, selected, and excluded counts plus the approved explanatory sentence.
3. Present Group Filter first and Custom Filter as advanced. Generate a plain-language summary from the entered rules without a server preview claim.
4. Preserve form values on validation failure, identify the invalid field, and show one syntactically valid example.
5. Verify identical `saveFilter` payloads and persisted filter structures for representative create/edit/delete flows.

## Task 7: Modernize Mapping and make staged changes explicit

**Files:**

- Create: `ts/mapping_state_ts.ts`
- Create: `ts/mapping_ts.ts`
- Modify: `ts/menu_ts.ts`
- Modify: `ts/network_ts.ts`
- Modify: `html/css/app-shell.css`
- Modify: `html/lang/en.json`
- Create: `src/ui_mapping_test.go`

**Steps:**

1. Characterize the current active/inactive data, multi-select, bulk edit, numbering, backup streams, probe, sorting, dummy guide, undo, and `saveEpgMapping` behavior before rendering changes.
2. Derive explicit Needs attention reasons from response-backed state: Missing EPG assignment, Invalid EPG assignment, Hidden from outputs, and Inactive. Default to Needs attention when nonempty, otherwise Active. Do not mislabel activation or filter selection as per-channel Online/Offline state.
3. Add search and available playlist/group/XMLTV/activation/attention filters, result/selection counts, and a sticky bulk toolbar using existing operations.
4. Rename editor completion to Apply to draft. Track changed channel IDs locally and expose a persistent Save mapping bar.
5. Guard navigation and unload while dirty with Save mapping, Discard draft, and Stay. Do not clear draft state until the server confirms `saveEpgMapping`; distinguish a proven synchronous output rebuild from a rebuild queued behind an active scan.
6. Preserve every existing field and operation in a responsive side panel/dialog, with advanced fields collapsed but searchable.
7. Test empty, small, near-480, and characterized very-large datasets; keyboard-only selection/edit/save; tablet operation; full payload/persistence parity.

## Task 8: Modernize Settings, Users, Log, and authentication pages

**Files:**

- Create: `ts/settings_page_ts.ts`
- Create: `ts/users_page_ts.ts`
- Create: `ts/log_page_ts.ts`
- Modify: `ts/settings_ts.ts`
- Modify: `ts/logs_ts.ts`
- Modify: `html/login.html`
- Modify: `html/create-first-user.html`
- Modify: `html/css/app-shell.css`
- Modify: `html/lang/en.json`
- Create: `src/ui_admin_test.go`

**Steps:**

1. Render Settings with secondary navigation General, Files, Streaming, Backup, Authentication. Promote EPG Source with a PMS/XEPG impact preview while preserving `saveSettings` values.
2. Keep backup/restore separate from global save and visibly destructive where applicable.
3. Explain WEB/PMS/M3U/XML/API permissions in Users and ensure passwords are never rendered as table values.
4. Add Log severity filters, search, layer labels where the current message metadata supports them, and keep Reset logs separate and destructive.
5. Apply the shared visual/accessibility foundation to login and first-user creation without changing authentication contracts.
6. Verify all admin commands and persisted structures are identical to the legacy flows.

## Task 9: Add a safe small-LLM configuration command API

**Files:**

- Modify: `src/struct-user.go`
- Modify: `src/api.go`
- Create: `src/api_config.go`
- Create: `src/api_config_test.go`
- Modify: `html/lang/en.json`
- Modify: `ts/users_page_ts.ts`

**Steps:**

1. Add a `CONFIG`/`config` user permission defaulting to false. Existing users and API clients receive no new authority.
2. Add authenticated API commands `config.describe`, `config.get`, `config.plan`, and `config.apply`. Reject them unless the API is enabled, API authentication succeeds, the principal has config permission, and setup is complete.
3. Return a versioned, machine-readable schema and sanitized current state. Redact URL userinfo/query secrets and never expose passwords, tokens, authentication hashes, backup data, or arbitrary settings maps.
4. Accept strict JSON with exactly one allowlisted action per request. Initial actions cover safe onboarding fields, M3U/HDHR/XMLTV source CRUD/refresh, Filter CRUD, and targeted Mapping patches/bulk channel IDs. Explicitly exclude users/authentication, paths, binaries/options, bind/port/TLS, backup/restore, application updates, uploads, and arbitrary raw maps.
5. `config.plan` performs no disk/network mutation and returns normalized action, consequences, warnings, `config_revision`, and deterministic `plan_digest`. `config.apply` requires both values, revalidates the exact action and revision, and calls the same domain operations used by the browser.
6. Bound request size, action count, strings, arrays, and channel IDs. Use generic authorization errors and never log source URLs or request bodies.
7. Add denial, redaction, schema, validation, no-mutation plan, stale-revision, digest-tamper, allowed-apply, and forbidden-field tests. Document the four-call agent loop in `config.describe`, not in a separate protocol or MCP dependency.

## Task 10: Build, embed, and run independent browser refinement loops

**Files:**

- Modify: browser assets as findings require
- Regenerate: `html/js/*.js`
- Regenerate: `src/webUI.go`
- Create: `docs/superpowers/reports/2026-08-26-ui-acceptance.md`

**Steps:**

1. Run `tsc -p ./ts/tsconfig.json` and fix every error.
2. Regenerate `src/webUI.go` once from a loopback-only, owner-only temp config with auto-update/files-update/SSDP disabled; confirm only intended embedded assets changed.
3. Start the embedded build without `-dev` against fresh small local fixtures for empty, PMS, XEPG-ready, and XEPG-attention states.
4. Run sequential independent browser-use reviews against reset fixtures:
   - domain expert familiar with xTeVe/Threadfin, checking parity and expert efficiency;
   - second domain expert focused on Mapping and operational diagnosis;
   - novice with only the goal “configure one playlist and guide and connect a client,” checking discoverability;
   - novice focused on keyboard/tablet use and terminology.
5. Give reviewers independent reports and no prior review conclusions. For each material finding, reproduce it, add a failing regression check when feasible, fix it, and rerun the affected reviewer scenario. Continue until two consecutive independent passes find no acceptance blocker or major regression.
6. Audit scenarios 1–8 and every regression gate in the approved spec, recording direct evidence and any deliberately characterized limitation without converting it into a pass.
7. Run final gates:

   ```sh
   tsc -p ./ts/tsconfig.json
   go test -count=1 -mod=vendor ./...
   go vet -mod=vendor ./...
   git diff --check
   ```

8. Review the exact diff, verify no fixture secrets or temp paths are tracked, commit atomically by task, push `feat/ui-modernization`, and report the exact remote SHA only after rereading it.
