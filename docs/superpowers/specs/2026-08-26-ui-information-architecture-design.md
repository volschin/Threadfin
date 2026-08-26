# Evolutionary UI and Information Architecture Modernization

**Status:** Draft for review

**Date:** 2026-08-26

**Scope:** Information architecture, interaction model, visual direction, and staged migration of the existing Threadfin web interface

**Out of scope:** Backend feature expansion, a full SPA rewrite, changes to stream processing, and changes to the persisted configuration format

## Executive summary

Threadfin should keep the concepts inherited from xTeVe—Playlist, XMLTV, Filter, Mapping, Users, Settings, and Log—but stop presenting them as eight unrelated destinations. The interface should explain their operational sequence:

```text
Sources                  Lineup                    Delivery
Playlist + XMLTV  ->  Filter + Mapping  ->  HDHR / M3U / XMLTV clients
      |                     |                         |
      +---------------------+-------------------------+
                        Operations
                 activity, warnings, logs
```

The recommended design is an evolutionary, pipeline-oriented application shell. Existing users retain familiar labels and backend behavior, while a new Overview page, grouped navigation, visible prerequisites, persistent save state, and task-oriented empty/error states make the workflow understandable to new users.

The first release slice should add the shell and Overview using data that Threadfin already returns. It should not rewrite Mapping first, introduce a new frontend framework, or change API and persistence contracts.

## Research basis

### Authoritative product model

The xTeVe documentation defines a stable domain model that Threadfin inherited:

- the initial configuration chooses tuner capacity and either PMS-managed EPG or XEPG;
- Playlist imports M3U or HDHomeRun sources;
- Filter reduces imported streams;
- XMLTV supplies guide data in XEPG mode;
- Mapping is where channels are activated, numbered, renamed, grouped, assigned to EPG channels, and exported;
- Users controls access to the web UI and generated endpoints;
- Settings controls ingestion, streaming, backup, and authentication;
- Log exposes warnings and errors.

The same documentation reveals important dependencies that the current navigation does not communicate. Automatic mapping applies only to new channels and only when `tvg-id` matches an XMLTV channel ID. A channel cannot be active without an XMLTV assignment, although the dummy guide can satisfy that requirement. Editing a channel with **Done** only stages the change; a separate **Save** generates the lineup and output files in the background.

Primary source: [xTeVe setup and configuration](https://github.com/xteve-project/xTeVe-Documentation/blob/master/en/configuration.md).

Threadfin's own README still directs users to the xTeVe documentation and describes Threadfin as xTeVe-based. Its distinguishing UI-related capabilities are largely concentrated in Mapping: separate active/inactive tables, shift-based bulk selection, backup streams, improved sorting, and bulk renumbering. Those capabilities should remain visible rather than be flattened into a generic CRUD interface.

Primary source: [Threadfin README](https://github.com/Threadfin/Threadfin).

### Recurring user problems

Forum and issue reports are anecdotal and span different xTeVe, Threadfin, Plex, and Jellyfin versions. They are not proof of current backend defects. They are strong evidence of recurring mental-model and feedback failures:

1. **The required sequence is invisible.** Users import a Playlist and XMLTV successfully, then see an empty Mapping page because XEPG, filters, load order, or activation prerequisites are not apparent. A Threadfin user with eight valid XMLTV channels saw no mapping entries until another user pointed to the EPG-source setting. See [Help With Threadfin](https://www.reddit.com/r/unRAID/comments/1gccm2o/help_with_threadfin/).

2. **Filter is misunderstood as optional search or exclusion.** Users expect imported streams to appear automatically, do not understand that filters determine which streams reach Mapping, and often leave Threadfin for an external playlist editor. See [XTeve setup](https://www.reddit.com/r/PleX/comments/myccv4/) and [invalid filter rule discussion](https://www.reddit.com/r/PleX/comments/16jjwta/xteve_plex_invalid_filter_rule/).

3. **Mapping requirements are hidden.** Red rows communicate failure but not the reason or next action. Users ask why XMLTV ID is required, why channels remain disabled, or why a dummy EPG must be selected. See [xTeVe channels not showing up in Plex](https://www.reddit.com/r/PleX/comments/bkkrtw/xteve_channels_not_showing_up_in_plex/) and [channel mapping is not automatic](https://github.com/xteve-project/xTeVe/issues/73).

4. **Large-list work is too manual.** Users report mapping hundreds or thousands of channels one at a time, difficulty sorting by group, and poor performance with very large playlists. The open xTeVe request to keep channel numbers organized by filter/group is a concrete expression of this need. See [xTeVe issue 148](https://github.com/xteve-project/xTeVe/issues/148), [Plex and xTeVe](https://www.reddit.com/r/PleX/comments/1dvu5i5/plex_and_xteve/), and [Threadfin with M3U4U](https://www.reddit.com/r/PleX/comments/1i8w3ar/help_with_iptv_via_threadfin_and_m3u4u/).

5. **Save state is unsafe and ambiguous.** The documented two-stage Mapping interaction—Done, then Save—is easy to miss. Users explicitly note that xTeVe does not warn when leaving with unsaved mappings. See [Plex keeps unmapping my channels](https://www.reddit.com/r/PleX/comments/kg4cbo/plex_keeps_unmapping_my_channels/) and the setup walkthrough that warns users to save before leaving [Plex Live TV with OpenVix](https://www.reddit.com/r/PleX/comments/12dinjp/plex_live_tv_dr_with_openvix/).

6. **Fault ownership is unclear.** Threadfin may have valid output while Plex still shows no guide or stale channels. Users repeatedly delete and recreate tuners because the UI does not separate source health, Threadfin processing, generated-output readiness, and downstream-client state. See [xTeVe IPTV for Plex DVR](https://forums.plex.tv/t/xteve-iptv-for-plex-dvr/278500?page=72) and [Plex LiveTV with IPTV summary](https://forums.plex.tv/t/plex-livetv-with-iptv-sumup/570035).

### Useful patterns in newer tools

Dispatcharr is not a feature target for Threadfin. Its documentation nevertheless validates several interaction patterns relevant to the same domain:

- a Getting Started surface that links directly to the next required action;
- M3U and EPG grouped as source management;
- a channel workspace with filterable columns, explicit hidden/active state, bulk actions, dummy EPG, auto-match actions, stream source details, and preview;
- generated client links presented as a first-class destination rather than buried in an information dialog.

Sources: [Dispatcharr Getting Started](https://dispatcharr.github.io/Dispatcharr-Docs/getting-started/), [M3U and EPG Manager](https://dispatcharr.github.io/Dispatcharr-Docs/m3u-epg-manager/), and [Channels](https://dispatcharr.github.io/Dispatcharr-Docs/channels/).

The lesson is not to copy Dispatcharr's larger feature set. It is to organize Threadfin around the same real-world nouns and progression while preserving Threadfin's smaller scope.

## Current Threadfin constraints

The proposal accounts for the current implementation rather than assuming a greenfield frontend:

- `html/index.html` provides one shell whose page content is created dynamically.
- `ts/base_ts.ts` defines the eight current menu entries and five settings groups.
- `ts/menu_ts.ts` combines navigation, tables, dialogs, Mapping, bulk actions, and save behavior in one large file.
- `ts/network_ts.ts` maintains the WebSocket connection and receives the complete UI response model.
- the response already includes version, EPG mode, source and client connection counts, output URLs, error/warning counts, settings, files, mappings, logs, users, and active/inactive stream data.
- the web assets are compiled with the checked-in TypeScript configuration and embedded into `src/webUI.go` through the documented development-mode build.

This makes a new Overview and shell possible before adding endpoints. It also argues against starting with a framework migration: the highest-value first slice can reuse existing data and commands.

## Alternatives considered

### A. Reskin the existing navigation

Keep the top navigation and page structure; replace colors, typography, spacing, icons, and table styles.

**Advantages:** Smallest code change and lowest immediate regression risk.

**Disadvantages:** The hidden process, prerequisites, save model, and fault boundaries remain hidden. New users still need external guides.

**Decision:** Rejected as insufficient.

### B. Pipeline-oriented shell with legacy vocabulary

Add Overview, group the existing destinations by purpose, and migrate screens one at a time. Preserve existing terms as visible labels and preserve backend commands and stored data.

**Advantages:** Addresses the research findings without a flag-day rewrite; familiar to existing users; provides independently releasable slices.

**Disadvantages:** Old and new screen code coexist temporarily; visual consistency improves over several releases rather than at once.

**Decision:** Recommended.

### C. Full task wizard or new SPA

Replace the application with a new guided workflow and frontend architecture in one program of work.

**Advantages:** Maximum freedom to normalize state and interactions.

**Disadvantages:** Highest compatibility and regression risk; expert users lose spatial memory; API gaps become redesign blockers; substantially larger test and migration surface.

**Decision:** Rejected for the current goal.

## Proposed information architecture

Desktop uses a persistent left sidebar. Smaller screens use the same hierarchy in a drawer; no separate mobile information architecture is introduced.

```text
Threadfin

Overview

SOURCES
  Playlist
  XMLTV

LINEUP
  Filter
  Mapping

DELIVERY
  Connections
  Users              only when web authentication exposes user management

SYSTEM
  Activity
  Settings
  Log

Account
  Logout             only when authenticated
```

Compatibility rules:

- **Playlist, XMLTV, Filter, Mapping, Users, Settings, and Log remain literal labels.** Documentation and experienced-user vocabulary continue to match the UI.
- The grouping labels are explanatory, not replacements for existing destinations.
- **Connections** promotes today's Server Information modal and output URLs into a page. The old Server Information affordance can link to it for at least one transition release.
- **Activity** contains active/inactive streams and connection utilization currently spread across the navigation header and bottom tables. It does not add historical analytics.
- PMS mode does not silently remove XMLTV and Mapping. They remain visible but explain that XEPG is required and link to the relevant setting. This avoids the “missing menu” problem and makes mode consequences discoverable.
- Existing URLs and backend menu keys remain stable during migration. Navigation can select a new presentation without changing persisted configuration.

## Overview page

Overview is an operational map, not a grid of decorative metrics. Its signature element is a horizontal **signal path** showing where data currently stops:

```text
[Playlist: ready] -> [Filter: 126 selected] -> [XMLTV: ready]
                                      -> [Mapping: 118 active / 8 need attention]
                                      -> [Outputs: ready]
```

Each stage has one status, a short explanation, and one context-sensitive action. Examples:

- no playlist: **Add playlist**;
- playlist loaded but no selected channels: **Create filter**;
- PMS mode: XMLTV stage reads **Managed by client** rather than warning;
- XEPG with no guide: **Add XMLTV**;
- mappings missing EPG assignment: **Review 8 channels**;
- outputs ready: **Connect Plex, Jellyfin, Emby, or another client**.

Below the signal path:

1. **Current activity** shows active streams and connection capacity already available in the response.
2. **Attention required** lists current warnings/errors and links to filtered Log or the affected stage.
3. **Output endpoints** shows DVR, M3U, and XMLTV URLs with explicit Copy actions and mode-dependent availability.
4. **Recent source state** shows last update and availability for each Playlist/XMLTV source.

The page must not infer that Plex, Jellyfin, or Emby successfully consumed an output unless Threadfin can actually observe that. It may say “Threadfin output ready” and show recent client connections, but not “Plex configured.”

## Screen-level interaction design

### Playlist and XMLTV

Both pages share a source-list pattern while keeping their existing labels:

- page title, one-sentence purpose, and primary **Add playlist** or **Add XMLTV** action;
- compact rows/cards showing name, source type, last update, availability, and counts;
- row actions named **Edit**, **Update now**, and **Delete**;
- update-in-progress, success, and error feedback remains attached to the affected source;
- empty states link to the next workflow step instead of showing an empty table;
- URL and local-path examples appear beside the input rather than only in external documentation.

M3U and XMLTV remain separate pages because they are separate, familiar concepts and have different data. Grouping them in navigation supplies the missing relationship without forcing a combined management screen.

### Filter

The Filter page must explain its effect before its syntax:

> Filters select which imported streams become channels available for Mapping.

The page shows imported, selected, and excluded counts. Group Filter and Custom Filter remain the two existing types. The form presents Group Filter first and marks Custom Filter as advanced. Existing include/exclude behavior is described with live plain-language summaries such as “Include names containing BBC or ITV; exclude names containing +1.”

Where current response data permits it, show the current result count before save. Do not add a new backend preview engine in the first slice. Invalid rules keep the form data, identify the field, and provide a valid example.

### Mapping

Mapping is the primary expert workspace and receives the most careful evolution.

The top of the page contains:

- segmented views: **Needs attention**, **Active**, and **Inactive**;
- search plus filters for Playlist, group, XMLTV source, activation state, and explicit mapping-attention reason;
- result and selection counts;
- a sticky bulk-action bar when rows are selected;
- a persistent dirty-state bar when changes are staged.

The default view is **Needs attention** when unresolved channels exist; otherwise **Active**. “Red versus green” is supplemented with explicit text such as **Missing EPG assignment**, **Invalid EPG assignment**, **Hidden from outputs**, or **Inactive**. Color is never the only state signal. Threadfin's current response does not contain per-channel reachability, so this screen must not label channels Online or Offline based on activation or filter selection.

Editing one channel opens a side panel on wide screens and a full-screen dialog on small screens. Existing fields and behavior remain: active state, number, name, description, logo, update-name/logo flags, category, group, XMLTV file/channel, backup streams, and probe. Advanced fields are collapsed but searchable.

Save semantics become explicit without changing the backend transaction:

- the editor action is **Apply to draft**, replacing ambiguous **Done**;
- the page-level action is **Save mapping**;
- the sticky bar reports the number of staged channels;
- navigating away with staged changes prompts **Save mapping**, **Discard draft**, or **Stay**;
- while outputs are generated, the UI shows **Saving mapping and rebuilding outputs**;
- success confirms **Mapping saved; outputs rebuilt** only when the server response proves the synchronous rebuild completed. If a running scan causes the rebuild to be queued, it instead confirms **Mapping saved; output rebuild requested**.

Bulk edit remains a first-class operation. It must support the fields Threadfin already supports and must not promise auto-matching confidence, server-side pagination, or undo history until backend contracts exist.

### Connections

Connections answers “What do I paste into my client?” It contains:

- HDHomeRun/DVR address;
- M3U URL when available;
- XMLTV URL and GZIP variant when available;
- authentication-specific variants only when relevant;
- Copy action and a short client-neutral explanation for each endpoint;
- client-specific expandable instructions for Plex, Jellyfin, and Emby, clearly separated from Threadfin health.

This page reuses current output values. It does not attempt automatic client configuration.

### Activity and Log

Activity presents the current active/inactive streams, source connection capacity, and client connection capacity. It is live operational state, not a long-term dashboard.

Log remains the canonical diagnostic record. It gains severity filters, search, and links from Overview warnings. **Reset logs** remains visibly destructive and separate from filtering. Errors should say which layer failed where possible: source download, parsing/filtering, mapping/output generation, streaming, or client connection.

### Settings and Users

Settings keeps the existing five categories but presents them as sections in a secondary navigation:

- General
- Files
- Streaming
- Backup
- Authentication

EPG Source is promoted to the top of General with an impact explanation. Changing PMS/XEPG must preview which pages and outputs become applicable before save. Backup and restore remain in the Backup section rather than competing with the global Save action.

Users keeps existing permission names—WEB, PMS, M3U, XML, API—because they are protocol contracts. Each receives a plain-language explanation. Passwords are never displayed as ordinary table content.

## Visual direction

Threadfin should look like a compact broadcast operations console, not a generic SaaS administration template.

### Palette

- **Signal black** `#101417`: application background
- **Console graphite** `#182027`: navigation and raised surfaces
- **Panel steel** `#24313A`: interactive surfaces and table headers
- **Threadfin cyan** `#27C2D1`: primary action, focus, active signal path
- **On-air green** `#43C58A`: healthy/active state
- **Warning amber** `#E4A83A` and **fault red** `#E26060`: attention states

### Type and density

- use a highly legible system sans-serif stack for controls and prose;
- use a tabular utility/monospace stack for channel numbers, stream counts, URLs, codecs, and logs;
- remove global letter spacing and uppercase headings;
- default to compact desktop density, with comfortable touch targets and an optional future density control only if real demand appears.

### Signature

The signal-path status line is the memorable Threadfin-specific element. It encodes the product's actual job and replaces the common but less useful collection of independent statistic cards.

Motion is limited to stage transitions, update progress, and side-panel entry. Reduced-motion preferences are honored.

## Accessibility and responsive behavior

- restore the viewport meta declaration on the main application page;
- use semantic buttons and links instead of clickable list items and table cells;
- every action is keyboard reachable and has a visible focus state;
- rows remain navigable without relying on double-click, right-click, or color;
- table headers expose sorting state;
- labels are programmatically associated with controls;
- dialogs trap focus and return focus to their invoker;
- live update and save results use restrained status announcements;
- on narrow screens, data tables show the identity and status columns first, with remaining fields in row details;
- Mapping is optimized for desktop but remains operable on a tablet; no claim is made that bulk administration is pleasant on a phone.

## Evolutionary implementation strategy

### Slice 1: foundation and Overview

- add design tokens and scoped shell styles;
- restore responsive viewport and semantic navigation;
- add grouped navigation while retaining all current labels and menu keys;
- build Overview and Connections from the existing response model;
- retain Server Information as a compatibility link;
- add browser-level smoke coverage for navigation, keyboard focus, responsive shell, and conditional PMS/XEPG states.

No API changes and no framework migration.

### Slice 2: source screens and setup guidance

- migrate Playlist and XMLTV lists/forms to the shared source pattern;
- improve progress, empty, and error states;
- revise the initial wizard to explain the workflow and finish on Overview;
- leave old Filter and Mapping behavior intact.

### Slice 3: Filter

- migrate Filter list/forms;
- add plain-language rule summaries and result counts available from existing data;
- validate parity for group and custom filters before removing the legacy renderer.

### Slice 4: Mapping

- first wrap existing Mapping data and commands in the new page shell;
- then introduce attention views, explicit staged-save state, responsive editor, and bulk toolbar in independently testable steps;
- preserve active/inactive semantics, selection behavior, backup streams, probe, sorting, and output rebuild behavior;
- do not remove the legacy Mapping path until parity tests and large-list tests pass.

### Slice 5: operations and administration

- migrate Activity, Log, Settings, Users, login, first-user creation, and configuration wizard;
- remove compatibility links only after at least one transition release and explicit review.

## Technical boundaries

- Keep the Go backend, WebSocket transport, request command names, response schema, persisted JSON, and generated output contracts stable unless a later screen demonstrates a concrete missing capability.
- Use the checked-in TypeScript pipeline for the first slices. Split new code by screen and shared component responsibility instead of extending `menu_ts.ts`.
- New styles must be scoped to the new shell/screens so old screens remain functional during coexistence.
- Keep Bootstrap only where it provides current behavior during migration. Do not make replacing Bootstrap a goal.
- Avoid a new state-management library. The first slices need a small typed client state derived from the existing response and explicit draft state for forms.
- Continue compiling TypeScript into `html/js/` and regenerating `src/webUI.go` in the same change whenever browser assets change.

## Verification and acceptance criteria

### Research-derived usability scenarios

1. A new user with one M3U and one XMLTV URL can identify the next required action at every step without opening external documentation.
2. A PMS-mode user understands why XMLTV/Mapping are inactive and how client-managed guide data differs from XEPG.
3. An XEPG user can distinguish imported streams, filtered channels, mapped active channels, and generated outputs.
4. A user with unmapped channels can filter directly to the reason each channel needs attention and apply a dummy guide in bulk using existing capabilities.
5. A user cannot leave Mapping with staged changes without an explicit choice.
6. A user can copy the correct DVR, M3U, and XMLTV endpoints from Connections and sees which are unavailable in the current mode.
7. A Threadfin source or processing error points to Threadfin diagnostics; the UI does not falsely claim downstream Plex/Jellyfin/Emby success.
8. Existing users can still find every legacy destination by its current name.

### Regression gates

- old and new presentations produce identical request commands and persisted values for migrated actions;
- Playlist, XMLTV, Filter, Mapping, Users, Settings, Log, authentication, wizard, backup/restore, stream probe, and copy actions retain behavior;
- Mapping is exercised with empty, small, and near-limit lineups plus a characterized very-large-list case;
- keyboard-only and tablet-width workflows pass for every migrated screen;
- `tsc -p ./ts/tsconfig.json` succeeds and `src/webUI.go` is regenerated;
- the repository's Go test and vet gates pass before commit.

## Non-goals

- supporting new source formats or clients;
- changing the 480-channel limits;
- automatic EPG matching beyond existing behavior;
- configuring Plex, Jellyfin, or Emby remotely;
- historical monitoring, analytics, or notification services;
- theming, plugin architecture, or arbitrary dashboard customization;
- replacing all frontend dependencies in the same modernization effort.

## Open review decisions

The implementation plan should not be written until these design decisions are approved:

1. Accept the grouped sidebar hierarchy while retaining the seven legacy destination names.
2. Accept **Connections** and **Activity** as the only new primary destinations.
3. Accept the signal-path Overview as the first implementation slice.
4. Accept explicit Mapping draft/save semantics without changing the backend transaction.
5. Accept no new frontend framework for the foundation slice; reassess only after the shell and source screens expose a concrete limitation.
