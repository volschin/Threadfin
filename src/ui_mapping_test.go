package src

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func runMappingNodeFixture(t *testing.T, name, source string, assets ...string) []byte {
	t.Helper()
	script := filepath.Join(t.TempDir(), name+".js")
	if err := os.WriteFile(script, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	args := []string{script}
	for _, asset := range assets {
		args = append(args, filepath.Join("..", "html", "js", asset))
	}
	output, err := exec.Command("node", args...).CombinedOutput()
	if err != nil {
		t.Fatalf("execute generated Mapping fixture %s: %v\n%s", name, err, output)
	}
	return output
}

func TestUIMappingStateGeneratedFixtures(t *testing.T) {
	output := runMappingNodeFixture(t, "mapping-state", mappingStateNodeScript, "mapping_state_ts.js")
	var got struct {
		EmptyDefault  string              `json:"emptyDefault"`
		Default       string              `json:"default"`
		Reasons       map[string][]string `json:"reasons"`
		Sizes         map[string]int      `json:"sizes"`
		Duplicate     []string            `json:"duplicate"`
		Filtered      []string            `json:"filtered"`
		Selected      []string            `json:"selected"`
		Visible       int                 `json:"visible"`
		SelectedAll   int                 `json:"selectedAll"`
		ShiftSelected []string            `json:"shiftSelected"`
		SelectAll     []string            `json:"selectAll"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode Mapping state fixture: %v\n%s", err, output)
	}
	if got.EmptyDefault != "active" || got.Default != "attention" {
		t.Fatalf("Mapping default segments = empty %q populated %q, want active/attention", got.EmptyDefault, got.Default)
	}
	wantReasons := map[string][]string{
		"missing-file": {"missing"}, "missing-guide": {"missing"},
		"invalid-source": {"invalid"}, "invalid-guide": {"invalid"},
		"hidden": {"hidden"}, "inactive": {"inactive"},
		"valid": {}, "dummy": {},
	}
	if !reflect.DeepEqual(got.Reasons, wantReasons) {
		t.Fatalf("Mapping attention reasons = %#v, want %#v", got.Reasons, wantReasons)
	}
	for _, size := range []string{"0", "8", "479", "480", "481", "5000"} {
		want, _ := map[string]int{"0": 0, "8": 8, "479": 479, "480": 480, "481": 481, "5000": 5000}[size]
		if got.Sizes[size] != want {
			t.Errorf("Mapping generated-JS fixture %s rows = %d, want %d", size, got.Sizes[size], want)
		}
	}
	if !reflect.DeepEqual(got.Duplicate, []string{"duplicate-a", "duplicate-b", "after"}) {
		t.Fatalf("stable duplicate-number sort = %v", got.Duplicate)
	}
	if !reflect.DeepEqual(got.Filtered, []string{"hidden"}) || !reflect.DeepEqual(got.Selected, []string{"hidden", "valid"}) {
		t.Fatalf("search/filter changed ID selection: filtered=%v selected=%v", got.Filtered, got.Selected)
	}
	if got.Visible != 1 || got.SelectedAll != 2 {
		t.Fatalf("Mapping selection counts = visible %d total %d, want 1/2", got.Visible, got.SelectedAll)
	}
	if !reflect.DeepEqual(got.ShiftSelected, []string{"missing-file", "missing-guide", "valid"}) || !reflect.DeepEqual(got.SelectAll, []string{"hidden", "invalid-guide"}) {
		t.Fatalf("Mapping shift/select-all behavior = shift %v visible all %v", got.ShiftSelected, got.SelectAll)
	}
}

func TestUIMappingDraftMutationsGeneratedFixture(t *testing.T) {
	output := runMappingNodeFixture(t, "mapping-mutations", mappingMutationNodeScript, "mapping_state_ts.js")
	var got struct {
		Single       map[string]any `json:"single"`
		Bulk         map[string]any `json:"bulk"`
		BulkChannel  map[string]any `json:"bulkChannel"`
		Dirty        []string       `json:"dirty"`
		Reverted     bool           `json:"reverted"`
		Unknown      any            `json:"unknown"`
		Collision    string         `json:"collision"`
		Invalid      bool           `json:"invalid"`
		Dummy        map[string]any `json:"dummy"`
		OriginalDeep string         `json:"originalDeep"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode Mapping mutation fixture: %v\n%s", err, output)
	}
	for key, want := range map[string]any{
		"x-channelID": "101", "x-active": false, "x-name": "Renamed", "x-description": "Description",
		"x-update-channel-name": true, "tvg-logo": "logo.png", "x-update-channel-icon": true,
		"x-category": "news", "x-group-title": "News", "x-xmltv-file": "-", "x-mapping": "-",
		"x-ppv-extra": "Extra", "x-backup-channel-1": "Backup One", "x-backup-channel-2": "Backup Two", "x-backup-channel-3": "Backup Three",
		"x-hide-channel": true,
	} {
		if got.Single[key] != want {
			t.Errorf("single Mapping mutation %s = %#v, want %#v", key, got.Single[key], want)
		}
	}
	if got.Collision != "101" {
		t.Fatalf("single-number collision result = %q, want 101", got.Collision)
	}
	if got.Bulk["a"] != "200" || got.Bulk["b"] != "201" || got.Bulk["c"] != "202" {
		t.Fatalf("bulk sequential numbering = %#v", got.Bulk)
	}
	for key, want := range map[string]any{"x-description": "Bulk", "x-active": true, "x-category": "bulk", "x-group-title": "Bulk group", "x-xmltv-file": "Threadfin Dummy", "x-mapping": "60_Minutes", "x-backup-channel-1": "Bulk backup"} {
		if got.BulkChannel[key] != want {
			t.Errorf("bulk Mapping mutation %s = %#v, want %#v", key, got.BulkChannel[key], want)
		}
	}
	if got.Invalid || !reflect.DeepEqual(got.Dirty, []string{"a", "b", "c"}) || !got.Reverted {
		t.Fatalf("Mapping dirty/revert/validation contract = invalid %t dirty %v reverted %t", got.Invalid, got.Dirty, got.Reverted)
	}
	if got.Unknown == nil || got.OriginalDeep != "preserved" {
		t.Fatalf("unknown/deep properties were not preserved: unknown=%#v original=%q", got.Unknown, got.OriginalDeep)
	}
	for key, want := range map[string]any{"x-xmltv-file": "Threadfin Dummy", "x-mapping": "PPV", "x-active": true, "x-update-channel-icon": false} {
		if got.Dummy[key] != want {
			t.Errorf("dummy assignment %s = %#v, want %#v", key, got.Dummy[key], want)
		}
	}
}

func TestUIMappingBulkVisibleOrderGeneratedFixture(t *testing.T) {
	output := runMappingNodeFixture(t, "mapping-bulk-order", mappingBulkOrderNodeScript, "mapping_state_ts.js")
	var got struct {
		VisibleSelected []string          `json:"visibleSelected"`
		Numbers         map[string]string `json:"numbers"`
		HiddenNumber    string            `json:"hiddenNumber"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode Mapping bulk-order fixture: %v\n%s", err, output)
	}
	if !reflect.DeepEqual(got.VisibleSelected, []string{"b", "a"}) {
		t.Fatalf("Mapping visible selected order = %v, want sorted/filter order [b a]", got.VisibleSelected)
	}
	if !reflect.DeepEqual(got.Numbers, map[string]string{"a": "301", "b": "300"}) || got.HiddenNumber != "30" {
		t.Fatalf("Mapping visible bulk numbering = %#v hidden=%q", got.Numbers, got.HiddenNumber)
	}
}

func TestUIMappingSaveAndTransportGeneratedFixture(t *testing.T) {
	output := runMappingNodeFixture(t, "mapping-save", mappingSaveNodeScript, "mapping_state_ts.js", "mapping_ts.js")
	var got struct {
		Payload          map[string]any `json:"payload"`
		Sync             string         `json:"sync"`
		Queued           string         `json:"queued"`
		BusyDirty        int            `json:"busyDirty"`
		FailedDirty      int            `json:"failedDirty"`
		PersistedDirty   int            `json:"persistedDirty"`
		Ambiguous        string         `json:"ambiguous"`
		RefetchCommand   string         `json:"refetchCommand"`
		RefetchedDirty   int            `json:"refetchedDirty"`
		ProbePayload     map[string]any `json:"probePayload"`
		ProbeEmpty       string         `json:"probeEmpty"`
		ProbeSuccess     string         `json:"probeSuccess"`
		PendingBlocked   bool           `json:"pendingBlocked"`
		PendingName      string         `json:"pendingName"`
		FailedRaceName   string         `json:"failedRaceName"`
		FailedRaceDirty  int            `json:"failedRaceDirty"`
		AmbiguousBlocked bool           `json:"ambiguousBlocked"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode Mapping save fixture: %v\n%s", err, output)
	}
	wantMap := map[string]any{"a": map[string]any{"x-channelID": "9", "x-active": true, "x-name": "Changed", "x-xmltv-file": "X1.xml", "x-mapping": "guide-1", "unknown": map[string]any{"deep": "preserved"}}}
	if got.Payload["cmd"] != "saveEpgMapping" || !reflect.DeepEqual(got.Payload["data"], map[string]any{"epgMapping": wantMap}) {
		t.Fatalf("Mapping save envelope changed: %#v", got.Payload)
	}
	if got.Sync != "Mapping saved; outputs rebuilt" || got.Queued != "Mapping saved; output rebuild requested" {
		t.Fatalf("response-backed Mapping copy = sync %q queued %q", got.Sync, got.Queued)
	}
	if got.BusyDirty != 1 || got.FailedDirty != 1 || got.PersistedDirty != 0 {
		t.Fatalf("Mapping save reconciliation dirty counts = busy %d failed %d persisted %d", got.BusyDirty, got.FailedDirty, got.PersistedDirty)
	}
	if got.Ambiguous != "ambiguous" || got.RefetchCommand != "getServerConfig" || got.RefetchedDirty != 0 {
		t.Fatalf("Mapping ambiguous transport/refetch contract = state %q command %q dirty %d", got.Ambiguous, got.RefetchCommand, got.RefetchedDirty)
	}
	if !got.PendingBlocked || got.PendingName != "Snapshot" || got.FailedRaceName != "Failed snapshot" || got.FailedRaceDirty != 1 || !got.AmbiguousBlocked {
		t.Fatalf("Mapping post-submit mutation lock = pending blocked %t name %q, status-false name %q dirty %d, ambiguous blocked %t", got.PendingBlocked, got.PendingName, got.FailedRaceName, got.FailedRaceDirty, got.AmbiguousBlocked)
	}
	if !reflect.DeepEqual(got.ProbePayload, map[string]any{"cmd": "probeChannel", "data": map[string]any{"probeUrl": "http://127.0.0.1/stream"}}) || got.ProbeEmpty != "Details unavailable" || !strings.Contains(got.ProbeSuccess, "1920x1080") {
		t.Fatalf("Mapping probe contract = payload %#v empty %q success %q", got.ProbePayload, got.ProbeEmpty, got.ProbeSuccess)
	}
}

func TestUIMappingNavigationGuardGeneratedFixture(t *testing.T) {
	output := runMappingNodeFixture(t, "mapping-navigation", mappingNavigationNodeScript, "navigation_ts.js")
	var got struct {
		Guarded      []string `json:"guarded"`
		Shows        []int    `json:"shows"`
		Pushes       []string `json:"pushes"`
		Replacements []string `json:"replacements"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode Mapping navigation fixture: %v\n%s", err, output)
	}
	if !reflect.DeepEqual(got.Guarded, []string{"settings", "logout", "log", "history:settings"}) {
		t.Fatalf("Mapping destination/logout/programmatic guards = %v", got.Guarded)
	}
	if !reflect.DeepEqual(got.Shows, []int{5, 7, 6, 5}) || !reflect.DeepEqual(got.Pushes, []string{"#settings", "#logout"}) {
		t.Fatalf("guard continuations did not resume exact navigation: shows=%v pushes=%v", got.Shows, got.Pushes)
	}
	if !reflect.DeepEqual(got.Replacements, []string{"#log", "#mapping", "#settings"}) {
		t.Fatalf("programmatic legacy destination did not preserve numeric/history contract: %v", got.Replacements)
	}
}

func TestUIMappingGeneratedDOMAccessibilityAndEditor(t *testing.T) {
	output := runMappingNodeFixture(t, "mapping-dom", mappingDOMNodeScript, "mapping_state_ts.js", "mapping_ts.js", "navigation_ts.js")
	var got struct {
		PageText                 string `json:"pageText"`
		EditorText               string `json:"editorText"`
		SelectAllLabel           string `json:"selectAllLabel"`
		EditorRole               string `json:"editorRole"`
		EditorLabelledBy         string `json:"editorLabelledBy"`
		AdvancedInitially        bool   `json:"advancedInitially"`
		AdvancedSearch           bool   `json:"advancedSearch"`
		BackupLists              int    `json:"backupLists"`
		SaveFocus                int    `json:"saveFocus"`
		DraftName                string `json:"draftName"`
		LiveStatus               string `json:"liveStatus"`
		UnloadPrevented          int    `json:"unloadPrevented"`
		UnloadReturnValue        string `json:"unloadReturnValue"`
		StayContinued            int    `json:"stayContinued"`
		DiscardContinued         int    `json:"discardContinued"`
		SaveContinued            int    `json:"saveContinued"`
		PendingSaveDisabled      bool   `json:"pendingSaveDisabled"`
		PendingDiscardDisabled   bool   `json:"pendingDiscardDisabled"`
		PendingStayEnabled       bool   `json:"pendingStayEnabled"`
		PendingStayClosed        bool   `json:"pendingStayClosed"`
		AmbiguousStayClosed      bool   `json:"ambiguousStayClosed"`
		TypedSearch              string `json:"typedSearch"`
		SearchFocused            bool   `json:"searchFocused"`
		SearchCaret              int    `json:"searchCaret"`
		FilterFocused            bool   `json:"filterFocused"`
		SelectAllFocused         bool   `json:"selectAllFocused"`
		KeyboardSelected         bool   `json:"keyboardSelected"`
		KeyboardFocused          bool   `json:"keyboardFocused"`
		PendingMutationsDisabled bool   `json:"pendingMutationsDisabled"`
		PendingEditorBlocked     bool   `json:"pendingEditorBlocked"`
		InitialEditorFocus       string `json:"initialEditorFocus"`
		TabTrapped               bool   `json:"tabTrapped"`
		ShiftTabTrapped          bool   `json:"shiftTabTrapped"`
		EditorEscapeClosed       bool   `json:"editorEscapeClosed"`
		EditorInvokerRestored    bool   `json:"editorInvokerRestored"`
		EditorEscapeDraft        string `json:"editorEscapeDraft"`
		GuardEscapeClosed        bool   `json:"guardEscapeClosed"`
		GuardEscapeContinued     int    `json:"guardEscapeContinued"`
		InitialSort              string `json:"initialSort"`
		ToggledSort              string `json:"toggledSort"`
		SidebarReturnFocus       string `json:"sidebarReturnFocus"`
		PendingReturnFocus       string `json:"pendingReturnFocus"`
		PendingSaveAvoided       bool   `json:"pendingSaveAvoided"`
		AmbiguousReturnFocus     string `json:"ambiguousReturnFocus"`
		RemovedReturnFocus       string `json:"removedReturnFocus"`
		ProgrammaticReturnFocus  string `json:"programmaticReturnFocus"`
		HistoryReturnFocus       string `json:"historyReturnFocus"`
		PMSText                  string `json:"pmsText"`
		PMSSettingsHash          string `json:"pmsSettingsHash"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode Mapping DOM fixture: %v\n%s", err, output)
	}
	for _, want := range []string{"Needs attention", "Active", "Inactive", "Missing EPG assignment", "Hidden from outputs", "results", "selected"} {
		if !strings.Contains(got.PageText, want) {
			t.Errorf("rendered Mapping page missing %q: %q", want, got.PageText)
		}
	}
	if strings.Contains(got.PageText, "Online") || strings.Contains(got.PageText, "Offline") {
		t.Fatal("rendered Mapping page invents channel reachability")
	}
	for _, want := range []string{"Active", "Channel number", "Name", "XMLTV source", "XMLTV channel", "Description", "Update channel name", "Logo", "Update channel logo", "Category", "Group", "PPV extra", "Backup channel 1", "Backup channel 2", "Backup channel 3", "Hidden from outputs", "Probe channel", "Apply to draft"} {
		if !strings.Contains(got.EditorText, want) {
			t.Errorf("Mapping editor missing %q: %q", want, got.EditorText)
		}
	}
	if got.SelectAllLabel != "Select all visible channels" || got.EditorRole != "dialog" || got.EditorLabelledBy != "mapping-editor-title" || got.AdvancedInitially || !got.AdvancedSearch {
		t.Fatalf("Mapping accessibility/editor disclosure contract = %+v", got)
	}
	if got.BackupLists != 1 || got.SaveFocus < 1 || got.DraftName != "Keyboard renamed" || got.LiveStatus != "polite" {
		t.Fatalf("Mapping shared options/apply/save focus contract = %+v", got)
	}
	if got.UnloadPrevented != 1 || got.UnloadReturnValue != "" || got.StayContinued != 0 || got.DiscardContinued != 1 || got.SaveContinued != 0 || !got.PendingSaveDisabled || !got.PendingDiscardDisabled || !got.PendingStayEnabled || !got.PendingStayClosed || !got.AmbiguousStayClosed {
		t.Fatalf("Mapping unload/Save-Discard-Stay guard contract = %+v", got)
	}
	if got.TypedSearch != "hi" || !got.SearchFocused || got.SearchCaret != 2 || !got.FilterFocused || !got.SelectAllFocused || !got.KeyboardSelected || !got.KeyboardFocused {
		t.Fatalf("Mapping continuous filter/selection keyboard contract = %+v", got)
	}
	if !got.PendingMutationsDisabled || !got.PendingEditorBlocked {
		t.Fatalf("Mapping pending mutation surfaces were not locked = %+v", got)
	}
	if got.InitialEditorFocus != "x-active" || !got.TabTrapped || !got.ShiftTabTrapped || !got.EditorEscapeClosed || !got.EditorInvokerRestored || got.EditorEscapeDraft != "Valid" || !got.GuardEscapeClosed || got.GuardEscapeContinued != 0 {
		t.Fatalf("Mapping dialog keyboard contract = %+v", got)
	}
	if got.InitialSort != "ascending" || got.ToggledSort != "descending" {
		t.Fatalf("Mapping sortable header semantics = initial %q toggled %q", got.InitialSort, got.ToggledSort)
	}
	if got.SidebarReturnFocus != "settings" || got.PendingReturnFocus != "mapping-heading" || !got.PendingSaveAvoided || got.AmbiguousReturnFocus != "mapping-heading" || got.RemovedReturnFocus != "mapping-heading" || got.ProgrammaticReturnFocus != "mapping-heading" || got.HistoryReturnFocus != "mapping-heading" {
		t.Fatalf("Mapping guard focus return contract = %+v", got)
	}
	if !strings.Contains(got.PMSText, "PMS mode") || !strings.Contains(got.PMSText, "XEPG") || !strings.Contains(got.PMSText, "client") || got.PMSSettingsHash != "#settings" {
		t.Fatalf("PMS Mapping explanation/settings route = text %q route %q", got.PMSText, got.PMSSettingsHash)
	}
}

func TestUIMappingGeneratedNetworkQueueAndTransport(t *testing.T) {
	output := runMappingNodeFixture(t, "mapping-network", mappingNetworkNodeScript, "mapping_state_ts.js", "mapping_ts.js", "network_ts.js")
	var got struct {
		QueuedDirty    int    `json:"queuedDirty"`
		QueuedState    string `json:"queuedState"`
		TransportDirty int    `json:"transportDirty"`
		TransportState string `json:"transportState"`
		RefetchCommand string `json:"refetchCommand"`
		RefetchPending bool   `json:"refetchPending"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode Mapping network fixture: %v\n%s", err, output)
	}
	if got.QueuedDirty != 1 || !strings.Contains(got.QueuedState, "retained") || got.TransportDirty != 1 || got.TransportState != "ambiguous" || got.RefetchCommand != "getServerConfig" || !got.RefetchPending {
		t.Fatalf("Mapping generated network queue/transport contract = %+v", got)
	}
}

func TestUIMappingGeneratedRendererFiveThousandRows(t *testing.T) {
	output := runMappingNodeFixture(t, "mapping-renderer-large", mappingLargeRendererNodeScript, "mapping_state_ts.js", "mapping_ts.js")
	var got struct {
		Rows         int   `json:"rows"`
		BackupLists  int   `json:"backupLists"`
		EditorOpen   bool  `json:"editorOpen"`
		EditorUsable bool  `json:"editorUsable"`
		ElapsedMS    int64 `json:"elapsedMs"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode Mapping 5,000-row renderer fixture: %v\n%s", err, output)
	}
	if got.Rows != 5000 || got.BackupLists != 1 || !got.EditorOpen || !got.EditorUsable {
		t.Fatalf("Mapping 5,000-row renderer contract = %+v", got)
	}
	if got.ElapsedMS > 15000 {
		t.Fatalf("Mapping synthetic 5,000-row render took %dms, ceiling 15000ms", got.ElapsedMS)
	}
	t.Logf("generated Mapping renderer built 5,000 rows in %dms (synthetic DOM; real-browser measurement remains Task 10)", got.ElapsedMS)
}

func TestUIMappingResponseEvidenceJSON(t *testing.T) {
	for _, result := range []MappingSaveResult{MappingOutputsRebuilt, MappingOutputRebuildRequested} {
		encoded, err := json.Marshal(ResponseStruct{MappingSaveResult: result})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(encoded), `"mappingSaveResult":"`+string(result)+`"`) {
			t.Fatalf("Mapping response evidence missing from %s", encoded)
		}
	}
}

func TestUIMappingAssetsAndEmbedParity(t *testing.T) {
	index, err := os.ReadFile(filepath.Join("..", "html", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	indexText := string(index)
	stateAt := strings.Index(indexText, `src="js/mapping_state_ts.js"`)
	pageAt := strings.Index(indexText, `src="js/mapping_ts.js"`)
	navigationAt := strings.Index(indexText, `src="js/navigation_ts.js"`)
	if stateAt < 0 || pageAt <= stateAt || navigationAt <= pageAt {
		t.Fatalf("Mapping scripts are absent or ordered incorrectly: state=%d page=%d navigation=%d", stateAt, pageAt, navigationAt)
	}
	styles, err := os.ReadFile(filepath.Join("..", "html", "css", "app-shell.css"))
	if err != nil {
		t.Fatal(err)
	}
	for _, selector := range []string{".tf-app .tf-mapping", ".tf-app .tf-mapping-save-bar", ".tf-app .tf-mapping-editor", "@media (max-width: 52rem)"} {
		if !strings.Contains(string(styles), selector) {
			t.Errorf("Mapping stylesheet missing %q", selector)
		}
	}
	language, err := os.ReadFile(filepath.Join("..", "html", "lang", "en.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, copy := range []string{"Needs attention", "Missing EPG assignment", "Invalid EPG assignment", "Hidden from outputs", "Apply to draft", "Save mapping", "Discard draft", "Stay"} {
		if !strings.Contains(string(language), copy) {
			t.Errorf("English Mapping copy missing %q", copy)
		}
	}
	if strings.Contains(strings.ToLower(string(language)[strings.Index(string(language), `"mapping"`):]), "online") || strings.Contains(strings.ToLower(string(language)[strings.Index(string(language), `"mapping"`):]), "offline") {
		t.Fatal("Mapping language invents per-channel Online/Offline state")
	}

	webUI = make(map[string]interface{})
	loadHTMLMap()
	for _, name := range []string{"html/index.html", "html/css/app-shell.css", "html/js/mapping_state_ts.js", "html/js/mapping_ts.js", "html/lang/en.json"} {
		t.Run(filepath.Base(name), func(t *testing.T) {
			encoded, ok := webUI[name].(string)
			if !ok {
				t.Fatalf("embedded Mapping asset %q is missing", name)
			}
			got, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(filepath.Join("..", name))
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != string(want) {
				t.Errorf("embedded Mapping asset %q does not match source", name)
			}
		})
	}
}

const mappingStateNodeScript = `
const fs = require("fs"), vm = require("vm");
const context = {console: {log() {}, warn() {}}}; vm.createContext(context); vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), context);
function channel(number, extra) { return Object.assign({"x-channelID": String(number), "x-active": true, "x-name": "Channel " + number, "_file.m3u.id": "M1", "x-group-title": "General", "x-xmltv-file": "X1.xml", "x-mapping": "guide-1"}, extra || {}); }
const mapping = {
  valid: channel(1), "missing-file": channel(2, {"x-xmltv-file": "-"}), "missing-guide": channel(3, {"x-mapping": "-"}),
  "invalid-source": channel(4, {"x-xmltv-file": "X9.xml"}), "invalid-guide": channel(5, {"x-mapping": "absent"}),
  hidden: channel(6, {"x-hide-channel": true, "x-name": "Hidden Sports"}), inactive: channel(7, {"x-active": false}),
  dummy: channel(8, {"x-xmltv-file": "Threadfin Dummy", "x-mapping": "360_Minutes"})
};
const server = {xepg: {epgMapping: mapping, xmltvMap: {"X1.xml": {"guide-1": {"display-name": "Guide One"}}, "Threadfin Dummy": {}}}};
const state = context.createMappingWorkspaceState(server); const empty = context.createMappingWorkspaceState({xepg: {epgMapping: {}, xmltvMap: {}}});
const reasons = {}; Object.keys(mapping).forEach(id => reasons[id] = context.mappingAttentionReasons(mapping[id], server.xepg.xmltvMap));
const sizes = {}; [0, 8, 479, 480, 481, 5000].forEach(size => { const rows = {}; for (let i = 0; i < size; i++) rows["id-" + i] = channel(i + 1); sizes[String(size)] = context.mappingVisibleRows(context.createMappingWorkspaceState({xepg: {epgMapping: rows, xmltvMap: server.xepg.xmltvMap}}), {segment: "active"}).length; });
const duplicateState = context.createMappingWorkspaceState({xepg: {epgMapping: {after: channel(11), "duplicate-a": channel(10), "duplicate-b": channel(10)}, xmltvMap: server.xepg.xmltvMap}});
const duplicate = context.mappingVisibleRows(duplicateState, {segment: "active", sort: "number"}).map(row => row.id);
context.mappingSetSelected(state, "valid", true); context.mappingSetSelected(state, "hidden", true);
const filtered = context.mappingVisibleRows(state, {segment: "attention", search: "sports", playlist: "M1", group: "General", xmltv: "X1.xml", activation: "active", reason: "hidden"}).map(row => row.id);
const rangeState = context.createMappingWorkspaceState(server); context.mappingSetSelected(rangeState, "valid", true); context.mappingSelectRange(rangeState, ["valid", "missing-file", "missing-guide"], "missing-guide", true, true); const shiftSelected = Array.from(rangeState.selected).sort(); rangeState.selected.clear(); context.mappingSelectVisible(rangeState, ["hidden", "invalid-guide"], true); const selectAll = Array.from(rangeState.selected).sort();
process.stdout.write(JSON.stringify({emptyDefault: empty.segment, default: state.segment, reasons, sizes, duplicate, filtered, selected: Array.from(state.selected).sort(), visible: context.mappingSelectedVisibleCount(state, filtered), selectedAll: state.selected.size, shiftSelected, selectAll}));
`

const mappingMutationNodeScript = `
const fs = require("fs"), vm = require("vm"); const context = {console: {log() {}, warn() {}}}; vm.createContext(context); vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), context);
function channel(number) { return {"x-channelID": String(number), "x-active": true, "x-name": "Original", "x-description": "", "x-update-channel-name": false, "tvg-logo": "", "x-update-channel-icon": false, "x-category": "", "x-group-title": "", "x-xmltv-file": "X1.xml", "x-mapping": "guide-1", "x-ppv-extra": "", "x-backup-channel-1": "-", "x-backup-channel-2": "-", "x-backup-channel-3": "-", "x-hide-channel": false, unknown: {deep: "preserved"}}; }
const source = {a: channel(99), b: channel(100), c: channel(102)}; const state = context.createMappingWorkspaceState({xepg: {epgMapping: source, xmltvMap: {"X1.xml": {"guide-1": {}}}}});
const originalDeep = source.a.unknown.deep;
const applied = context.mappingApplyChannelPatch(state, ["a"], {"x-channelID": "100", "x-active": true, "x-name": "Renamed", "x-description": "Description", "x-update-channel-name": true, "tvg-logo": "logo.png", "x-update-channel-icon": true, "x-category": "news", "x-group-title": "News", "x-xmltv-file": "-", "x-mapping": "-", "x-ppv-extra": "Extra", "x-backup-channel-1": "Backup One", "x-backup-channel-2": "Backup Two", "x-backup-channel-3": "Backup Three", "x-hide-channel": true}, {numberChanged: true});
const collision = state.draft.a["x-channelID"];
const single = JSON.parse(JSON.stringify(state.draft.a));
const bulkResult = context.mappingApplyChannelPatch(state, ["a", "b", "c"], {"x-description": "Bulk", "x-active": true, "x-category": "bulk", "x-group-title": "Bulk group", "x-xmltv-file": "Threadfin Dummy", "x-mapping": "60_Minutes", "x-backup-channel-1": "Bulk backup"}, {sequentialStart: "200"});
const invalid = context.mappingApplyChannelPatch(state, ["a"], {}, {sequentialStart: "not-a-number"}).ok;
const dirty = context.mappingDirtyIDs(state).sort(); const unknown = state.draft.a.unknown; const bulk = {a: state.draft.a["x-channelID"], b: state.draft.b["x-channelID"], c: state.draft.c["x-channelID"]}; const bulkChannel = JSON.parse(JSON.stringify(state.draft.b));
context.mappingAssignDummy(state, ["b"], "PPV"); const dummy = JSON.parse(JSON.stringify(state.draft.b)); context.mappingRevertDraft(state); const reverted = context.mappingDirtyIDs(state).length === 0 && state.draft.a["x-name"] === "Original";
process.stdout.write(JSON.stringify({single, bulk, bulkChannel, dirty, reverted, unknown, collision, invalid, dummy, originalDeep, applied: applied.ok, bulkResult: bulkResult.ok}));
`

const mappingBulkOrderNodeScript = `
const fs = require("fs"), vm = require("vm"); const context = {console: {log() {}, warn() {}}}; vm.createContext(context); vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), context);
function channel(number, name, group) { return {"x-channelID": String(number), "x-active": true, "x-name": name, "x-group-title": group, "x-xmltv-file": "X1.xml", "x-mapping": "guide-1"}; }
const state = context.createMappingWorkspaceState({xepg: {epgMapping: {a: channel(20, "Zulu", "shown"), b: channel(10, "Alpha", "shown"), hidden: channel(30, "Hidden", "hidden")}, xmltvMap: {"X1.xml": {"guide-1": {}}}}});
context.mappingSetSelected(state, "hidden", true); context.mappingSetSelected(state, "a", true); context.mappingSetSelected(state, "b", true);
const rows = context.mappingVisibleRows(state, {segment: "active", group: "shown", sort: "number", descending: false});
const visibleSelected = context.mappingSelectedVisibleIDs(state, rows); context.mappingApplyChannelPatch(state, visibleSelected, {}, {sequentialStart: "300"});
process.stdout.write(JSON.stringify({visibleSelected, numbers: {a: state.draft.a["x-channelID"], b: state.draft.b["x-channelID"]}, hiddenNumber: state.draft.hidden["x-channelID"]}));
`

const mappingSaveNodeScript = `
const fs = require("fs"), vm = require("vm"); const sent = [];
class CaptureServer { constructor(cmd) { this.cmd = cmd; } request(data) { sent.push({cmd: this.cmd, data: JSON.parse(JSON.stringify(data))}); } }
const context = {console: {log() {}, warn() {}}, Server: CaptureServer, document: {addEventListener() {}, getElementById() { return null; }}, window: {addEventListener() {}}, SERVER: {}, SERVER_CONNECTION: false};
vm.createContext(context); vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), context); vm.runInContext(fs.readFileSync(process.argv[3], "utf8"), context);
function authoritative(name) { return {xepg: {epgMapping: {a: {"x-channelID": "9", "x-active": true, "x-name": name, "x-xmltv-file": "X1.xml", "x-mapping": "guide-1", unknown: {deep: "preserved"}}}, xmltvMap: {"X1.xml": {"guide-1": {}}}}}; }
function start(name) { context.initializeMappingWorkspace(authoritative("Before"), true); context.mappingApplyChannelPatch(context.mappingWorkspaceState, ["a"], {"x-name": name}); context.saveMappingDraft(); return JSON.parse(JSON.stringify(sent.pop())); }
const payload = start("Changed"); context.completeMappingRequest("saveEpgMapping", payload.data, Object.assign({status: true, mappingSaveResult: "outputsRebuilt"}, authoritative("Changed"))); const sync = context.mappingWorkspaceState.feedback;
const queuedPayload = start("Queued"); context.completeMappingRequest("saveEpgMapping", queuedPayload.data, Object.assign({status: true, mappingSaveResult: "outputRebuildRequested"}, authoritative("Queued"))); const queued = context.mappingWorkspaceState.feedback;
const busyPayload = start("Busy"); context.completeMappingRequest("saveEpgMapping", busyPayload.data, {status: false, err: "busy"}, "busy"); const busyDirty = context.mappingDirtyIDs(context.mappingWorkspaceState).length;
const failedPayload = start("Failed"); context.completeMappingRequest("saveEpgMapping", failedPayload.data, Object.assign({status: false, err: "not persisted"}, authoritative("Before"))); const failedDirty = context.mappingDirtyIDs(context.mappingWorkspaceState).length;
const persistedPayload = start("Persisted"); context.completeMappingRequest("saveEpgMapping", persistedPayload.data, Object.assign({status: false, err: "output failed"}, authoritative("Persisted"))); const persistedDirty = context.mappingDirtyIDs(context.mappingWorkspaceState).length;
const ambiguousPayload = start("Maybe"); context.completeMappingRequest("saveEpgMapping", ambiguousPayload.data, {status: false, err: "transport"}, "transport"); const ambiguous = context.mappingWorkspaceState.saveState; const refetch = sent.pop();
const ambiguousBlocked = context.mappingApplyChannelPatch(context.mappingWorkspaceState, ["a"], {"x-name": "Too late"}).ok === false;
context.completeMappingRequest("getServerConfig", {}, Object.assign({status: true}, authoritative("Maybe"))); const refetchedDirty = context.mappingDirtyIDs(context.mappingWorkspaceState).length;
const snapshotPayload = start("Snapshot"); const pendingBlocked = context.mappingApplyChannelPatch(context.mappingWorkspaceState, ["a"], {"x-name": "Late edit"}).ok === false; const pendingName = context.mappingWorkspaceState.draft.a["x-name"]; context.completeMappingRequest("saveEpgMapping", snapshotPayload.data, Object.assign({status: true, mappingSaveResult: "outputsRebuilt"}, authoritative("Snapshot")));
const failedRacePayload = start("Failed snapshot"); context.mappingApplyChannelPatch(context.mappingWorkspaceState, ["a"], {"x-name": "Late failed edit"}); context.completeMappingRequest("saveEpgMapping", failedRacePayload.data, Object.assign({status: false, err: "not persisted"}, authoritative("Before"))); const failedRaceName = context.mappingWorkspaceState.draft.a["x-name"], failedRaceDirty = context.mappingDirtyIDs(context.mappingWorkspaceState).length;
context.requestMappingProbe("http://127.0.0.1/stream"); const probePayload = sent.pop(); context.completeMappingRequest("probeChannel", {probeUrl: "http://127.0.0.1/stream"}, {status: true, probeInfo: {}}); const probeEmpty = context.mappingProbeFeedback; context.completeMappingRequest("probeChannel", {probeUrl: "http://127.0.0.1/stream"}, {status: true, probeInfo: {resolution: "1920x1080", frameRate: "50", audioChannel: "stereo"}}); const probeSuccess = context.mappingProbeFeedback;
process.stdout.write(JSON.stringify({payload, sync, queued, busyDirty, failedDirty, persistedDirty, ambiguous, refetchCommand: refetch.cmd, refetchedDirty, pendingBlocked, pendingName, failedRaceName, failedRaceDirty, ambiguousBlocked, probePayload, probeEmpty, probeSuccess}));
`

const mappingNavigationNodeScript = `
const fs = require("fs"), vm = require("vm"); const guarded = [], shows = [], pushes = [], replacements = []; let continuation;
class ShowContent { constructor(index) { this.index = index; } show() { shows.push(this.index); } }
const history = {state: null, pushState(_state, _unused, hash) { pushes.push(hash); }, replaceState(_state, _unused, hash) { replacements.push(hash); }};
const context = {console: {log() {}, warn() {}}, menuItems: Array.from({length: 8}, (_, i) => ({value: ["Playlist","XMLTV","Filter","Mapping","Users","Settings","Log","Logout"][i], initializeTableHeader() {}})), SERVER: {settings: {"authentication.web": true}}, ShowContent,
  document: {getElementById() { return null; }, querySelectorAll() { return []; }, querySelector() { return null; }}, window: {history, location: {hash: ""}, addEventListener() {}, setTimeout(fn) { fn(); }},
  renderOverview() {}, renderConnections() {}, renderActivity() {}, renderNavigation() {}, enableGroupSelection() {}, renderLegacyMenuAdapters() {}, bindServerInformationCompatibilityLink() {},
  mappingHasDirtyDraft() { return true; }, mappingRequestNavigation(next) { continuation = next; return false; }
};
vm.createContext(context); vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), context); context.currentDestination = "mapping";
function guardedDestination(destination, mode) { continuation = undefined; mode(); if (!continuation) throw new Error("navigation was not guarded: " + destination); guarded.push(destination); const resume = continuation; continuation = undefined; resume(); }
guardedDestination("settings", () => context.openDestination("settings", true)); context.currentDestination = "mapping";
guardedDestination("logout", () => context.openDestination("logout", true)); context.currentDestination = "mapping";
guardedDestination("log", () => context.openLegacyMenu(6, true));
context.currentDestination = "mapping"; history.state = {threadfinDestination: "settings"}; continuation = undefined; context.restoreDestinationFromHistory(); if (!continuation) throw new Error("history navigation was not guarded"); guarded.push("history:settings"); const historyResume = continuation; continuation = undefined; historyResume();
process.stdout.write(JSON.stringify({guarded, shows, pushes, replacements}));
`

const mappingDOMNodeScript = `
const fs = require("fs"), vm = require("vm");
let document;
class Element {
  constructor(tag) { this.tagName = String(tag).toUpperCase(); this.children = []; this.parentElement = null; this.attributes = {}; this.className = ""; this.textContent = ""; this.hidden = false; this.value = ""; this.name = ""; this.type = ""; this.checked = false; this.disabled = false; this.open = false; this.listeners = {}; this.focusCount = 0; this.selectionStart = 0; this.selectionEnd = 0; }
  set innerHTML(value) { this.children = []; this.textContent = value; }
  get innerHTML() { return this.textContent; }
  appendChild(child) { child.parentElement = this; this.children.push(child); return child; }
  removeChild(child) { this.children = this.children.filter(item => item !== child); child.parentElement = null; return child; }
  setAttribute(name, value) { this.attributes[name] = String(value); if (name === "name") this.name = String(value); if (name === "id") this.id = String(value); }
  getAttribute(name) { return this.attributes[name]; }
  removeAttribute(name) { delete this.attributes[name]; }
  addEventListener(name, listener) { this.listeners[name] = listener; }
  focus() { this.focusCount++; document.activeElement = this; }
  setSelectionRange(start, end) { this.selectionStart = start; this.selectionEnd = end; }
  contains(target) { return this === target || this.children.some(child => child.contains && child.contains(target)); }
  querySelector(selector) { return descendants(this).find(item => matches(item, selector)) || null; }
  querySelectorAll(selector) { return descendants(this).filter(item => matches(item, selector)); }
}
function descendants(node) { return node.children.reduce((all, child) => all.concat(child, descendants(child)), []); }
function matches(item, selector) {
  if (selector === ".tf-app") return item.className.split(" ").includes("tf-app");
  if (selector === "input") return item.tagName === "INPUT";
  if (selector === "input, select") return item.tagName === "INPUT" || item.tagName === "SELECT";
  if (selector === "button") return item.tagName === "BUTTON";
  if (selector === "input[name], select[name]") return (item.tagName === "INPUT" || item.tagName === "SELECT") && !!item.name;
  if (selector === "button, input, select, summary, [tabindex]") return ["BUTTON", "INPUT", "SELECT", "SUMMARY"].includes(item.tagName) || item.getAttribute("tabindex") !== undefined;
  if (selector === "[data-mapping-focus]") return item.getAttribute("data-mapping-focus") !== undefined;
  if (selector === "[data-mapping-mutation]") return item.getAttribute("data-mapping-mutation") !== undefined;
  if (selector.startsWith(".")) return item.className.split(" ").includes(selector.slice(1));
  const named = /^\[name="([^"]+)"\]$/.exec(selector); if (named) return item.name === named[1];
  return false;
}
function all(root) { return [root].concat(descendants(root)); }
function text(node) { return (node.textContent || "") + " " + node.children.map(text).join(" "); }
const body = new Element("body"), app = new Element("div"), mainMenu = new Element("nav"), host = new Element("section"); app.className = "tf-app"; mainMenu.id = "main-menu"; host.id = "content"; app.appendChild(mainMenu); app.appendChild(host); body.appendChild(app);
document = {body, activeElement: null, createElement(tag) { return new Element(tag); }, createTextNode(value) { const node = new Element("span"); node.textContent = value; return node; }, getElementById(id) { return all(body).find(item => item.id === id) || null; }, querySelector(selector) { return matches(app, selector) ? app : app.querySelector(selector); }, querySelectorAll(selector) { return all(body).filter(item => matches(item, selector)); }, contains(item) { return body.contains(item); }, addEventListener() {}};
const sent = [], windowEvents = {}; class CaptureServer { constructor(cmd) { this.cmd = cmd; } request(data) { sent.push({cmd: this.cmd, data: JSON.parse(JSON.stringify(data || {}))}); } }
const history = {state: null, pushes: [], replacements: [], pushState(state, _unused, hash) { this.state = state; this.pushes.push(hash); }, replaceState(state, _unused, hash) { this.state = state; this.replacements.push(hash); }};
class ShowContent { constructor(index) { this.index = index; } show() {} }
const context = {console: {log() {}, warn() {}}, document, window: {history, location: {hash: ""}, addEventListener(name, listener) { windowEvents[name] = listener; }, setTimeout(listener) { listener(); }}, Server: CaptureServer, HTMLInputElement: Element, HTMLSelectElement: Element, ShowContent, menuItems: Array.from({length: 8}, (_, index) => ({value: ["Playlist", "XMLTV", "Filter", "Mapping", "Users", "Settings", "Log", "Logout"][index], initializeTableHeader() {}})), renderOverview() {}, renderConnections() {}, renderActivity() {}, enableGroupSelection() {}, SERVER: {settings: {"authentication.web": true, files: {m3u: {M1: {name: "Playlist One"}}, xmltv: {X1: {name: "Guide One"}}}}, xepg: {xmltvMap: {"X1.xml": {"guide-1": {"display-name": "Guide"}}}, epgMapping: {
  valid: {"x-channelID": "1", "x-active": true, "x-name": "Valid", "_file.m3u.id": "M1", "x-group-title": "General", "x-xmltv-file": "X1.xml", "x-mapping": "guide-1", "tvg-name": "Valid", url: "http://127.0.0.1/stream"},
  missing: {"x-channelID": "2", "x-active": true, "x-name": "Missing", "_file.m3u.id": "M1", "x-group-title": "General", "x-xmltv-file": "-", "x-mapping": "-", "tvg-name": "Missing"},
  hidden: {"x-channelID": "3", "x-active": true, "x-name": "Hidden", "x-hide-channel": true, "_file.m3u.id": "M1", "x-group-title": "General", "x-xmltv-file": "X1.xml", "x-mapping": "guide-1", "tvg-name": "Hidden"}
}}}, currentDestination: "mapping"};
vm.createContext(context); vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), context); vm.runInContext(fs.readFileSync(process.argv[3], "utf8"), context); vm.runInContext(fs.readFileSync(process.argv[4], "utf8"), context);
function key(target, value, shiftKey) { const event = {key: value, shiftKey: !!shiftKey, prevented: false, preventDefault() { this.prevented = true; }, stopPropagation() {}}; if (target.listeners.keydown) target.listeners.keydown(event); return event; }
context.SERVER.settings.epgSource = "PMS"; context.renderMappingPage(host); const pmsText = text(host); const pmsSettings = all(host).find(item => item.tagName === "BUTTON" && item.textContent === "Open EPG Source settings"); if (pmsSettings) pmsSettings.listeners.click(); const pmsSettingsHash = history.pushes[history.pushes.length - 1] || ""; context.currentDestination = "mapping"; context.SERVER.settings.epgSource = "XEPG";
context.renderMappingPage(host); const pageText = text(host); const selectAll = all(host).find(item => item.getAttribute("aria-label") === "Select all visible channels"); const live = all(host).find(item => item.className === "tf-mapping-counts"); const backupLists = all(host).filter(item => item.id === "mapping-backup-options").length;
let sortButton = all(host).find(item => item.tagName === "BUTTON" && item.textContent === "Channel"), initialSort = sortButton.parentElement.getAttribute("aria-sort"); sortButton.listeners.click(); sortButton = all(host).find(item => item.tagName === "BUTTON" && item.textContent === "Channel"); const toggledSort = sortButton.parentElement.getAttribute("aria-sort");
let typing = all(host).find(item => item.getAttribute("aria-label") === "Search mapping"); typing.focus(); typing.value = "h"; typing.setSelectionRange(1, 1); typing.listeners.input(); typing = document.activeElement; if (document.contains(typing) && typing.listeners.input) { typing.value += "i"; typing.setSelectionRange(2, 2); typing.listeners.input(); } const typedSearch = context.mappingCurrentQuery.search, searchFocused = document.contains(document.activeElement) && document.activeElement.getAttribute("data-mapping-focus") === "filter-search", searchCaret = document.activeElement.selectionStart;
context.mappingCurrentQuery.search = ""; context.mappingCurrentQuery.segment = "active"; context.mappingWorkspaceState.segment = "active"; context.renderMappingPage(host); let playlistFilter = all(host).find(item => item.getAttribute("data-mapping-focus") === "filter-playlist"); playlistFilter.focus(); playlistFilter.value = "M1"; playlistFilter.listeners.change(); const filterFocused = document.contains(document.activeElement) && document.activeElement.getAttribute("data-mapping-focus") === "filter-playlist"; let visibleSelector = all(host).find(item => item.getAttribute("data-mapping-focus") === "select-all-visible"); visibleSelector.focus(); visibleSelector.checked = true; visibleSelector.listeners.change(); const selectAllFocused = document.contains(document.activeElement) && document.activeElement.getAttribute("data-mapping-focus") === "select-all-visible"; context.mappingWorkspaceState.selected.clear(); context.renderMappingPage(host); let rowControl = all(host).find(item => item.getAttribute("data-mapping-focus") === "row-valid"); rowControl.focus(); rowControl.checked = true; rowControl.listeners.click({shiftKey: false}); const firstKeyboardSelected = context.mappingWorkspaceState.selected.has("valid"); rowControl = document.activeElement; if (document.contains(rowControl) && rowControl.listeners.click) { rowControl.checked = false; rowControl.listeners.click({shiftKey: false}); } const keyboardSelected = firstKeyboardSelected && !context.mappingWorkspaceState.selected.has("valid"), keyboardFocused = document.contains(document.activeElement) && document.activeElement.getAttribute("data-mapping-focus") === "row-valid";
const invoker = all(host).find(item => item.getAttribute("data-mapping-edit") === "valid"); context.openMappingEditor(["valid"], invoker); let editor = document.getElementById("mapping-editor"), editorText = text(editor), advanced = all(editor).find(item => item.className === "tf-mapping-editor-advanced"), advancedInitially = advanced.open; const initialEditorFocus = document.activeElement.name;
let fieldSearch = all(editor).find(item => item.getAttribute("aria-label") === "Search editor fields"), apply = all(editor).find(item => item.id === "mapping-apply"); apply.focus(); const tabEvent = key(editor, "Tab", false), tabTrapped = tabEvent.prevented && document.activeElement === fieldSearch; const shiftTabEvent = key(editor, "Tab", true), shiftTabTrapped = shiftTabEvent.prevented && document.activeElement === apply;
let name = all(editor).find(item => item.name === "x-name"); name.value = "Unapplied"; name.setAttribute("data-changed", "true"); key(editor, "Escape", false); const editorEscapeClosed = !document.getElementById("mapping-editor"), editorInvokerRestored = document.activeElement === invoker, editorEscapeDraft = context.mappingWorkspaceState.draft.valid["x-name"];
context.openMappingEditor(["valid"], invoker); editor = document.getElementById("mapping-editor"); advanced = all(editor).find(item => item.className === "tf-mapping-editor-advanced"); fieldSearch = all(editor).find(item => item.getAttribute("aria-label") === "Search editor fields"); fieldSearch.value = "backup"; fieldSearch.listeners.input(); const advancedSearch = advanced.open;
name = all(editor).find(item => item.name === "x-name"); name.value = "Keyboard renamed"; name.setAttribute("data-changed", "true"); const form = all(editor).find(item => item.tagName === "FORM"); context.applyMappingEditor(["valid"], form); const save = document.getElementById("mapping-save");
const unload = {returnValue: "unset", prevented: 0, preventDefault() { this.prevented++; }}; windowEvents.beforeunload(unload);
context.renderNavigation(); const sidebar = all(mainMenu).find(item => item.getAttribute("data-destination") === "settings"); sidebar.focus(); sidebar.listeners.click(); let guard = context.mappingGuardDialog; all(guard).find(item => item.getAttribute("data-mapping-guard") === "stay").listeners.click(); const sidebarReturnFocus = document.activeElement.getAttribute("data-destination") || document.activeElement.id; const stayContinued = 0;
let discardContinued = 0; context.mappingRequestNavigation(() => discardContinued++, undefined); guard = context.mappingGuardDialog; all(guard).find(item => item.getAttribute("data-mapping-guard") === "discard").listeners.click();
context.mappingApplyChannelPatch(context.mappingWorkspaceState, ["valid"], {"x-name": "Saved guard"}); context.renderMappingPage(host); let saveContinued = 0; context.mappingRequestNavigation(() => saveContinued++, undefined); guard = context.mappingGuardDialog; const pendingSave = all(guard).find(item => item.getAttribute("data-mapping-guard") === "save"), pendingDiscard = all(guard).find(item => item.getAttribute("data-mapping-guard") === "discard"), pendingStay = all(guard).find(item => item.getAttribute("data-mapping-guard") === "stay"); pendingSave.listeners.click(); const pendingSaveDisabled = pendingSave.disabled, pendingDiscardDisabled = pendingDiscard.disabled, pendingStayEnabled = !pendingStay.disabled, pendingMutationsDisabled = all(body).filter(item => item.getAttribute("data-mapping-mutation") !== undefined).every(item => item.disabled); context.openMappingEditor(["valid"]); const pendingEditorBlocked = !document.getElementById("mapping-editor"); pendingStay.listeners.click(); const pendingStayClosed = !context.mappingGuardDialog, pendingReturnFocus = document.activeElement.id, pendingSaveAvoided = pendingSaveDisabled && document.activeElement !== document.getElementById("mapping-save"); const request = sent.pop(); const response = JSON.parse(JSON.stringify(context.SERVER)); response.status = true; response.mappingSaveResult = "outputsRebuilt"; response.xepg.epgMapping = request.data.epgMapping; context.completeMappingRequest("saveEpgMapping", request.data, response);
context.mappingApplyChannelPatch(context.mappingWorkspaceState, ["valid"], {"x-name": "Ambiguous guard"}); context.renderMappingPage(host); context.mappingRequestNavigation(() => saveContinued++, undefined); guard = context.mappingGuardDialog; all(guard).find(item => item.getAttribute("data-mapping-guard") === "save").listeners.click(); const ambiguousRequest = sent.pop(); context.completeMappingRequest("saveEpgMapping", ambiguousRequest.data, {status: false, err: "transport"}, "transport"); sent.pop(); context.completeMappingRequest("getServerConfig", {}, {status: false, err: "refetch failed"}); const ambiguousStay = all(guard).find(item => item.getAttribute("data-mapping-guard") === "stay"); ambiguousStay.listeners.click(); const ambiguousStayClosed = !context.mappingGuardDialog, ambiguousReturnFocus = document.activeElement.id;
const removedInvoker = new Element("button"); removedInvoker.id = "removed-invoker"; app.appendChild(removedInvoker); removedInvoker.focus(); context.mappingRequestNavigation(() => saveContinued++, removedInvoker); guard = context.mappingGuardDialog; app.removeChild(removedInvoker); all(guard).find(item => item.getAttribute("data-mapping-guard") === "stay").listeners.click(); const removedReturnFocus = document.activeElement.id;
sidebar.focus(); context.openDestination("settings", true); guard = context.mappingGuardDialog; all(guard).find(item => item.getAttribute("data-mapping-guard") === "stay").listeners.click(); const programmaticReturnFocus = document.activeElement.id;
sidebar.focus(); history.state = {threadfinDestination: "settings"}; context.restoreDestinationFromHistory(); guard = context.mappingGuardDialog; key(guard, "Escape", false); const guardEscapeClosed = !context.mappingGuardDialog, historyReturnFocus = document.activeElement.id, guardEscapeContinued = 0;
process.stdout.write(JSON.stringify({pageText, editorText, selectAllLabel: selectAll.getAttribute("aria-label"), editorRole: editor.getAttribute("role"), editorLabelledBy: editor.getAttribute("aria-labelledby"), advancedInitially, advancedSearch, backupLists, saveFocus: save.focusCount, draftName: "Keyboard renamed", liveStatus: live.getAttribute("aria-live"), unloadPrevented: unload.prevented, unloadReturnValue: unload.returnValue, stayContinued, discardContinued, saveContinued, pendingSaveDisabled, pendingDiscardDisabled, pendingStayEnabled, pendingStayClosed, ambiguousStayClosed, typedSearch, searchFocused, searchCaret, filterFocused, selectAllFocused, keyboardSelected, keyboardFocused, pendingMutationsDisabled, pendingEditorBlocked, initialEditorFocus, tabTrapped, shiftTabTrapped, editorEscapeClosed, editorInvokerRestored, editorEscapeDraft, guardEscapeClosed, guardEscapeContinued, initialSort, toggledSort, sidebarReturnFocus, pendingReturnFocus, pendingSaveAvoided, ambiguousReturnFocus, removedReturnFocus, programmaticReturnFocus, historyReturnFocus, pmsText, pmsSettingsHash}));
`

const mappingLargeRendererNodeScript = `
const fs = require("fs"), vm = require("vm"); let document;
class Element {
  constructor(tag) { this.tagName = String(tag).toUpperCase(); this.children = []; this.parentElement = null; this.attributes = {}; this.className = ""; this.textContent = ""; this.hidden = false; this.value = ""; this.name = ""; this.type = ""; this.checked = false; this.disabled = false; this.open = false; this.listeners = {}; this.selectionStart = 0; this.selectionEnd = 0; }
  set innerHTML(value) { this.children = []; this.textContent = value; } get innerHTML() { return this.textContent; }
  appendChild(child) { child.parentElement = this; this.children.push(child); return child; } removeChild(child) { this.children = this.children.filter(item => item !== child); child.parentElement = null; return child; }
  setAttribute(name, value) { this.attributes[name] = String(value); if (name === "id") this.id = String(value); if (name === "name") this.name = String(value); } getAttribute(name) { return this.attributes[name]; } removeAttribute(name) { delete this.attributes[name]; }
  addEventListener(name, listener) { this.listeners[name] = listener; } focus() { document.activeElement = this; } setSelectionRange(start, end) { this.selectionStart = start; this.selectionEnd = end; }
  contains(target) { return this === target || this.children.some(child => child.contains && child.contains(target)); }
  querySelector(selector) { return walk(this).slice(1).find(item => matches(item, selector)) || null; } querySelectorAll(selector) { return walk(this).slice(1).filter(item => matches(item, selector)); }
}
function walk(root) { const result = [], stack = [root]; while (stack.length) { const item = stack.pop(); result.push(item); for (let i = item.children.length - 1; i >= 0; i--) stack.push(item.children[i]); } return result; }
function matches(item, selector) { if (selector === ".tf-app") return item.className.split(" ").includes("tf-app"); if (selector === "input") return item.tagName === "INPUT"; if (selector === "input, select") return item.tagName === "INPUT" || item.tagName === "SELECT"; if (selector === "input[name], select[name]") return (item.tagName === "INPUT" || item.tagName === "SELECT") && !!item.name; if (selector === "button, input, select, summary, [tabindex]") return ["BUTTON", "INPUT", "SELECT", "SUMMARY"].includes(item.tagName) || item.getAttribute("tabindex") !== undefined; if (selector === "[data-mapping-focus]") return item.getAttribute("data-mapping-focus") !== undefined; if (selector.startsWith(".")) return item.className.split(" ").includes(selector.slice(1)); const named = /^\[name="([^"]+)"\]$/.exec(selector); return named ? item.name === named[1] : false; }
const body = new Element("body"), app = new Element("div"), host = new Element("section"); app.className = "tf-app"; host.id = "content"; app.appendChild(host); body.appendChild(app);
document = {body, activeElement: null, createElement(tag) { return new Element(tag); }, createTextNode(value) { const item = new Element("span"); item.textContent = value; return item; }, getElementById(id) { return walk(body).find(item => item.id === id) || null; }, querySelector(selector) { return matches(app, selector) ? app : app.querySelector(selector); }, querySelectorAll(selector) { return walk(body).filter(item => matches(item, selector)); }, contains(item) { return body.contains(item); }, addEventListener() {}};
const epgMapping = {}; for (let i = 0; i < 5000; i++) epgMapping["id-" + i] = {"x-channelID": String(i + 1), "x-active": true, "x-name": "Channel " + i, "tvg-name": "Backup " + i, "_file.m3u.id": "M1", "x-group-title": "General", "x-xmltv-file": "X1.xml", "x-mapping": "guide"};
const context = {console: {log() {}, warn() {}}, document, window: {addEventListener() {}}, HTMLInputElement: Element, HTMLSelectElement: Element, Server: class { request() {} }, SERVER: {settings: {files: {}}, xepg: {epgMapping, xmltvMap: {"X1.xml": {guide: {}}}}}, currentDestination: "mapping"}; vm.createContext(context); vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), context); vm.runInContext(fs.readFileSync(process.argv[3], "utf8"), context);
const started = Date.now(); context.renderMappingPage(host); const elapsedMs = Date.now() - started; const rendered = walk(host), rows = rendered.filter(item => item.getAttribute("data-mapping-id") !== undefined).length, backupLists = rendered.filter(item => item.id === "mapping-backup-options").length; context.openMappingEditor(["id-4999"]); const editor = document.getElementById("mapping-editor"), editorOpen = !!editor, editorUsable = !!editor && !!editor.querySelector('[name="x-name"]') && !!document.getElementById("mapping-apply") && editor.contains(document.activeElement); process.stdout.write(JSON.stringify({rows, backupLists, editorOpen, editorUsable, elapsedMs}));
`

const mappingNetworkNodeScript = `
const fs = require("fs"), vm = require("vm"); const sockets = [];
class FakeWebSocket { constructor(url) { this.url = url; this.readyState = FakeWebSocket.CONNECTING; this.OPEN = FakeWebSocket.OPEN; this.sent = []; sockets.push(this); } send(value) { this.sent.push(value); } open() { this.readyState = FakeWebSocket.OPEN; this.onopen.call(this, {}); } close(code = 1000) { if (this.readyState === FakeWebSocket.CLOSED) return; this.readyState = FakeWebSocket.CLOSED; if (this.onclose) this.onclose.call(this, {code}); } emitError() { this.onerror.call(this, {}); } respond(response) { response.requestId = JSON.parse(this.sent[this.sent.length - 1]).requestId; this.onmessage.call(this, {data: JSON.stringify(response)}); } }
FakeWebSocket.CONNECTING = 0; FakeWebSocket.OPEN = 1; FakeWebSocket.CLOSING = 2; FakeWebSocket.CLOSED = 3;
const context = {console: {log() {}, warn() {}}, document: {cookie: "", addEventListener() {}, getElementById() { return null; }}, window: {location: {protocol: "http:", hostname: "127.0.0.1", port: "34400"}, addEventListener() {}}, location: {reload() {}}, WebSocket: FakeWebSocket, alert() {}, showElement() {}, createLayout() {}, SERVER: {}, UNDO: {}, SERVER_CONNECTION: false, WS_AVAILABLE: false, setTimeout(callback, delay) { return {callback, delay}; }, clearTimeout() {}};
vm.createContext(context); vm.runInContext(fs.readFileSync(process.argv[2], "utf8"), context); vm.runInContext(fs.readFileSync(process.argv[3], "utf8"), context); vm.runInContext(fs.readFileSync(process.argv[4], "utf8"), context);
function server(name) { return {xepg: {xmltvMap: {"X1.xml": {guide: {}}}, epgMapping: {a: {"x-channelID": "1", "x-active": true, "x-name": name, "x-xmltv-file": "X1.xml", "x-mapping": "guide"}}}}; }
context.initializeMappingWorkspace(server("Before"), true); context.mappingApplyChannelPatch(context.mappingWorkspaceState, ["a"], {"x-name": "Queued"}); vm.runInContext('new Server("saveEpgMapping").request({epgMapping: mappingDeepClone(mappingWorkspaceState.draft)})', context); vm.runInContext('new Server("getServerConfig").request({})', context); sockets[0].open(); sockets[0].respond({status: false, err: "not saved"}); const queuedDirty = context.mappingDirtyIDs(context.mappingWorkspaceState).length, queuedState = context.mappingWorkspaceState.feedback; sockets[0].respond({status: true, xepg: {epgMapping: server("Before").xepg.epgMapping, xmltvMap: server("Before").xepg.xmltvMap}});
context.initializeMappingWorkspace(server("Before"), true); context.mappingApplyChannelPatch(context.mappingWorkspaceState, ["a"], {"x-name": "Maybe"}); vm.runInContext('new Server("saveEpgMapping").request({epgMapping: mappingDeepClone(mappingWorkspaceState.draft)})', context); sockets[0].emitError(); const transportDirty = context.mappingDirtyIDs(context.mappingWorkspaceState).length, transportState = context.mappingWorkspaceState.saveState; sockets[1].open(); const refetchCommand = JSON.parse(sockets[1].sent[0]).cmd;
process.stdout.write(JSON.stringify({queuedDirty, queuedState, transportDirty, transportState, refetchCommand, refetchPending: context.THREADFIN_CONNECTION.active && context.THREADFIN_CONNECTION.active.command === "getServerConfig"}));
`
