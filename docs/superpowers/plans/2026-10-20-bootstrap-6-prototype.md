# Bootstrap 6 functional prototype

Approved design: ../specs/2026-10-20-bootstrap-6-ui-modernization-design.md.
Acceptance: functional Mapping prototype, independently deployed in the cluster.

1. Pin Bootstrap 6 alpha and local frontend dependencies. Bundle the existing ordered controllers into one ESM entry per page, with explicit native Bootstrap imports. Keep tsc output available for existing controller tests. Remove CDN/runtime globals and bind page entry events in modules.
2. Introduce layered application tokens, system/light/dark preference and responsive navigation. Convert shared dialogs to native Bootstrap Dialog without changing backend transport or save semantics.
3. Modernize Mapping: native responsive Drawer, searchable native Combobox for EPG, derived removable filter chips, clear draft/save/rebuild state. Preserve bulk-edit, dirty-navigation and validation behavior.
4. Verify compiled/embedded assets, Go tests/vet, desktop/mobile browser workflows with independent synthetic fixtures. Review resulting diff and update Draft PR #37.
5. Publish a commit-specific preview image; deploy through a separate HomeOps GitOps PR with independent PVC and fixtures. Verify Flux, workload, endpoint and browser save/reload behavior. Keep production data and release tags untouched.

Implementation boundary: one ESM lexical scope preserves the existing controller collaboration while eliminating runtime Bootstrap globals. Deep imports include only used native plugins. This prototype does not require a wholesale controller module rewrite.

Validation evidence and remaining device qualification will be recorded with the implementation. Browser emulation does not certify physical Safari/iPad hardware.

## Follow-up: source and Settings forms

Authorized next slice: complete Playlist/XMLTV and Settings modernization before practical user trials.

- Render source and Settings fields with Bootstrap form-field/label/control markup and direct help/error associations; retain the existing control factories, command payloads and partial saves.
- Keep essential source identity/location and buffering controls visible; group proxy/header overrides in native details. Group advanced Settings by their existing section, without hiding primary EPG, schedule, streaming, backup or access controls.
- Show validation at the affected field, reveal its section/disclosure before focusing, and keep settings save/request feedback visible. Preserve values when switching tabs/disclosures and when save fails.
- Verify source payload parity, Settings types/labels, native browser edit/save/reload and error handling, responsive fit and focus. Regenerate embedded assets, review the completed patch, update Draft PR #37 and deploy only the isolated cluster preview through GitOps.
