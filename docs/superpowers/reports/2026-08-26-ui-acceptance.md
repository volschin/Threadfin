# Evolutionary UI modernization acceptance

Date: 2026-08-26

Scope: the approved information architecture, all migrated legacy workflows, the first-run experience, and the default-off typed configuration API on `feat/ui-modernization`.

## Verdict

PASS. No blocker or major finding remains. The final independent expert and novice passes were consecutive and found no reproducible blocker or major regression. All browser fixtures were temporary, owner-only, used local dummy M3U/XMLTV data, listened on loopback only, and disabled or runtime-suppressed update and SSDP traffic.

## Independent browser evidence

| Perspective | Coverage | Result | Recording |
| --- | --- | --- | --- |
| Threadfin/xTeVe domain expert | Full information architecture, PMS/XEPG truth, Mapping attention/draft guard, first run, responsive shell | Initial majors fixed; final regression pass and fresh first-run pass | Recordings `threadfin-domain-overall-20260826`, `threadfin-domain-regression-20260826`, `threadfin-first-run-final-20260826` |
| Operations/Mapping expert | Attention reasons, bulk dummy, save/rebuild, Connections, Activity/Log, tablet keyboard editor | PASS, no blocker/major | Recording `threadfin-task10-operations-mapping` |
| Setup novice | One playlist, one guide, next action, client endpoints, low-height wizard | PASS after progress/footer refinements | Recordings `threadfin-novice-first-run`, `threadfin-novice-wizard-regression-cfe2040d`, `threadfin-novice-next-visible-1e38b606` |
| Keyboard/tablet novice | All destinations, Mapping dialog and departure guard, Settings and Log at tablet and low height; shared source-modal naming and focus lifecycle | PASS; reported focus escape was withdrawn after an independent exact-key audit; final modal reuse had no blocker, major, or minor finding | Recordings `threadfin-novice-accessibility`, `threadfin-novice-mapping-focus-audit`, `threadfin-modal-a11y-8f18158d` |

The final fixture build used SHA-256 `8f18158def75ab57ef296228e0648dbcab45593b3ea0bb3711af85f215d71b1c`. All six fixture listeners were verified on `127.0.0.1` only. The ready fixture reported one selected stream, one active Mapping channel, one XMLTV channel/programme, and usable local DVR/M3U/XMLTV endpoints.

## Research-derived scenarios

1. **New user with one M3U and XMLTV: PASS.** A novice completed the fresh XEPG wizard from visible guidance only. Finish loaded the normal Overview document, which reported Playlist and XMLTV `1 / 1 ready`, one active Mapping channel, and ready outputs. Progress remains `1/4 -> 2/4 -> 3/4 -> 4/4`; the primary action remains visible at `780x437`.
2. **PMS mental model: PASS.** Overview, XMLTV, Mapping, Settings, and Connections state that guide management belongs to the connected client in PMS mode. Direct Mapping remains findable and offers an EPG Source Settings action instead of a misleading empty XEPG workspace.
3. **XEPG signal-path distinctions: PASS.** Overview separately reports imported/ready sources, selected streams, XMLTV readiness, active/unresolved Mapping, and output readiness. The UI does not infer downstream client success.
4. **Attention reasons and bulk dummy guide: PASS.** Browser evidence covered Missing EPG assignment and Invalid EPG assignment, visible bulk selection, Threadfin Dummy/PPV staging, and the dirty save bar. Generated fixtures also cover Hidden from outputs and Inactive.
5. **Cannot abandon staged Mapping silently: PASS.** Browser evidence covered the Save mapping / Discard draft / Stay guard. Stay retains the draft; safe Save followed the rebuild lifecycle into Connections. Generated tests cover Discard, pending/ambiguous saves, History, programmatic navigation, and focus return.
6. **Correct client endpoints and mode availability: PASS.** Ready XEPG exposed and copied DVR, M3U, and XMLTV values with confirmation. PMS visibly marks M3U/XMLTV guide outputs unavailable and explains why.
7. **Truthful error attribution: PASS.** Source, Filter, Mapping, and transport fixtures retain values/drafts and show Threadfin request or processing failures. Activity and Connections describe only current Threadfin state and explicitly avoid claims about Plex/Jellyfin/Emby success. Log keeps DEBUG/WARNING/ERROR filters separate from destructive reset.
8. **Legacy destinations remain findable: PASS.** Independent keyboard navigation reached Playlist, XMLTV, Filter, Mapping, Users when enabled, Settings, Log, and Logout under their existing names; Overview, Connections, and Activity are additive. Numeric backend destinations continue through the explicit legacy adapter.

## Regression gates

- **Request and persistence parity: PASS.** Generated fixtures execute the exact legacy envelopes for Playlist/XMLTV CRUD and updates, Group/Custom Filter create/edit/delete, full Mapping save, Users create/edit/delete, partial Settings saves, authentication, backup, restore, stream probe, Log reset, and copy actions.
- **Mapping sizes: PASS/characterized.** Empty and small lineups passed in browser. Generated state tests cover `0`, `8`, `479`, `480`, `481`, and `5000` channels. The 5,000-row renderer is deliberately characterized by its generated-DOM performance/operability test rather than claimed as a browser fixture.
- **Keyboard and responsive behavior: PASS.** Browser passes covered every primary destination, tablet Mapping, dialog trapping/return, departure guard, menu expansion, and low-height setup. Generated DOM tests cover source/filter/user dialogs, Settings tabs, authentication labels, Log controls, sortable Mapping headers, continuous filter focus, and mobile navigation focus.
- **First-run safety: PASS.** Fresh settings create the UUID temp directory before permission checks; CLI bind survives network discovery and is persisted; the unconfigured process skips automatic binary update and SSDP; enabled SSDP starts exactly once after successful completion, while a start failure leaves the wizard active.
- **Embedded production assets: PASS.** Source TypeScript compiles into `html/js`; all scoped asset parity tests compare those files with `src/webUI.go`. The browser fixtures use the embedded binary without `-dev`.
- **Typed configuration API: PASS.** Tests cover the default-off CONFIG authority, generic fail-closed authentication, strict bounded describe/get/plan/apply schemas, HMAC-bound revisions/actions, typed mutation adapters, secret redaction/non-replay, provider SSRF/redirect/dial/body protections, and serialized shared mutation locks.

## Closed findings

- PMS Mapping generic empty state -> replaced with the client-managed/XEPG explanation and Settings route.
- Wizard changed the hash but stayed in the standalone setup document -> completion now forces the normal `/web/#overview` document reload.
- Ready fixture lacked XMLTV compatibility counts -> fixture corrected; real generated M3U/XMLTV content and copy actions verified.
- Empty Overview required a second Add click -> Add playlist/XMLTV actions open the corresponding source dialog directly.
- Source chooser lacked a reliable accessible dialog name/initial focus -> role, modal name, shown-focus, and invoker return verified.
- Filter empty copy contradicted default-selected streams -> copy now distinguishes absence of saved rules from current selected counts.
- Guide decision progress briefly changed denominator -> decision remains `2 of 4`; PMS shortens only after selection.
- Low-height Next action was below the initial viewport -> low-height fixed action verified at `780x437`.
- Mapping focus escape report -> not reproducible; native Tab cycles through every editor control and wraps inside the dialog.

## Bounded limitations

- No fixture supplied a representative live ERROR log entry; response-backed WARNING filtering passed in browser and generated ERROR rendering/filtering tests pass.
- The 5,000-row case is an automated generated-DOM characterization, not a claim of pleasant phone administration.
- Native select-option coordinates are a browser-harness limitation; standard keyboard selection passed.
